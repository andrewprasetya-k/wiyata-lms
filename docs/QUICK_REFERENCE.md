# EDUVERSE LMS - QUICK REFERENCE GUIDE

> **Historical snapshot warning:** This quick reference is an older analysis aid.
> Some routes, context behavior, feature status, and environment notes may be
> stale. Do not use it as the source of truth for active school/role context,
> current routes, AI material summary, chat, email, storage, student notes, or
> implementation status. Verify against current code,
> `docs/PROJECT_CONTEXT_HANDOFF.md`, `backend/docs/API_SUMMARY.md`, and focused
> docs in `backend/docs/api/`.

## Architecture Overview
```
Request Flow:
  HTTP → Middleware (Auth, RBAC, School context)
       → Handler (parse DTO, extract userID)
       → Service (business logic, notifications)
       → Repository (DB query via GORM)
       → PostgreSQL
```

## Key Entities Relationships
```
School (tenant root)
  ├─ User + SchoolUser + UserRole = membership with RBAC
  ├─ AcademicYear → Term → Class → SubjectClass (class + subject + teacher)
  ├─ Material (lives in SubjectClass, NOT Class)
  ├─ Assignment (lives in SubjectClass, NOT Class)
  │   ├─ Submission (1 per student per assignment, upsert)
  │   └─ Assessment (1 per submission, upsert)
  ├─ Enrollment (class membership: student/teacher)
  ├─ Feed (CLASS level, not SubjectClass level)
  ├─ Comment (polymorphic: material, assignment, feed, submission, comment)
  └─ Notification (best-effort activity signals)
```

## Critical Design Decisions
| Concept | Rule | Impact |
|---------|------|--------|
| **Class vs SubjectClass** | Materials/assignments live in SubjectClass, NOT Class | Feed (class) ≠ Material/Assignment (subject class) |
| **Submission** | 1 per student per assignment (upsert) | Resubmit overwrites |
| **Assessment** | 1 per submission (upsert) | Re-grade overwrites |
| **Feed Authorization** | Teacher can post if teaches ≥1 subject in class | Not all teachers can post to class |
| **Notifications** | Best-effort, don't cascade on failure | Main action succeeds even if notif fails |
| **Soft Deletes** | All entities except linking tables | Data recoverable |
| **Comments** | Polymorphic (SourceType + SourceID) | Can comment on material, assignment, feed, submission, comment |
| **Invitation accept (existing user)** | Two endpoints, not one branching on optional auth: `POST /invitations/:token/accept` (public, new-user) vs `POST /invitations/:token/accept-authenticated` (JWT-required, existing-user) | Backend verifies `authenticated user's email == invitation.email`; frontend never trusted for that check |
| **School-role combination (Phase 9.3)** | `admin`+`teacher` is the only allowed combination on one school membership; `student` can never combine with `teacher` or `admin` | One backend validator (`domain.ValidateSchoolRoleCombination`) called from role sync/assign, CSV import/direct-create, and invitation accept — not just the frontend |
| **Audit log (Phase 10.1–10.10)** | DB row (`edv.logs`) is always the source of truth; `/api/ws/audit` broadcasts fire-and-forget only after the row commits | A missed/dropped WebSocket event is invisible to REST — the live feed is convenience, never load-bearing. Read (`LogQueryService`) and write (`LogService`) are separate services on purpose. |

## Routes Quick Map
```
/api/schools              - School CRUD, soft/hard delete
/api/academic-years       - Academic year CRUD + activate/deactivate
/api/terms                - Term CRUD + activate/deactivate
/api/subjects             - Subject CRUD per school
/api/classes              - Class CRUD per term
/api/subject-classes      - Assign teachers to subject classes
/api/enrollments          - Enroll students/teachers to classes
/api/users                - User CRUD
/api/school-users         - User → School membership
/api/rbac/roles           - Role CRUD
/api/rbac/user-roles      - Role assignment
/api/materials            - Material CRUD + progress tracking
/api/assignments          - Assignment/submission/assessment CRUD
/api/feeds                - Feed (class-level) CRUD
/api/comments             - Comment (polymorphic) CRUD
/api/grades               - Grade calculation, weight config
/api/medias               - Media upload, metadata, retrieval
/api/notifications        - Notification CRUD + read status
/api/logs                 - Audit logs: filtered/paginated REST (super-admin platform-wide + school-pinned variants) + `/api/ws/audit` live feed — see backend/docs/api/log.md
/api/dashboard            - Dashboard aggregates per role (student/teacher/admin/super-admin); Phase 7 added work-queue widgets, grading backlog, performance rollup, and growth trends — see backend/docs/api/dashboard.md
/api/invitations          - Public invitation accept (new-user + Phase 8 authenticated existing-user path) — see backend/docs/api/invitation.md
/api/admin/school-member-invitations - Admin invitation create/list/revoke; `fullName` optional since Phase 8 — see backend/docs/api/school_member_invitations.md
/api/login                - Auth login
/api/register             - Auth register
```

## Authorization Layers
```
1. Middleware: AuthRequired (JWT) → RequireSchoolMember → RequireRole
2. Handler: ownership check — resource.SchoolID must equal activeSchoolID from context
3. Service: Custom checks (e.g., teacher must teach in class for feed; student must be enrolled)
4. Repository: GORM soft delete filtering
```

Handler ownership check pattern (Go):
```go
schoolID := getXxxSchoolID(c)  // reads school_id from gin context or SchoolId header
resource, err := h.service.GetByID(id)
if resource.SchoolID != schoolID {
    c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: resource does not belong to active school"})
    return
}
```

## Key Service Patterns

### Notification Triggering (Best-Effort)
```go
// Pattern: Notify after main action succeeds
err := mainAction()
if mainAction fails {
  return err
}

// Trigger notifications (don't fail on errors)
_ = notifService.Create(...) // Error ignored
```

### Attachment Lifecycle
```go
// Create: Link after main entity created
entity.ID = newID
for mediaID := range mediaIDs {
  attService.Link(Attachment{SourceID: entity.ID, MediaID: mediaID})
}

// Update: Unlink old, link new
attService.UnlinkBySource(entity.ID)
for mediaID := range newMediaIDs {
  attService.Link(Attachment{SourceID: entity.ID, MediaID: mediaID})
}

// Delete: Attachment orphaned (manual cleanup only)
```

### Transaction Composition
```go
// Single repository: the repository owns the transaction.
func (r *xRepository) DoThing() error {
  return r.db.Transaction(func(tx *gorm.DB) error {
    // ...
    return nil
  })
}

// Multiple repositories: the service owns the transaction and binds
// each repository to it via WithTx(tx).
s.db.Transaction(func(tx *gorm.DB) error {
  if err := s.repoA.WithTx(tx).DoA(); err != nil {
    return err
  }
  if err := s.repoB.WithTx(tx).DoB(); err != nil {
    return err
  }
  return s.repoC.WithTx(tx).DoC()
})
```

### Multipart Form Handling (MaterialHandler)
```go
// Two flows:
1. JSON: {"mediaIds": [...], "medias": [...inline]}  → service.Create()
2. Multipart: form-data with file uploads           → service.Create() with UploadFile[]

// File upload:
- Check size ≤ 10MB
- Upload to storage provider
- Create Media record
- Link Attachment
- Best-effort cleanup on error
```

## Common Queries
```sql
-- Get materials by subject class
SELECT * FROM edv.materials 
WHERE mat_scl_id = ? 
ORDER BY created_at DESC;

-- Get assignments with submissions and assessments
SELECT a.*, s.*, asm.* FROM edv.assignments a
LEFT JOIN edv.submissions s ON s.sbm_asg_id = a.asg_id
LEFT JOIN edv.assessments asm ON asm.asm_sbm_id = s.sbm_id
WHERE a.asg_scl_id = ? AND a.deleted_at IS NULL;

-- Get user's roles in school
SELECT r.rol_name FROM edv.roles r
JOIN edv.user_roles ur ON ur.urol_rol_id = r.rol_id
JOIN edv.school_users su ON su.scu_id = ur.urol_scu_id
WHERE su.scu_usr_id = ? AND su.scu_sch_id = ?;

-- Get enrollments by class
SELECT * FROM edv.enrollments
WHERE enr_cls_id = ? AND deleted_at IS NULL;

-- Get feed by class with pagination
SELECT f.*, c.* FROM edv.feeds f
LEFT JOIN edv.comments c ON c.cmn_source_id = f.fds_id
WHERE f.fds_cls_id = ? AND f.deleted_at IS NULL
ORDER BY f.created_at DESC
LIMIT ? OFFSET ?;
```

## Environment Variables Required
```
DB_DSN              PostgreSQL connection string
JWT_SECRET          Secret key for JWT signing
JWT_EXPIRY          Token expiry duration (overridden in code to 24h)
STORAGE_PROVIDER    supabase | local | s3 (currently stub)
```

## Error Code Summary
```
200 OK              - Success
201 Created         - Resource created
400 Bad Request     - Invalid input, binding errors
401 Unauthorized    - Missing/invalid JWT
403 Forbidden       - Insufficient permissions, not school member
404 Not Found       - Record doesn't exist
500 Internal Error  - DB or system error
```

## Common Side Effects

| Action | Side Effects |
|--------|--------------|
| Material created | Files uploaded, media created, attachments linked, notifications sent |
| Assignment created | Attachments linked, notifications sent to students |
| Feed created | Authorization check (teacher must teach), attachments linked, notifications sent |
| Comment created | Notification sent to content owner (except self) |
| Submission created | Attachments linked, "is late" computed |
| Assessment created | Notification sent to student |
| Material deleted | Attachments orphaned, comments remain |
| SubjectClass unassigned | Materials/assignments remain (orphaned), no student impact (enrollment is at class level) |

## Known Issues & TODOs

### Critical Issues
- ⚠️ gofmt non-compliance

### Deferred Features (Out of Scope)
- ❌ Real file storage (S3/Supabase)
- ❌ Realtime chat WebSocket
- ❌ Student notes UI
- ❌ Email notifications
- ❌ Signed/private download URLs
- ❌ Thumbnail auto-generation

## Testing Quick Commands
```bash
# Build
go build ./...

# Format check (currently many violations)
gofmt -l .

# Run (requires .env with DB_DSN, JWT_SECRET)
go run ./cmd/api

# Run backend unit tests
go test ./...
```

Tests live in `backend/internal/service/` as `*_test.go` files (same package, standard library only, stub structs). Covered: grade weight validation, assignment deadline enforcement, student note access.

## File Structure
```
backend/
├── cmd/api/main.go           - Route setup, dependency injection
├── internal/
│   ├── domain/               - 19 entity models
│   ├── dto/                  - Request/response contracts
│   ├── handler/              - 23 HTTP handlers
│   ├── service/              - 21 business logic services
│   ├── repository/           - 22 GORM repository interfaces
│   ├── middleware/           - Auth, RBAC, school context
│   └── storage/              - File upload providers
├── schema.md                 - Database schema (DBML format)
├── AGENT.md                  - Engineering context
└── PROJECT_CONTEXT.md        - Business context
```

## Frontend Patterns

### Error extraction (`frontend/src/utils/error.ts`)
```ts
import { getApiError } from '@/utils/error'

catch (error) {
  toast.error(getApiError(error))   // extracts .data.error → .data.message → .message → fallback
}
```
HTTP-status-specific handling (403/404 custom messages) stays inline in the catch block.

### Async stale guard
```ts
// Capture identity before await
const roomId = selectedRoom.value.roomId
try {
  const data = await fetchMessages(roomId)
  if (selectedRoom.value?.roomId !== roomId) return  // discard stale response
  messages.value = data
} catch (e) {
  if (selectedRoom.value?.roomId !== roomId) return
  error.value = getApiError(e)
} finally {
  if (selectedRoom.value?.roomId === roomId) isLoading.value = false
}
```
Apply this pattern in any `watch(async...)` or `onMounted(async...)` that loads data keyed by a selection (class, room, subject class, etc.).

---

## REMEMBER:
1. **School Context Required**: Most routes need SchoolId header or schoolCode URL param
2. **JWT Identity**: UserID extracted from JWT, always trusted source
3. **Soft Deletes**: Orphaned data recoverable; manual cleanup for attachments
4. **Best-Effort Notifications**: Don't cascade failures; main action succeeds
5. **Upsert Semantics**: Submissions and assessments—last write wins
6. **Polymorphic Design**: Comments work on multiple source types
7. **Authorization**: Middleware first, service validation second
