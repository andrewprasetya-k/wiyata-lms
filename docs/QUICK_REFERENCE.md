# EDUVERSE LMS - QUICK REFERENCE GUIDE

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
/api/logs                 - Audit logs
/api/dashboard            - Dashboard aggregates
/api/login                - Auth login
/api/register             - Auth register
```

## Authorization Layers
```
1. Handler: middleware.RequireRole("teacher", "admin", ...)
2. Service: Custom checks (e.g., teacher must teach in class for feed)
3. Repository: GORM soft delete filtering
4. Middleware: JWT + school context + role mapping
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
- ⚠️ Route ordering: GET /assignments/status/:id swallowed by GET /assignments/:assignmentId
- ⚠️ No unit tests
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

# Test (none exist yet)
go test ./...
```

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

---

## REMEMBER:
1. **School Context Required**: Most routes need SchoolId header or schoolCode URL param
2. **JWT Identity**: UserID extracted from JWT, always trusted source
3. **Soft Deletes**: Orphaned data recoverable; manual cleanup for attachments
4. **Best-Effort Notifications**: Don't cascade failures; main action succeeds
5. **Upsert Semantics**: Submissions and assessments—last write wins
6. **Polymorphic Design**: Comments work on multiple source types
7. **Authorization**: Middleware first, service validation second
