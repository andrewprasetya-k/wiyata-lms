# 🧠 Project Handoff Context: Wiyata LMS Backend

## 📌 Project Overview

Wiyata is a Learning Management System (LMS) built with **Go (Gin Framework)** and **GORM**. The system uses a multi-school architecture where users can belong to one or more schools.

## 🏗️ Architectural Patterns (Strictly Follow These)

1.  **Layered Architecture**:
    - `domain`: DB models and TableName definitions.
    - `dto`: Request/Response structures (JSON).
    - `repository`: Raw GORM queries.
    - `service`: Business logic, validation, and cross-service coordination.
    - `handler`: HTTP request parsing and DTO mapping.

2.  **ID vs Code Convention**:
    - **Internal (DB)**: Use UUID for all Primary and Foreign Keys.
    - **External (URL)**: Use `schoolCode` or `subjectCode` in API paths for human-readability.
    - **Converter**: Always use `schoolService.ConvertCodeToID(code)` in the Service layer to translate codes to internal IDs.
    - **Route Pattern (Option A)**:
      - Default detail: `/:id` (UUID).
      - List by parent: `/school/:schoolCode`.
      - Detail by code: `/school/:schoolCode/:subjectCode`.

3.  **Data Integrity & Error Handling**:
    - Repositories **MUST** check `RowsAffected == 0` on Update/Delete/Patch and return `gorm.ErrRecordNotFound` if no row was modified.
    - **Standardized Error Masking**: All handlers **MUST** use `HandleError(c, err)` for service/DB errors and `HandleBindingError(c, err)` for JSON binding/validation. This prevents leaking raw database details (like column names or constraint names) to the client.
    - Centralized error logic is located in `internal/handler/error_handler.go`.

4.  **API Standards**:
    - Standardized `SchoolHeaderDTO` (ID, Name, Code, Logo) or `ClassHeaderDTO` used across all modules when returning school/class context in list responses.
    - Use `Preload` in Repositories to ensure related data (like Creator names or School info) is included in responses.
    - All activation/deactivation actions use the `PATCH` method.

5.  **Material & Assignment Integration**:
    - Both `Material` and `Assignment` belong to `SubjectClass` (using `mat_scl_id` and `asg_scl_id`), not directly to `Class`. This ensures they are linked to a specific subject and teacher within a class.
    - **Upsert Logic**: Submissions and Assessments **MUST** use upsert logic (update if existing for the same user/assignment) to prevent duplicate record accumulation.

## 🔐 Security

- Passwords are hashed using **Bcrypt**.
- RBAC system is implemented: `User` -> `SchoolUser` -> `UserRole` -> `Role` (Global Roles).
- Permissions have been removed in favor of Pure RBAC (Role-based access checks).
- JWT middleware is implemented with 24-hour token expiry.

## 📂 Documentation

Full API specs are available in `backend/docs/api/`. Refer to these files before changing any endpoint behavior.

## 🛠️ Tech Stack Info

- **Database**: PostgreSQL (Supabase).
- **Env**: Managed via `.env` (loaded in `main.go`).
- **Build**: Run `go build ./...` to verify all modules.

## 🎯 Recent Implementations (March 2026)

### **Grade Book System** ✅

- **Files**: `internal/domain/assignment.go` (AssessmentWeight), `internal/dto/assignment_dto.go`, `internal/repository/assessment_weight_repo.go`, `internal/service/grade_service.go`, `internal/handler/grade_handler.go`
- **Endpoints**: `/api/grades/*` - Configure weights, calculate final grades, grade reports
- **Features**: Weighted grade calculation, letter grade conversion, class reports
- **Documentation**: `docs/api/grade.md`

### **Notification System** ✅

- **Files**: `internal/domain/notification.go`, `internal/dto/notification_dto.go`, `internal/repository/notification_repo.go`, `internal/service/notification_service.go`, `internal/handler/notification_handler.go`
- **Endpoints**: `/api/notifications/*` - CRUD notifications, unread count, mark as read
- **Database**: Table `edv.notifications` with index on `(ntf_usr_id, is_read, created_at DESC)`
- **Documentation**: `docs/api/notification.md`

### **RBAC Enhancement** ✅

- **Super Admin Clarification**: System admin (not school admin), read access to all schools, no academic operations without school role
- **Documentation**: Updated `docs/api/rbac.md` with detailed access matrix

---

_Generated on: 12-03-2026. Use this as a prompt for future turns._
