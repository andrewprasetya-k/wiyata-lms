# 📜 Audit Log Module Documentation

Base URL: `/api/logs` (REST) · `/api/ws/audit` (WebSocket)

Built across Phase 10.1–10.12. See also `docs/AUDIT_LOGGING.md` for the
narrative architecture/roadmap overview — this document remains the
detailed REST/WebSocket contract and full taxonomy reference. Source of
truth is always the database
(`edv.logs`, extended by `backend/scripts/migrations/0003_extend_logs_for_audit.sql`
— see `backend/schema.md`); the WebSocket feed is a distribution layer only,
never a second source of data.

## 1. Architecture

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

Key invariants:
- Writing an audit row never depends on the broadcast succeeding — `LogService.Record` (the single choke point both `Log` and `LogBatch` funnel through) publishes only *after* `Repository.Create` returns without error, and the broadcaster is nil-safe.
- The REST read surface (`LogQueryService` / `internal/handler/log_handler.go`) is a **separate service** from `LogService`, added in Phase 10.9 specifically so the write path never has to change for read-surface work.
- `LogRepository`'s write method (`Create`) has never changed since Phase 10.4. Only read methods (`GetBySchool`, `GetByUser`, `GetByCorrelationID`, `Search`, `GetByID`) and `WithTx` were added on top of it.

## 2. REST Endpoints

All require `AuthRequired`. Pagination defaults to `page=1`, `limit=20` (max `100`), sorted `created_at DESC` (fixed — no sort override).

### 2.1 `GET /api/logs`
Unrestricted, platform-wide search.

- **Auth:** `RequireSystemSuperAdmin` — super admin only.
- **Query parameters (all optional):** `schoolId`, `scope` (`platform`|`school`), `action`, `entityType`, `severity` (`LOW`|`MEDIUM`|`HIGH`), `actorUserId`, `correlationId`, `search` (ILIKE against `log_action`, `entity_type`, and the actor's name/email), `dateFrom`, `dateTo` (RFC3339 or bare `YYYY-MM-DD`; `dateTo` is treated as end-of-day), `page`, `limit`.
- Use `schoolId` here (instead of the school-pinned routes below) to narrow without losing platform-wide access.

### 2.2 `GET /api/logs/:id`
Unrestricted detail lookup (any row, any scope).

- **Auth:** `RequireSystemSuperAdmin`.
- Includes `metadata` (raw JSON string — parse client-side), `ipAddress`, `userAgent`, none of which the list response returns.

### 2.3 `GET /api/logs/school/:schoolId/search`
Filtered/paginated search pinned to one school.

- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`.
- **Ownership:** handler forces `schoolId = :schoolId` regardless of any `schoolId` query value the client sends.
- Same query parameters as §2.1 minus `schoolId` (ignored) — `scope` is accepted but moot, since a school-pinned query only ever contains `scope=school` rows.

### 2.4 `GET /api/logs/school/:schoolId/entries/:id`
Detail lookup pinned to one school.

- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`.
- **Ownership:** handler verifies the fetched row's `schoolId` equals `:schoolId`; mismatches return `403 Forbidden`.

### 2.5 `GET /api/logs/school/:schoolId` (legacy, unchanged since before Phase 10.9)
Simple paginated list, no filters, no detail endpoint of its own.

- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`.
- Still backed by `LogService.GetBySchool` — kept for backward compatibility; prefer §2.3/§2.4 for anything new.

**List response shape** (`LogListItemDTO`, no metadata):
```json
{
  "data": [
    {
      "logId": "uuid",
      "action": "member.role.synced",
      "entityType": "school_user",
      "entityId": "uuid",
      "scope": "school",
      "severity": "HIGH",
      "schoolId": "uuid",
      "schoolName": "SMA Contoh",
      "schoolCode": "SMA001",
      "actorUserId": "uuid",
      "actorName": "Admin Budi",
      "actorEmail": "budi@example.sch.id",
      "correlationId": null,
      "createdAt": "2026-07-20T09:00:00Z"
    }
  ],
  "totalItems": 500,
  "page": 1,
  "limit": 20,
  "totalPages": 25
}
```

**Detail response** (`LogDetailDTO`) is `LogListItemDTO` plus `metadata` (raw JSON string), `ipAddress`, `userAgent`.

## 3. WebSocket: `GET /api/ws/audit`

Real-time audit feed, added Phase 10.10. Reuses the same `Hub`/`Client` WebSocket engine chat already uses (`internal/realtime/hub.go`, `client.go`) — a **separate `Hub` instance** from chat's, so audit and chat connections never share a room namespace. No new WebSocket framework was introduced.

### 3.1 Authentication
Manual handshake auth (not gin middleware) — the same pattern chat's `/api/ws/chat` and the sidebar's `/api/events/sidebar` already use, since the raw browser WebSocket API can't set custom headers.

```
GET /api/ws/audit?token=<jwt>&channel=<channel>
```

- `token` — JWT, same as `Authorization: Bearer`.
- `channel` — required, one of the channels in §3.2.

### 3.2 Channels
- `platform` — sentinel room (`realtime.AuditPlatformRoom`), not a real school ID.
- `{schoolId}` — a real school's UUID, used as the room key directly (no `school:` prefix in the actual room key — `Hub.clients` partitions by whatever string a client registers under).

### 3.3 Permission (enforced server-side, at connect time — not per-broadcast)
| Channel | School Admin | System Super Admin |
|---|---|---|
| `platform` | ❌ `403` | ✅ |
| `{ownSchoolId}` | ✅ | ✅ |
| `{other school's id}` | ❌ `403` | ✅ |

Checked via `AuthService.GetContext(userID)` (the same service method `SidebarStreamHandler` already uses) — `globalRoles` for super-admin bypass, `memberships[].roles` for the `admin` role check on a specific school. No new RBAC service method was added for this.

Once connected, broadcasts are pure room fanout — there is no per-event permission re-check; access control happens entirely at the handshake.

### 3.4 Payload
```json
{
  "type": "audit_log_created",
  "channel": "uuid-or-platform",
  "payload": {
    "logId": "uuid",
    "action": "member.role.synced",
    "entityType": "school_user",
    "entityId": "uuid",
    "scope": "school",
    "severity": "HIGH",
    "schoolId": "uuid",
    "actorUserId": "uuid",
    "correlationId": null,
    "createdAt": "2026-07-20T09:00:00Z"
  }
}
```
Same shape as `LogListItemDTO`, deliberately **without** `actorName`, `actorEmail`, `schoolName`, `schoolCode`, or `metadata` — resolving those would need extra queries on the hot write path. The frontend fetches the full row over REST when a row is opened.

### 3.5 Reconnect
Frontend (`frontend/src/services/auditLogSocket.ts`) reuses the exact reconnect policy already established by chat (`chatSocket.ts`): backoff steps `1s → 2s → 5s → 10s`, gives up after 5 consecutive failed-before-open attempts, and never reconnects if the socket was closed by the client itself (e.g. component unmount, filter change moving off a school channel).

### 3.6 Bulk import (CSV) behavior
`LogService.LogBatch` (used by `AdminSchoolMemberImportService.Commit`) only broadcasts the **parent** row (`member.imported`), never the per-row `member.created` children — otherwise a single CSV import of hundreds of rows would flood every connected viewer with one event per row. The parent's `correlationId` is enough for a viewer to notice the batch and pull the full set over REST (`GetByCorrelationID`) if they need it.

## 4. Audit Taxonomy

Pattern: `<domain>.<subject>.<verb_past>`, free-form strings (not a Postgres enum — validated only where relevant in application code). Severity is one of `LOW` / `MEDIUM` / `HIGH` / `CRITICAL` (`domain.LogSeverity*` constants — `CRITICAL` added Phase 10.12, used only by `user.deleted`). 63 actions across 19 service files, grouped by domain below, exactly as implemented (verified against every `logService.Log(...)`/`LogBatch(...)` call site as of Phase 10.12):

### Authentication (`internal/service/auth_service.go`, `user_service.go`, `email_verification_service.go`, `password_reset_service.go`)
| Action | Severity | Notes |
|---|---|---|
| `auth.login.success` | LOW | |
| `auth.login.failed` | MEDIUM | `ActorContext.UserID` is intentionally empty — the caller's identity is exactly what's unknown on a failed login. |
| `auth.registered` | MEDIUM | Also triggers `auth.login.success` right after (auto-login-after-register), by design — two real events. |
| `auth.password.changed` | HIGH | Two distinct call sites emit this same action, disambiguated by metadata's `method` field: (1) super-admin-on-behalf-of-another-user reset (`PATCH /users/change-password/:id`, `UserService.ChangePassword`) — metadata has no `method` field, `user_id` is the *target* user, not necessarily the caller; (2) self-service (`PATCH /me/change-password`, `AuthService.ChangePassword`, Phase 11.1) — metadata `{"user_id": ..., "method": "self_service"}`, `user_id`/actor/target are always the same person. |
| `auth.password.change.failed` | MEDIUM | Phase 11.1. Self-service change-password only (`AuthService.ChangePassword`) — the super-admin reset path has no equivalent failure log. Metadata: `user_id`, `reason` (`"invalid_current_password"` or `"rate_limited"` — the latter once the per-user 5-failed-attempt/15-minute lock, keyed `"change_password:" + userID` via a dedicated `middleware.InMemoryRateLimiterStore` instance, is exhausted). |
| `auth.email.verified` | LOW | |
| `auth.verification.resent` | LOW | Phase 10.14. Emitted by `EmailVerificationService.Resend` — actor-initiated request for a new verification email, distinct from the derived-side-effect exclusions elsewhere in the taxonomy. Metadata: `user_id`, `email`. |
| `auth.password.reset.requested` | LOW | Phase 11.2. Emitted by `PasswordResetService.Request` (`POST /forgot-password`), **always** — regardless of whether the email actually resolved to a real account (a genuine miss doesn't distinguish itself in the log any more than it does in the API response, by design). `ActorContext.UserID` is intentionally always empty, even when the email did resolve — this action records that a request happened, not who received the resulting email. Metadata: `email` only. |
| `auth.password.reset.completed` | HIGH | Phase 11.2. Emitted by `PasswordResetService.Reset` (`POST /reset-password/:token`) only after the token is atomically validated and consumed — `ActorContext.UserID` is populated at this point (the token resolves to a specific user by then), same two-stage shape as `auth.email.verified`. Metadata: `user_id`. Severity matches `auth.password.changed`'s HIGH tier — same consequence class (a password changed), different trigger (token-based, not session-based). |
| `member.login` | LOW | Phase 10.11. School-scoped, emitted only when the user has an active school membership at login (the membership used for `DefaultContext`) — one row per login, not one per membership. Metadata: `login_method`, `user_id`, `school_id`. |

### RBAC (`internal/service/rbac_service.go`)
| Action | Severity |
|---|---|
| `member.role.synced` | HIGH |
| `member.role.assigned` | HIGH |
| `member.role.removed` | HIGH |
| `rbac.role.deleted` | HIGH |

`CreateRole`/`UpdateRole` (global role definition CRUD) remain unaudited — flagged as a gap in the Phase 10.12 architecture audit but not in that phase's implementation scope.

### Member Management (`admin_school_member_import_service.go`, `school_member_invitation_service.go`, `invitation_service.go`, `school_user_service.go`)
| Action | Severity |
|---|---|
| `member.created` | MEDIUM |
| `member.removed` | MEDIUM |
| `member.restored` | MEDIUM |
| `member.imported` (parent, CSV; children reuse `member.created`, same severity) | MEDIUM |
| `member.invited` | LOW |
| `member.invitation.revoked` | LOW |
| `member.invitation.accepted` | LOW |
| `member.enrolled` | MEDIUM | Phase 10.12. `SchoolUserService.Enroll` (`POST /school-users/enroll`) — a second, separate "join a school" path distinct from `member.created` (CSV/direct-create). Metadata: `user_id`, `school_id`. No `role` — enrollment carries no role at all; role assignment is a separate, later RBAC step. |
| `member.unenrolled` | MEDIUM | Phase 10.12. `SchoolUserService.Unenroll` (`DELETE /school-users/:userId`). Metadata: `user_id`, `school_id` (from `ActorContext`, since the endpoint itself receives no schoolID). |

### Enrollment (`internal/service/enrollment_service.go`)
| Action | Severity |
|---|---|
| `enrollment.created` | MEDIUM |
| `enrollment.updated` | MEDIUM |
| `enrollment.removed` | MEDIUM |

### Subject Class (`internal/service/subject_class_service.go`)
| Action | Severity |
|---|---|
| `subject_class.assigned` | MEDIUM |
| `subject_class.reassigned` | MEDIUM |
| `subject_class.unassigned` | MEDIUM |

### School (`internal/service/school_service.go`)
| Action | Severity |
|---|---|
| `school.created` | MEDIUM |
| `school.updated` | MEDIUM |
| `school.deleted` | MEDIUM |
| `school.restored` | MEDIUM |
| `school.hard_deleted` | HIGH |

### Academic Year (`internal/service/academic_year_service.go`) — Phase 10.12
| Action | Severity | Notes |
|---|---|---|
| `academic_year.created` | MEDIUM | Metadata: `year_name` |
| `academic_year.updated` | MEDIUM | Metadata: `year_name` |
| `academic_year.deleted` | HIGH | Metadata: `year_name` (fetched before delete) |
| `academic_year.activated` | HIGH | Metadata: `year_name`, `active_before`, `active_after`. No `academic_year.deactivated` — out of scope. |

### Term (`internal/service/term_service.go`) — Phase 10.12
| Action | Severity | Notes |
|---|---|---|
| `term.created` | MEDIUM | Metadata: `term_name`, `academic_year` (id, not resolved name — avoids an extra query) |
| `term.updated` | MEDIUM | same |
| `term.deleted` | HIGH | same, fetched before delete |
| `term.activated` | HIGH | same. No `term.deactivated` — out of scope. |

### Class (`internal/service/class_service.go`) — Phase 10.12
| Action | Severity | Notes |
|---|---|---|
| `class.created` | MEDIUM | Metadata: `class_name`, `created_by` (substituted for a `homeroom_teacher` concept that doesn't exist on `domain.Class`) |
| `class.updated` | MEDIUM | same |
| `class.deleted` | HIGH | same, fetched before delete |

### Subject (`internal/service/subject_service.go`) — Phase 10.12
| Action | Severity | Notes |
|---|---|---|
| `subject.created` | MEDIUM | Metadata: `subject_name`, `subject_code` |
| `subject.updated` | MEDIUM | same |
| `subject.deleted` | HIGH | same, fetched before delete |

### Material (`internal/service/material_service.go`) — Phase 10.12
| Action | Severity | Notes |
|---|---|---|
| `material.created` | MEDIUM | Metadata: `title`, `class_id` (= `SubjectClassID` — Material has no direct Class relation) |
| `material.updated` | MEDIUM | same |
| `material.deleted` | MEDIUM | same, fetched before delete |

### Assignment Lifecycle (`internal/service/assignment_service.go`) — Phase 10.12
Distinct from the pre-existing Assessment actions below (Phase 10.7, unchanged).
| Action | Severity | Notes |
|---|---|---|
| `assignment.created` | MEDIUM | Metadata: `assignment_title`, `class_id` (= `SubjectClassID`), `due_date` |
| `assignment.updated` | MEDIUM | same |
| `assignment.deleted` | HIGH | same, fetched before delete |
| `assignment.submitted` | MEDIUM | Metadata: `assignment_id`, `student_id` |
| `assignment.submission.updated` | MEDIUM | same |
| `assignment.submission.deleted` | HIGH | same |
| `assignment.category.created` | MEDIUM | Metadata: `category_name` |

`assignment.category.updated`/`assignment.category.deleted` are **not implemented** — no `UpdateCategory`/`DeleteCategory` capability exists anywhere in the service/repository/handler layer; adding audit logging for a mutation that doesn't exist would require adding new business functionality, which was out of scope for Phase 10.12.

### Assignment Assessment (`internal/service/assignment_service.go`, Phase 10.7; severity corrected Phase 10.14)
| Action | Severity |
|---|---|
| `assignment.assessed` | MEDIUM |
| `assignment.assessment.updated` | MEDIUM |
| `assignment.assessment.deleted` | HIGH |

`assignment.assessment.deleted` was raised from MEDIUM to HIGH in Phase 10.14 for consistency with its sibling delete actions in the same domain (`assignment.deleted`, `assignment.submission.deleted`), both already HIGH.

### Grade (`internal/service/grade_service.go`)
| Action | Severity |
|---|---|
| `grade.weights.configured` | HIGH |

### Feed & Comment (`internal/service/feed_service.go`, `comment_service.go`) — Phase 10.14
| Action | Severity | Notes |
|---|---|---|
| `feed.deleted` | MEDIUM | Metadata: `class_id`, `deleted_by_role` (`"author"` if the deleter created the feed, `"admin"` otherwise — the only other path `ensureCanMutateFeed` allows). Matches `material.deleted`'s severity precedent for content entities. |
| `comment.deleted` | MEDIUM | Metadata: `source_type`, `source_id`, `deleted_by_role` (`"author"` or `"admin"`, same convention as `feed.deleted`). |

`Create`/`Update` on both Feed and Comment remain unaudited by design (high-volume, low-consequence, comparable to chat messages — see `docs/AUDIT_LOGGING.md` §19). Only the moderation-relevant `Delete` path was closed in Phase 10.14, since it was the one confirmed-reachable case where an admin/teacher can remove another user's content with no trace.

### User CRUD (`internal/service/user_service.go`) — Phase 10.12, scope: platform (System Super Admin only, `/users` routes)
| Action | Severity | Notes |
|---|---|---|
| `user.created` | HIGH | Metadata: `email`. No `role` — `domain.User` has no global role concept (roles are per-school). |
| `user.updated` | HIGH | same |
| `user.deleted` | **CRITICAL** | same, fetched before delete. First and only use of the `CRITICAL` tier. |

Known interaction: `RBACService.CreateSuperAdmin` internally calls `UserService.Create`, so bootstrapping a super admin now emits **both** `platform.super_admin.created` and `user.created` for the same user — a minor duplicate on a rare, super-admin-only bootstrap path, not suppressed (see Known Limitations).

### Platform (`super_admin_bootstrap_service.go`, `rbac_service.go`, `cmd/api/main.go` retention job — Phase 10.17)
| Action | Severity | Notes |
|---|---|---|
| `platform.school.bootstrapped` | HIGH | |
| `platform.super_admin.created` | HIGH | |
| `platform.logs.retention_cleanup` | LOW | Phase 10.17. Emitted by the in-process retention cleanup job (`startAuditLogRetentionJob`, `cmd/api/main.go`) after every batch-delete run, even when `deleted_count` is 0 — so there is a permanent record that a cleanup pass happened, even though the deleted rows themselves are gone. `ActorContext.UserID` is intentionally empty (system-initiated, no human actor — same convention as `auth.login.failed`). `entityType` is `"log"`, `entityId` is `nil` (the action targets a batch of rows, not one entity). Metadata: `deleted_count` (int64), `cutoff_date` (`YYYY-MM-DD`, the retention cutoff used for that run). |

All business-mutation domains identified in the Phase 10.12 architecture audit are now covered except: RBAC role-definition CRUD (`CreateRole`/`UpdateRole`), `assignment.category.updated`/`.deleted` (no underlying capability), and the intentionally-excluded domains (Notification, Attachment, Media upload, StudentNote, Chat messages, Feed/Comment creation — see `docs/AUDIT_LOGGING.md` §19 for the full reasoning).

### Retention Policy (Phase 10.17)

- **Flat retention: 90 days** across all severities, via `AUDIT_LOG_RETENTION_DAYS` (env, read once at startup). Unset or `0` disables the cleanup job entirely — no rows are ever auto-deleted, and there is no other behavior change.
- **Exception: `CRITICAL` (`user.deleted`) is exempt** — retained permanently, never auto-deleted, regardless of the configured retention value. Enforced unconditionally in `startAuditLogRetentionJob`, not derived from config.
- **Mechanism**: an in-process `time.Ticker`-based goroutine (`startAuditLogRetentionJob`, started in `cmd/api/main.go` alongside the existing WebSocket hub goroutines), ticking every 24h. Deliberately does **not** run immediately at process start — the first cleanup fires after the first 24h tick, so a plain backend restart can never itself trigger a bulk delete.
- **Batching**: `LogRepository.DeleteOlderThan` deletes in batches of 10,000 rows (`DELETE ... WHERE log_id IN (SELECT log_id ... LIMIT 10000)`), looped until a batch affects 0 rows — avoids a single long-held lock/transaction against `edv.logs` on a table that's also on the hot write path.
- **Self-auditing**: every cleanup run (including no-op runs) emits `platform.logs.retention_cleanup` — see the Platform table above.
- See `docs/AUDIT_LOGGING.md` §18 and `docs/PERFORMANCE_AUDIT.md` for the fuller design rationale (why flat retention, why CRITICAL is exempt, why an in-process ticker rather than a cron/job-queue dependency).

## 5. Permission Matrix (REST + WebSocket combined)

| | School Admin | System Super Admin |
|---|---|---|
| List/detail, own school | ✅ (`/logs/school/:schoolId/...`) | ✅ |
| List/detail, another school | ❌ `403` | ✅ (`/logs`, `/logs/:id`, or `/logs/school/:otherId/...`) |
| List/detail, platform scope rows | ❌ (never reachable — school-pinned routes only ever contain `scope=school` rows) | ✅ |
| WebSocket `school:{ownSchoolId}` | ✅ | ✅ |
| WebSocket `school:{other}` | ❌ `403` at handshake | ✅ |
| WebSocket `platform` | ❌ `403` at handshake | ✅ |

## 6. Known Limitations

- **WebSocket payload is intentionally partial** — no `actorName`/`actorEmail`/`schoolName`/`schoolCode`/`metadata`. A row that arrived live shows raw IDs for those fields until the viewer reloads or opens the detail drawer (REST). This is a deliberate trade-off to avoid extra queries on the write path, not a bug.
- **Bulk CSV import broadcasts only the parent row live** — child `member.created` rows exist in the database (and are returned by REST/`GetByCorrelationID`) but are not individually pushed over WebSocket.
- **Live updates pause, rather than guess, when `search`/`dateFrom`/`dateTo` filters are active in the viewer** — the WebSocket payload can't be matched against those filters reliably client-side, so the frontend skips prepending/count-updating while they're set, and relies on REST once the user reloads. Other filters (severity, scope, entityType, actorUserId, schoolId, correlationId) do work live.
- **`GetByCorrelationID` (repository) has no REST endpoint of its own** — it's used internally by the bulk-import broadcast/known-limitation flow above, not exposed to the frontend directly.
- **Database remains the only source of truth.** If the WebSocket connection is down, misses an event, or the browser tab was closed, nothing is lost — the next REST list/reload reflects the true state. The socket is additive convenience, never required for correctness.
- **Retention is now implemented** (Phase 10.17 — see the Retention Policy section above); archival and CSV/Excel export remain unimplemented, still explicitly out of scope.
- **`CreateSuperAdmin` double-logs** (Phase 10.12): bootstrapping a super admin emits both `platform.super_admin.created` and `user.created` for the same user, since it calls the now-instrumented `UserService.Create` internally. Rare (bootstrap-only), low-impact, not suppressed.
- **RBAC role-definition CRUD is still unaudited** — `CreateRole`/`UpdateRole` (global role creation/rename) have no audit trail, unlike `rbac.role.deleted`. Identified in the Phase 10.12 architecture audit, not in that phase's implementation scope.
- **`assignment.category.updated`/`.deleted` do not exist** — there is no underlying `UpdateCategory`/`DeleteCategory` capability in the codebase to audit.
- Several domains remain intentionally unaudited by design (not a gap): Notification, Attachment linking, Media upload, Student Notes, Chat messages, Feed/Comment creation — see `docs/AUDIT_LOGGING.md` §19 for the reasoning behind each.
