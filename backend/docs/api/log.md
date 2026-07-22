# 📜 Audit Log Module Documentation

Base URL: `/api/logs` (REST) · `/api/ws/audit` (WebSocket)

Built across Phase 10.1–10.10. Source of truth is always the database
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

Pattern: `<domain>.<subject>.<verb_past>`, free-form strings (not a Postgres enum — validated only where relevant in application code). Severity is one of `LOW` / `MEDIUM` / `HIGH` (`domain.LogSeverity*` constants). Grouped by domain below, exactly as implemented (verified against every `logService.Log(...)`/`LogBatch(...)` call site as of Phase 10.10):

### Authentication (`internal/service/auth_service.go`, `user_service.go`, `email_verification_service.go`)
| Action | Severity | Notes |
|---|---|---|
| `auth.login.success` | LOW | |
| `auth.login.failed` | MEDIUM | `ActorContext.UserID` is intentionally empty — the caller's identity is exactly what's unknown on a failed login. |
| `auth.registered` | MEDIUM | Also triggers `auth.login.success` right after (auto-login-after-register), by design — two real events. |
| `auth.password.changed` | HIGH | Super-admin-only endpoint (`PATCH /users/change-password/:id`); metadata's `user_id` is the *target* user, not necessarily the caller. |
| `auth.email.verified` | LOW | |

### RBAC (`internal/service/rbac_service.go`)
| Action | Severity |
|---|---|
| `member.role.synced` | HIGH |
| `member.role.assigned` | HIGH |
| `member.role.removed` | HIGH |
| `rbac.role.deleted` | HIGH |

### Member Management (`admin_school_member_import_service.go`, `school_member_invitation_service.go`, `invitation_service.go`)
| Action | Severity |
|---|---|
| `member.created` | MEDIUM |
| `member.removed` | MEDIUM |
| `member.restored` | MEDIUM |
| `member.imported` (parent, CSV; children reuse `member.created`, same severity) | MEDIUM |
| `member.invited` | LOW |
| `member.invitation.revoked` | LOW |
| `member.invitation.accepted` | LOW |

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

### Assignment (`internal/service/assignment_service.go`)
| Action | Severity |
|---|---|
| `assignment.assessed` | MEDIUM |
| `assignment.assessment.updated` | MEDIUM |
| `assignment.assessment.deleted` | MEDIUM |

### Grade (`internal/service/grade_service.go`)
| Action | Severity |
|---|---|
| `grade.weights.configured` | HIGH |

### Platform (`super_admin_bootstrap_service.go`, `rbac_service.go`)
| Action | Severity |
|---|---|
| `platform.school.bootstrapped` | HIGH |
| `platform.super_admin.created` | HIGH |

Not yet audited: Academic Years/Terms/Classes, Subjects, Materials, Feed/Comments/Chat, Notifications, Media, Student Notes, Dashboard — explicitly out of scope for Phase 10.x (see Phase 10.1's severity-grouped inventory for why).

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
- Retention/archival/export (CSV/Excel) were explicitly out of scope for every Phase 10.x sub-phase and remain unimplemented.
