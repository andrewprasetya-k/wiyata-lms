# Audit Logging

Consolidated architecture and roadmap reference for the Wiyata LMS audit
log, covering Phase 10.1 through Phase 10.12. This is the narrative
companion to `backend/docs/api/log.md` (the detailed REST/WebSocket
contract and full taxonomy table) — read this document first for the
"why" and "how it fits together," then `log.md` for exact request/response
shapes and every action's precise severity/metadata.

Last verified against codebase: Phase 10.13 (documentation sync).

## 1. Overview

The audit log gives School Admins and the System Super Admin a
trustworthy, queryable record of who did what, to which entity, and
when — for security review, dispute resolution, and day-to-day
operational visibility. It is deliberately **not** a general-purpose
event stream: high-volume, low-consequence actions (chat messages,
notification read-state, media uploads, personal student notes) are
excluded on purpose (see §19).

The database row is always the single source of truth. The REST API and
the WebSocket live feed are both read-only views over it — a missed or
dropped WebSocket event is invisible to REST, never a data-loss risk.

## 2. Architecture

```
Business action succeeds
        ↓
LogService.Log(...) / LogBatch(...)   — builds the row from domain.ActorContext
        ↓
LogRepository.Create(...)             — the write; source of truth
        ↓
events.AuditBroadcaster (interface)   — fire-and-forget, only after commit
        ↓
realtime.AuditHubBroadcaster → realtime.Hub.BroadcastToRoom
        ↓
WebSocket (/api/ws/audit)
        ↓
Audit Viewer (frontend, live feed prepends into the existing REST-backed list)
```

Three layers, each independently replaceable:
- **Write** (`LogService`) — the only thing every business service depends on.
- **Read** (`LogQueryService`) — a deliberately separate service (Phase 10.9) so read-surface work never touches the write path.
- **Realtime** (`events.AuditBroadcaster` → `realtime.Hub`) — additive convenience layered on top of the write path, never required for correctness.

## 3. Write Flow

1. A business service completes its mutation (and, if the mutation spans multiple repositories, its own `db.Transaction(...)` block has already returned successfully).
2. The service calls `s.logService.Log(actor, action, entityType, entityID, severity, metadata)` — **always after** the mutation, **never inside** its transaction.
3. `LogService.Record` (the single choke point `Log` and `LogBatch` both funnel through) builds a `domain.Log` row and calls `LogRepository.Create`.
4. The service discards any error from the log call (`_ = s.logService.Log(...)`) — an audit failure must never fail the business action.
5. After the row commits, `Record` fires the realtime broadcast (§5), also best-effort.

`LogRepository.Create` — the write method — has not changed since Phase 10.4. Only read methods and `WithTx` were added on top of it since.

## 4. Read Flow

`LogQueryService` (`Search`/`GetByID`, Phase 10.9) is a separate service from `LogService`, backing `LogHandler`'s REST endpoints:

- `GET /logs`, `GET /logs/:id` — unrestricted, platform-wide (System Super Admin only).
- `GET /logs/school/:schoolId/search`, `GET /logs/school/:schoolId/entries/:id` — pinned to one school (that school's admin, or super admin).
- `GET /logs/school/:schoolId` — legacy simple list, kept for backward compatibility.

Full parameter/response detail: `backend/docs/api/log.md` §2, `backend/docs/API_SUMMARY.md`.

## 5. Realtime Flow

```
LogService.Record (after row commit)
        ↓
events.AuditBroadcaster (nil-safe interface)
        ↓
realtime.AuditHubBroadcaster
        ↓
realtime.Hub.BroadcastToRoom(room, event)   — room = "platform" or a schoolId
        ↓
GET /api/ws/audit                            — separate Hub instance from chat's
        ↓
Frontend live viewer (prepends into the REST-backed list, "Live" indicator)
```

Reuses the existing chat `Hub`/`Client` WebSocket engine (a second,
independent instance) rather than introducing a new framework. Payload
is intentionally partial — no actor name/email, school name, or metadata
— to avoid extra queries on the hot write path; the frontend fetches the
full row over REST when a live entry is opened. Bulk CSV import
broadcasts only the parent `member.imported` row, never its per-row
`member.created` children, to avoid flooding connected viewers.

## 6. ActorContext

```go
type ActorContext struct {
    UserID       string
    SchoolUserID *string
    SchoolID     *string
    Scope        string // domain.LogScopeSchool | domain.LogScopePlatform
}
```

Built by `internal/handler/audit_context.go`'s `buildActorContext(c, scope)`:
- `UserID` — from the JWT (`middleware.GetUserID(c)`), always present.
- `SchoolID` / `SchoolUserID` — read from gin context keys `school_id` / `school_user_id`, which are **only ever set by `middleware.RequireSchoolMember`**.

**The single most important invariant in this whole system**: a route
that uses standalone `middleware.RequireRole(...)` without a preceding
`RequireSchoolMember` in the same chain will silently produce
`ActorContext.SchoolID = nil` for any audit log written from it — because
`RequireRole` resolves the school ID as a local variable for its own
role check and never persists it to gin context. This exact bug has now
been found and fixed **8 times**:
- Phase 10.11: `AssignRole`, `RemoveRole`, `UpdateUserRoles` (RBAC).
- Phase 10.12: `AcademicYear.Create`, `Term.Create`, `Subject.Create`, `SchoolUser.Enroll`, `SchoolUser.Unenroll`.

Any new school-scoped mutation route must have `RequireSchoolMember`
before `RequireRole` — check this first, every time, before adding a new
audited action.

## 7. Severity

`domain.LogSeverityLow` / `Medium` / `High` / `Critical` — free text, not
a Postgres enum (validated only in application code, by convention).
`CRITICAL` was added in Phase 10.12 for `user.deleted` — the first and
only action at that tier.

| Tier | Meaning | Examples |
|---|---|---|
| LOW | High-frequency, low individual consequence | `auth.login.success`, `member.login`, `auth.email.verified` |
| MEDIUM | Routine CRUD | `material.created`, `enrollment.created`, `subject.updated` |
| HIGH | Permission/credential changes, destructive actions, cross-cutting config | `member.role.synced`, `assignment.deleted`, `auth.password.changed`, `grade.weights.configured` |
| CRITICAL | Irreversible, platform-wide, highest blast radius | `user.deleted` |

## 8. Scope

`domain.LogScopeSchool` (`"school"`) or `domain.LogScopePlatform`
(`"platform"`) — determines which read-surface endpoint can return the
row. `school` for anything happening inside a specific school's context
(the large majority of actions); `platform` only for routes gated by
`RequireSystemSuperAdmin` with no school concept: `/users` CRUD, RBAC
role-definition management, school bootstrap/restore/hard-delete.

One deliberate scope inconsistency, carried over from Phase 10.6 and not
changed since: `school.created`/`updated`/`deleted` are `scope=school`,
but `school.restored`/`hard_deleted` are `scope=platform` (those two
routes require `RequireSystemSuperAdmin` only, no active school
membership) — meaning a School Admin can never see their own school's
restore/hard-delete history via the school-pinned endpoint. Functionally
defensible, but worth knowing.

## 9. Entity Type

A polymorphic `entityType` string paired with `entityId` — no foreign
key (the target table varies by action: `material`, `assignment`,
`school_user`, `academic_year`, `term`, `class`, `subject`,
`assignment_category`, `submission`, `user`, `school`, `role`,
`invitation`, `enrollment`, `subject_class`, `assessment`). Purely
descriptive, used by the frontend viewer to label rows and (potentially,
future) to build entity deep-links.

## 10. Correlation ID

Links a `LogBatch` parent row to its per-row children that were written
in the same batch — currently used only by CSV member import
(`member.imported` parent, `member.created` children, one shared
`correlation_id`). Ordinary `Log` calls always leave it `null`.
`LogRepository.GetByCorrelationID` exists but has no REST endpoint of
its own — used internally by the bulk-import known-limitation flow, not
exposed directly to the frontend.

## 11. Metadata Convention

Diff-only, never a full entity snapshot, never a sensitive field
(passwords, tokens, hashes, raw invitation tokens). Written using values
already available in memory wherever possible — a small number of delete
actions (Material, Assignment, AcademicYear, Term, Class, Subject, User)
fetch the row once immediately before deleting specifically so the log
can carry a human-readable title/name, since GORM soft-delete makes the
row inaccessible via normal queries afterward.

Where the literal field a taxonomy convention calls for isn't available
without an extra query, the nearest available identifier is used instead
and the substitution documented in `log.md` — e.g. Term logs
`academic_year` as the id (not a resolved name), Material/Assignment log
`class_id` as the `SubjectClassID` (there is no direct Class relation),
Class logs `created_by` in place of a nonexistent `homeroom_teacher`
concept, and `user`/`school_user` actions omit `role` entirely (neither
`domain.User` nor the enrollment DTO carries a role at the point those
actions fire).

## 12. Audit Taxonomy

63 actions across 19 service files as of Phase 10.12. Full table with
severity and metadata: `backend/docs/api/log.md` §4. Summary by domain:

| Domain | Actions | Since |
|---|---|---|
| Authentication | `auth.login.success/failed`, `auth.registered`, `auth.password.changed`, `auth.email.verified`, `member.login` | 10.5–10.8, 10.11 |
| RBAC | `member.role.synced/assigned/removed`, `rbac.role.deleted` | 10.5–10.8 |
| Member Management | `member.created/removed/restored/imported`, `member.invited`, `member.invitation.revoked/accepted`, `member.enrolled`, `member.unenrolled` | 10.5–10.8, 10.12 |
| Enrollment | `enrollment.created/updated/removed` | 10.5–10.8 |
| Subject Class | `subject_class.assigned/reassigned/unassigned` | 10.5–10.8 |
| School | `school.created/updated/deleted/restored/hard_deleted` | 10.5–10.8 |
| Academic Year | `academic_year.created/updated/deleted/activated` | 10.12 |
| Term | `term.created/updated/deleted/activated` | 10.12 |
| Class | `class.created/updated/deleted` | 10.12 |
| Subject | `subject.created/updated/deleted` | 10.12 |
| Material | `material.created/updated/deleted` | 10.12 |
| Assignment Lifecycle | `assignment.created/updated/deleted`, `assignment.submitted`, `assignment.submission.updated/deleted`, `assignment.category.created` | 10.12 |
| Assignment Assessment | `assignment.assessed`, `assignment.assessment.updated/deleted` | 10.7 |
| Grade | `grade.weights.configured` | 10.5–10.8 |
| User CRUD | `user.created/updated/deleted` | 10.12 |
| Platform | `platform.school.bootstrapped`, `platform.super_admin.created` | 10.5–10.8 |

Known gaps (not implemented, tracked): RBAC `CreateRole`/`UpdateRole`;
`assignment.category.updated`/`.deleted` (no underlying
Update/Delete-category capability exists in the codebase at all).

## 13. Permission Matrix

| | School Admin | System Super Admin |
|---|---|---|
| List/detail, own school | ✅ (`/logs/school/:schoolId/...`) | ✅ |
| List/detail, another school | ❌ `403` | ✅ |
| List/detail, platform scope rows | ❌ (unreachable via school-pinned routes) | ✅ |
| WebSocket `{ownSchoolId}` | ✅ | ✅ |
| WebSocket `{other school}` | ❌ `403` at handshake | ✅ |
| WebSocket `platform` | ❌ `403` at handshake | ✅ |

## 14. REST API

Full contract: `backend/docs/api/log.md` §2, `backend/docs/API_SUMMARY.md`.
Summary: `GET /logs`, `GET /logs/:id` (super admin, unrestricted);
`GET /logs/school/:schoolId/search`, `GET /logs/school/:schoolId/entries/:id`
(school-pinned); `GET /logs/school/:schoolId` (legacy). All paginated
(`page`, `limit` max 100), sorted `created_at DESC` (fixed).

## 15. WebSocket

`GET /api/ws/audit?token=&channel=` — full contract: `backend/docs/api/log.md`
§3. Manual handshake auth (JWT as query param). `channel` is `platform`
or a school ID. Permission checked once at connect via
`AuthService.GetContext`, not per-broadcast.

## 16. Frontend Viewer

`frontend/src/pages/common/AuditLogsPage.vue` — shared page mounted at
both `/admin/audit-logs` and `/superadmin/audit-logs`. Table with
pagination, debounced search, severity/scope/entityType/actor/date
filters, detail drawer with a JSON metadata viewer, and a live WebSocket
feed (`frontend/src/services/auditLogSocket.ts`, same reconnect policy as
chat's socket) that prepends new rows with a highlight + "Live"
indicator. Live updates pause (rather than guess) while `search`/
`dateFrom`/`dateTo` filters are active, since the WebSocket payload can't
be reliably matched against those filters client-side.

## 17. Cara Menambah Audit Action Baru

See the "Audit Logging Cheat Sheet" in `docs/QUICK_REFERENCE.md` for the
step-by-step developer checklist (middleware chain check, ActorContext
construction, `Log` call placement, scope/severity/entity/metadata
decision table, `LogBatch` usage). Repeated in full there rather than
here to keep one canonical location.

## 18. Retention Policy (Phase 10.17)

- **Flat retention: 90 days** across all severities, via the `AUDIT_LOG_RETENTION_DAYS` env var (read once at startup). Unset or `0` disables the cleanup job entirely — no rows are ever auto-deleted, and there is no other behavior change in that case.
- **Exception: `CRITICAL` (`user.deleted`) is exempt** — retained permanently, never auto-deleted, regardless of the configured value. This exemption is hardcoded in `startAuditLogRetentionJob` (`cmd/api/main.go`), not driven by config, so it can't be accidentally disabled by an env change.
- **Mechanism**: `LogRepository.DeleteOlderThan(cutoff, excludeSeverities)` deletes in batches of 10,000 rows (`DELETE ... WHERE log_id IN (SELECT log_id ... LIMIT 10000)`, looped until a batch affects 0 rows) rather than one giant statement, to avoid a long-held lock/transaction on `edv.logs` — a table that's also on the hot write path for every mutation in the app.
- **Schedule**: an in-process `time.Ticker`-based goroutine (`startAuditLogRetentionJob`), started in `cmd/api/main.go` alongside the existing WebSocket hub goroutines, ticking every 24h. It deliberately does **not** run immediately at process start — the first cleanup fires only after the first 24h tick, so a plain backend restart can never itself trigger a bulk delete. No external job scheduler/cron dependency was introduced (none existed in the codebase before this — confirmed via a `go.mod`/`time.NewTicker`/`time.AfterFunc` grep during the Phase 10.16 design review).
- **Self-auditing**: every cleanup run — including no-op runs that delete 0 rows — emits `platform.logs.retention_cleanup` (LOW severity, `entityType: "log"`, metadata `{deleted_count, cutoff_date}`), so there's a permanent record that a cleanup pass happened even though the deleted rows themselves are gone. The actor is empty (`ActorContext{Scope: platform}`, no `UserID`) since this is a system-initiated event with no human actor — the same convention already used by `auth.login.failed`.
- See `backend/docs/api/log.md` §4 ("Retention Policy" subsection under the Audit Taxonomy) for the taxonomy-level detail, and `docs/PERFORMANCE_AUDIT.md` for the original operational-hardening design discussion this implements.

## 19. Known Limitations

- WebSocket payload is intentionally partial (no actor name/email, school name, or metadata) — a deliberate trade-off, not a bug.
- Bulk CSV import broadcasts only the parent row live; child rows exist in the database and REST but aren't individually pushed.
- Live updates pause when `search`/`dateFrom`/`dateTo` filters are active in the viewer.
- `GetByCorrelationID` has no REST endpoint of its own.
- `RBACService.CreateSuperAdmin` double-logs (`platform.super_admin.created` + `user.created`) since Phase 10.12, because it internally calls the now-instrumented `UserService.Create`. Rare (bootstrap-only), not suppressed.
- RBAC role-definition CRUD (`CreateRole`/`UpdateRole`) remains unaudited.
- `assignment.category.updated`/`.deleted` do not exist — there is no underlying capability to audit.
- `POST /academic-years`, `/terms`, `/subjects` still accept `schoolId`/scoping fields in the request body without validating them against the caller's active school header — a pre-existing, documented, unrelated LOW-risk gap (not part of the audit-logging initiative; see `docs/PROJECT_CONTEXT_HANDOFF.md` §27).
- **Retention is now implemented** (§18); archival and export (CSV/Excel) remain unimplemented — still explicitly out of scope.
- Database is always the only source of truth; the WebSocket layer can be down, drop events, or reconnect without any data loss — the next REST reload always reflects the true state.

## 20. Future Roadmap

Ordered by the priority established in the Phase 10.12 architecture audit:

1. Close the RBAC role-definition CRUD gap (`CreateRole`/`UpdateRole`).
2. Taxonomy consistency cleanup: the systemic 2-segment-creation vs 3-segment-followup naming split (`member.invited` vs `member.invitation.*`; `assignment.assessed` vs `assignment.assessment.*`); unify the role/permission action prefix currently split across `rbac.*`/`member.role.*`/`platform.*`; make an explicit product decision on the `school.*` scope split (§8).
3. Read-surface enhancements (frontend-only, no backend risk): actor picker (replace the raw UUID text box), entity deep-link, diff viewer for actions that already carry paired before/after metadata, saved filters.
4. WebSocket severity gate (stop broadcasting LOW-severity events like every login, which currently costs a broadcast attempt for zero realized viewer value), final production sign-off review. (Retention policy, the other Phase 10.16 operational-hardening item, is done — §18.)

Intentionally out of scope, by design, not by omission: auditing
`NotificationService`, `AttachmentService`, `MediaService` upload paths,
`StudentNoteService`, `ChatService` message creation, and
`FeedService`/`CommentService` creation. These are either high-volume
with low individual consequence (chat messages, media uploads, feed
posts — auditing them would be noise, not oversight), derived side
effects of an already-audited action (a notification created because an
assignment was graded would double-log the same event under a different
name), or privacy-sensitive personal content (student notes — auditing
personal study material would be surveillance, not institutional
oversight). `FeedService`/`CommentService` *deletion* (specifically the
moderation case — an admin/teacher removing someone else's post) and
chat room lifecycle (create/rename/add/remove member, as distinct from
message content) are noted as possible future nice-to-haves, not
scheduled.
