Engineering Context Report

1. What This App Does
   Wiyata LMS is a backend API for a multi-school learning management system. It supports schools, users, school memberships, RBAC roles, academic years, terms, subjects, classes,
   enrollments, learning materials, feeds, comments, assignments, submissions, grading, notifications, logs, and dashboards. See backend/PROJECT_CONTEXT.md:3 and backend/docs/
   API_SUMMARY.md:1.
2. Main Tech Stack And Runtime
   Backend-only Go app using Gin, GORM, PostgreSQL/Supabase, JWT, bcrypt, and dotenv. The server starts from backend/cmd/api/main.go:16, loads DB_DSN from .env, connects with GORM
   Postgres using PreferSimpleProtocol, and listens on :8080.
3. Architecture And Folder Structure
   The app follows a strict layered architecture documented in backend/PROJECT_CONTEXT.md:7:
   domain for GORM models/table names, dto for request/response contracts, repository for GORM queries, service for business logic, and handler for HTTP parsing/mapping. Dependency
   wiring is manual in backend/cmd/api/main.go:33.
4. Key Domains/Modules
   Core modules: school, academic year, term, user, school user, subject, RBAC, class, subject class, enrollment, media, attachment, material, feed, comment, assignment, grade,
   notification, log, dashboard. API groups are registered in backend/cmd/api/main.go:133.
5. Data Flow And Integrations
   Request flow is Gin route -> middleware -> handler -> service -> repository -> PostgreSQL. Public auth routes are /api/login and /api/register; everything after that uses
   AuthRequired() JWT middleware. School-scoped routes rely on SchoolId header or schoolCode URL param, then RBAC checks roles through repository queries. Database schema is
   documented in backend/schema.md:40. File/media support exists, but real S3/Supabase storage integration is still marked as TODO.
6. Local Setup Commands
   From backend:

go mod download
cp .env.example .env # not present currently; create .env manually if needed
go run ./cmd/api

Required env keys detected from .env and code: DB_DSN, JWT_SECRET; JWT_EXPIRY exists in .env but auth code currently hardcodes 24 hours in backend/internal/service/
auth_service.go:35.

7. Test, Lint, And Build Commands
   Verified:

go test ./...
go build ./...

Both pass. Unit tests exist under backend/internal/service/ (student_note_service_test.go, grade_service_test.go, assignment_service_test.go, and others). Formatting check:

gofmt -l .

This reports many files as not gofmt-formatted, including backend/cmd/api/main.go:1, handlers, DTOs, services, repositories, and domains.

8. Coding Conventions Detected
   Use UUID internally and human-readable schoolCode/subjectCode externally; convert codes to IDs in service layer. Use centralized HandleError and HandleBindingError from backend/
   internal/handler/error_handler.go:13. Repositories should check RowsAffected == 0 on writes. Soft deletes are common. Response DTOs use app-specific JSON names like schoolId,
   schoolName, etc.
   When a service must orchestrate multiple repositories inside one atomic transaction, the service owns the `db.Transaction(...)` block and binds each repository to it via
   `repository.WithTx(tx)`. Repositories remain the only place for data access — services should not issue raw GORM queries when an equivalent repository method already exists.
9. Potential Pitfalls
   Some docs say assignments belong to SubjectClass, but backend/schema.md:265 still shows asg_cls_id, while project context says asg_scl_id. Shell startup has an unrelated
   issue: /Users/andrewprasetya/.zprofile:13: unmatched " appears on every command. backend/tmp/main is a built binary artifact in the repo tree.
   Note: GET /assignments/status/:id route ordering was previously a pitfall but has since been secured with RequireSchoolMember + RequireRole middleware.
10. Authorization Pattern
    The enforced authorization chain for mutation endpoints is:
    AuthRequired → RequireSchoolMember (sets school_id, school_user_id in context) → RequireRole → handler ownership check.
    Handler ownership check: after fetching the resource by ID, verify resource.SchoolID == activeSchoolID (or equivalent field). Return 403 if mismatch. This prevents cross-school
    mutations even when a user holds admin role in multiple schools.
11. Assessment Weight Transaction
    ConfigureWeights uses AssessmentWeightRepository.ReplaceBySubject(), which runs delete + insert in a single GORM transaction. Never use separate DeleteBySubject + Create calls
    for weight replacement — partial failures would leave weights in an inconsistent state.
12. Backend Test Strategy
    Tests live in backend/internal/service/ as *_test.go files, same package (package service). Use stub structs implementing the repository interface — no mock framework, no testify.
    Only test business rules in the service layer. Do not test handler JSON binding, simple CRUD repos, or DTO shapes. Sentinel errors use errors.Is(); non-sentinel errors use
    strings.Contains(err.Error(), "...").
13. Recommended Next Steps
    Run a gofmt-only cleanup. Add focused tests around auth/RBAC context middleware and invitation accept flows. Remove or ignore generated tmp artifacts. Implement the high-priority
    TODOs in backend/TODO.md in this order: real file storage, notification triggers, assignment extension flow.
