# EDUVERSE LMS - COMPLETE CODEBASE ANALYSIS REPORT

> **Historical snapshot warning:** This document is an older analysis snapshot and
> several sections may be stale. Do not use it as the source of truth for active
> school/role context, route registration, AI material summary, chat, email,
> storage, student notes, or current implementation status. Verify current
> behavior against code, `docs/PROJECT_CONTEXT_HANDOFF.md`,
> `backend/docs/API_SUMMARY.md`, and focused docs in `backend/docs/api/`.

## EXECUTIVE SUMMARY

Wiyata LMS is a multi-tenant learning management system built with Go (Gin), PostgreSQL, and JWT auth. The application implements a strict 4-layer architecture: Handler → Service → Repository → Domain/Database. The system manages schools, users, academic structures (academic years, terms, classes, subject classes), learning materials, assignments with submissions/grading, communication (feeds, comments), and notifications.

---

## 1. ROUTES INVENTORY

### 1.1 Authentication Routes (Public)

```
POST   /api/login                           - Login user
POST   /api/register                        - Register new user
```

### 1.2 School Management Routes (Protected)

```
POST   /api/schools                         - Self-service Create School (any authenticated user with verified email — no longer super_admin-only; see docs/PROJECT_CONTEXT_HANDOFF.md §13)
GET    /api/schools                         - List schools (super_admin)
GET    /api/schools/summary                 - School summary (super_admin)
GET    /api/schools/check-code/:schoolCode  - Check code availability
GET    /api/schools/:schoolCode             - Get school details (school member)
PATCH  /api/schools/:schoolCode             - Update school (admin/super_admin, school member)
PATCH  /api/schools/restore/:schoolCode     - Restore deleted school (super_admin)
DELETE /api/schools/:schoolCode             - Soft delete school (admin/super_admin, school member)
DELETE /api/schools/permanent/:schoolCode   - Hard delete school (super_admin)
```

### 1.3 Academic Structure Routes (Protected)

```
POST   /api/academic-years                  - Create academic year (admin/super_admin)
GET    /api/academic-years                  - List all academic years
GET    /api/academic-years/:id              - Get academic year by ID
GET    /api/academic-years/school/:schoolCode - Get by school (school member)
PATCH  /api/academic-years/:id              - Update (admin/super_admin)
PATCH  /api/academic-years/activate/:id     - Activate (admin/super_admin)
PATCH  /api/academic-years/deactivate/:id   - Deactivate (admin/super_admin)
DELETE /api/academic-years/:id              - Delete (admin/super_admin)

POST   /api/terms                           - Create term (admin/super_admin)
GET    /api/terms                           - List all terms
GET    /api/terms/:id                       - Get term by ID
GET    /api/terms/academic-year/:academicYearId - Get by academic year
PATCH  /api/terms/:id                       - Update (admin/super_admin)
PATCH  /api/terms/activate/:id              - Activate (admin/super_admin)
PATCH  /api/terms/deactivate/:id            - Deactivate (admin/super_admin)
DELETE /api/terms/:id                       - Delete (admin/super_admin)

POST   /api/subjects                        - Create subject (admin/super_admin)
GET    /api/subjects                        - List all subjects
GET    /api/subjects/:id                    - Get subject by ID
GET    /api/subjects/school/:schoolCode     - Get by school (school member)
GET    /api/subjects/school/:schoolCode/:subjectCode - Get by code (school member)
PATCH  /api/subjects/:id                    - Update (admin/super_admin)
DELETE /api/subjects/:id                    - Delete (admin/super_admin)

POST   /api/classes                         - Create class (admin/teacher, school member)
GET    /api/classes                         - List all classes
GET    /api/classes/:id                     - Get class by ID
PATCH  /api/classes/:id                     - Update (admin/teacher, school member)
DELETE /api/classes/:id                     - Delete (admin, school member)
```

### 1.4 Subject Class Routes (Protected)

```
POST   /api/subject-classes/assign          - Assign teacher to subject class (admin/teacher)
GET    /api/subject-classes/my-teaching     - Get my teaching (teacher, school member)
GET    /api/subject-classes/class/:classId  - Get by class (school member)
GET    /api/subject-classes/:id             - Get by ID (school member)
PATCH  /api/subject-classes/:id             - Update (admin/teacher)
DELETE /api/subject-classes/:id             - Unassign (admin)
```

### 1.5 Enrollment Routes (Protected)

```
POST   /api/enrollments                     - Enroll member (admin/teacher, school member)
GET    /api/enrollments/class/:classId      - Get by class
GET    /api/enrollments/member/:schoolUserId - Get by member
GET    /api/enrollments/:id                 - Get by ID
PATCH  /api/enrollments/:id                 - Update (admin/teacher, school member)
DELETE /api/enrollments/:id                 - Unenroll (admin/teacher, school member)
```

### 1.6 User Management Routes (Protected)

```
POST   /api/users                           - Create user (admin/super_admin)
GET    /api/users                           - List all users (admin/super_admin)
GET    /api/users/:id                       - Get user by ID
PATCH  /api/users/:id                       - Update user
PATCH  /api/users/change-password/:id       - Change password
DELETE /api/users/:id                       - Delete user (admin/super_admin)

POST   /api/school-users/enroll             - Enroll user to school (admin/super_admin)
GET    /api/school-users/school/:schoolCode - Get members by school (school member)
GET    /api/school-users/user/:userId       - Get schools by user
DELETE /api/school-users/:userId            - Unenroll from school (admin/super_admin)
```

### 1.7 RBAC Routes (Protected)

```
POST   /api/rbac/roles                      - Create role (super_admin)
GET    /api/rbac/roles                      - Get all roles
GET    /api/rbac/roles/:id                  - Get role by ID
PATCH  /api/rbac/roles/:id                  - Update role (super_admin)
DELETE /api/rbac/roles/:id                  - Delete role (super_admin)

POST   /api/rbac/user-roles                 - Assign role (admin/super_admin)
DELETE /api/rbac/user-roles                 - Remove role (admin/super_admin)
GET    /api/rbac/user-roles/:schoolUserId   - Get user roles
PATCH  /api/rbac/user-roles/:schoolUserId   - Update user roles (admin/super_admin)

POST   /api/rbac/super-admin                - Create super admin (super_admin)
```

### 1.8 Material Routes (Protected)

```
POST   /api/materials                       - Create material (teacher, school member)
GET    /api/materials                       - List materials
GET    /api/materials/:id                   - Get material by ID
PATCH  /api/materials/:id                   - Update (teacher, school member)
DELETE /api/materials/:id                   - Delete (teacher/admin, school member)
POST   /api/materials/progress              - Update progress
```

### 1.9 Assignment Routes (Protected)

```
POST   /api/assignments/categories          - Create category (admin, school member)
GET    /api/assignments/categories/school/:schoolCode - Get categories by school (school member)

POST   /api/assignments                     - Create assignment (teacher, school member)
GET    /api/assignments/subject-class/:subjectClassId - Get by subject class
GET    /api/assignments/status/:id          - Get assignment status
GET    /api/assignments/my-submission/:assignmentId - Get my submission (student, school member)
GET    /api/assignments/:assignmentId       - Get submissions by assignment
PATCH  /api/assignments/:id                 - Update (teacher, school member)
DELETE /api/assignments/:id                 - Delete (teacher/admin, school member)

POST   /api/assignments/submit/:assignmentId - Submit (student, school member)
GET    /api/assignments/submit/:submissionId - Get submission by ID
PATCH  /api/assignments/submit/:submissionId - Update submission (student, school member)
DELETE /api/assignments/submit/:submissionId - Delete submission (student, school member)

POST   /api/assignments/assess/:submissionId - Assess (teacher, school member)
PATCH  /api/assignments/assess/:submissionId - Update assessment (teacher, school member)
DELETE /api/assignments/assess/:submissionId - Delete assessment (teacher, school member)
```

### 1.10 Feed & Comment Routes (Protected)

```
POST   /api/feeds                           - Create feed (teacher/admin, school member)
GET    /api/feeds/class/:classId            - Get by class
GET    /api/feeds/:id                       - Get by ID
PATCH  /api/feeds/:id                       - Update
DELETE /api/feeds/:id                       - Delete

POST   /api/comments                        - Create comment
GET    /api/comments                        - Get by source
GET    /api/comments/:id                    - Get by ID
PATCH  /api/comments/:id                    - Update
DELETE /api/comments/:id                    - Delete
```

### 1.11 Grade Routes (Protected)

```
POST   /api/grades/weights                  - Configure weights (admin/teacher)
GET    /api/grades/weights/subject/:subjectId - Get weights by subject (school member)
GET    /api/grades/class/:classId/subject/:subjectId - Get class report (teacher/admin)
```

### 1.12 Media & Attachment Routes (Protected)

```
POST   /api/medias/upload                   - Upload media
POST   /api/medias/metadata                 - Record metadata
GET    /api/medias/:id                      - Get media by ID
DELETE /api/medias/:id                      - Delete media
```

### 1.13 Notification Routes (Protected)

```
GET    /api/notifications                   - Get notifications
GET    /api/notifications/unread-count      - Get unread count
PATCH  /api/notifications/read/:id          - Mark as read
PATCH  /api/notifications/read-all          - Mark all as read
DELETE /api/notifications/:id               - Delete notification
```

### 1.14 Log & Dashboard Routes (Protected)

```
GET    /api/logs/school/:schoolId           - Get logs by school
GET    /api/dashboard/student/:userId       - Student dashboard
GET    /api/dashboard/teacher/:schoolUserId - Teacher dashboard
GET    /api/dashboard/admin/:schoolId       - Admin dashboard (Phase 7: work-queue widgets, grading backlog — see backend/docs/api/dashboard.md)
GET    /api/dashboard/super-admin           - Super admin dashboard (Phase 7: work-queue widgets, growth trend charts — see backend/docs/api/dashboard.md)
```

### 1.15 Invitation Routes

```
GET    /api/invitations/:token                       - Get invitation metadata (public; includes `existingUser` since Phase 8)
POST   /api/invitations/:token/accept                 - Accept as new user (public; name/password registration, unchanged)
POST   /api/invitations/:token/accept-authenticated    - Accept as existing user (Phase 8; requires JWT, verifies email matches invitation server-side)

POST   /api/admin/school-member-invitations           - Create invitation (admin; `fullName` optional since Phase 8, see backend/docs/api/school_member_invitations.md)
GET    /api/admin/school-member-invitations           - List invitations (admin)
PATCH  /api/admin/school-member-invitations/:id/revoke - Revoke invitation (admin)
```

Full contracts: `backend/docs/api/invitation.md` (public accept flow) and `backend/docs/api/school_member_invitations.md` (admin create/list/revoke).

---

## 2. HANDLER LAYER

### 2.1 Structure

- **Pattern**: HTTP request parsing → DTO binding → service invocation → response mapping
- **Error Handling**: Centralized `HandleError()` and `HandleBindingError()` from `error_handler.go`
- **Context**: Extracts `userID` via `middleware.GetUserID(c)` from JWT claims
- **Response Format**: JSON with `gin.H{}` maps or typed DTOs

### 2.2 Key Handlers (23 total)

1. **AuthHandler**: Login, Register
2. **SchoolHandler**: CRUD schools, check availability, soft/hard delete
3. **AcademicYearHandler**: CRUD academic years with activate/deactivate
4. **TermHandler**: CRUD terms with activate/deactivate
5. **UserHandler**: CRUD users, change password
6. **SchoolUserHandler**: Enroll/unenroll users to schools
7. **SubjectHandler**: CRUD subjects per school
8. **RBACHandler**: Roles, user role assignments, super admin creation
9. **ClassHandler**: CRUD classes
10. **SubjectClassHandler**: Assign teachers to subject classes
11. **EnrollmentHandler**: Enroll/unenroll students/teachers to classes
12. **MaterialHandler**: CRUD materials with multipart file upload support
13. **AssignmentHandler**: CRUD assignments/submissions/assessments/categories
14. **FeedHandler**: CRUD feeds with class-level communication
15. **CommentHandler**: CRUD comments on polymorphic sources (material, assignment, feed, submission, comment)
16. **GradeHandler**: Configure weights, calculate grades
17. **MediaHandler**: Upload, metadata recording, retrieval, deletion
18. **NotificationHandler**: Fetch, mark read, delete notifications
19. **LogHandler**: Fetch logs by school
20. **DashboardHandler**: Student/teacher dashboard data

### 2.3 Handler Patterns Observed

- **Multipart Form Support**: MaterialHandler supports both JSON and `multipart/form-data`
- **Pagination**: Query params `page` (default 1), `limit` (default varies)
- **School Context**: Via `SchoolId` header or `schoolCode` URL param (enforced by middleware)
- **User Identity**: From JWT middleware context (`middleware.GetUserID(c)`)
- **Role-based Authorization**: Delegated to middleware `RequireRole()`
- **DTO Mapping**: Request DTOs for input validation, Response DTOs for output

---

## 3. SERVICE LAYER

### 3.1 Architecture

- **21 Service Interfaces** with focused responsibilities
- **Dependency Injection**: Services depend on repositories, other services, and storage providers
- **Best-Effort Notifications**: Notification failures don't cascade (errors ignored with `_`)
- **Transaction Safety**: Soft deletes via GORM, RowsAffected checks on updates

### 3.2 Service Categories

#### Academic Management Services

- **AcademicYearService**: Create, list, activate/deactivate academic years per school
- **TermService**: CRUD terms within academic years, toggle active status
- **SubjectService**: CRUD subjects per school
- **ClassService**: CRUD classes within terms, school-scoped

#### User & Authorization Services

- **UserService**: CRUD users, password management, activation status
- **SchoolUserService**: Manage user memberships in schools
- **AuthService**: Login/register, JWT token generation
- **RBACService**: Role assignment, permission checking, super admin management

#### Learning Content Services

- **SubjectClassService**: Assign teachers to subject-class combinations
- **MaterialService**:
  - Create (supports JSON + file uploads via multipart)
  - File storage integration (upload to provider, record metadata)
  - Progress tracking (mark completed, track last_opened_at)
  - Best-effort notification to enrolled students
- **EnrollmentService**: Enroll students/teachers to classes with role differentiation

#### Assignment & Assessment Services

- **AssignmentService**:
  - Create/update/delete assignments with categories
  - Submit (upsert) submissions
  - Assess (upsert) assessments with score/feedback
  - Track "is late" via deadline comparison
  - Notify students when assignments created
- **GradeService**: Calculate weighted grades per student per subject

#### Communication & Content Services

- **FeedService**:
  - Create feed at CLASS level (not subject)
  - Authorization: teachers must teach ≥1 subject in class
  - Notify all class members (except creator) on post
- **CommentService**:
  - Polymorphic comments (source_type/source_id)
  - Notify content owner on comment (except self-comment)
- **AttachmentService**: Link media to sources (material, assignment, feed, submission, comment)

#### System Services

- **NotificationService**: Create/fetch/mark-read notifications
- **MediaService**: Upload files, record metadata, delete with storage cleanup
- **LogService**: Log user actions with metadata
- **DashboardService**: Aggregate dashboard data per role (student/teacher/admin/super-admin). Phase 7 extended admin/super-admin with work-queue widgets, grading backlog, and platform growth trends — see `docs/PROJECT_CONTEXT_HANDOFF.md` §25 and `backend/docs/api/dashboard.md`

### 3.3 Key Patterns

1. **Upload File Handling**:

   ```go
   // MaterialService.Create()
   - Upload file to storage provider
   - Create Media record with storage path
   - Link Attachment to material
   - Best-effort cleanup on error
   ```

2. **Notification Triggering**:
   - **Assignment Created**: Notify all students in class
   - **Material Added**: Notify all students in class
   - **Feed Posted**: Notify all class members except creator
   - **Comment Added**: Notify content owner (except self-comment)
   - All are best-effort (errors swallowed)

3. **Authorization Checks**:
   - Feed creation: Teacher must teach ≥1 subject in class
   - School scoping: Class must belong to school

4. **Attachment Management**:
   - Manual link/unlink (no cascade deletes)
   - Unlinking on update (delete old, create new)

---

## 4. REPOSITORY LAYER

### 4.1 Structure

- **22 Repository Interfaces** (one per domain entity/concept)
- **Implementation**: Direct GORM queries against PostgreSQL
- **Soft Deletes**: Handled via GORM's `DeletedAt` column
- **Validation**: RowsAffected checks on updates/deletes (throws `gorm.ErrRecordNotFound` if 0)
- **Preloading**: Aggressive preloading to avoid N+1 queries

### 4.2 Repository Types

#### Core Entity Repositories

1. **SchoolRepository**: CRUD schools, code→ID conversion, soft/hard delete
2. **UserRepository**: CRUD users, find by email, activate/deactivate
3. **SchoolUserRepository**: Get school members, get user's schools
4. **AcademicYearRepository**: CRUD per school, activate/deactivate
5. **TermRepository**: CRUD per academic year, activate/deactivate
6. **SubjectRepository**: CRUD per school
7. **ClassRepository**: CRUD per term, get by school, school ID lookup
8. **SubjectClassRepository**: Assign teacher to subject-class, get by class, teacher teaching checks
9. **EnrollmentRepository**: Get members by class/school user, count students

#### Content Repositories

10. **MaterialRepository**: CRUD, get by subject class with pagination
11. **AssignmentRepository**: CRUD assignments, submissions, assessments
12. **AttachmentRepository**: Link/unlink media to content sources
13. **MediaRepository**: CRUD media files
14. **FeedRepository**: CRUD feeds by class with pagination
15. **CommentRepository**: CRUD comments, get by source (polymorphic), count by source

#### System Repositories

16. **NotificationRepository**: CRUD notifications, get unread count
17. **LogRepository**: Create logs, get by school
18. **GradeRepository**: Calculate student grades
19. **AssessmentWeightRepository**: Set/get weights by subject-category
20. **RBACRepository**: Role checks (is super admin, in school, has role)
21. **DashboardRepository**: Aggregate dashboard metrics per role, including Phase 7 additions (classes/subject-classes/subjects needing attention, grading backlog, school performance rollup, super-admin work queue, growth trends) — see `backend/docs/api/dashboard.md`
22. **ContentOwnerRepository**: Polymorphic owner lookup (for comment notifications)

### 4.3 Key Query Patterns

#### Preloading Strategy

```go
// Example: GetAssignmentWithSubmissions
db.Preload("Category").
   Preload("SubjectClass.Subject").
   Preload("SubjectClass.Class").
   Preload("Submissions.User").
   Preload("Submissions.Assessment.Assessor")
```

#### Pagination

```go
offset := (page - 1) * limit
query.Limit(limit).Offset(offset).Order("created_at desc")
```

#### Soft Deletes

```go
// GORM handles deleted_at automatically
// Queries ignore soft-deleted records unless .Unscoped()
Delete(&model) // Sets deleted_at, doesn't remove
```

#### RowsAffected Safety

```go
result := db.Save(model)
if result.RowsAffected == 0 {
    return gorm.ErrRecordNotFound // Throws error if nothing updated
}
```

---

## 5. DOMAIN MODELS (19 entities)

### 5.1 Entity Dependency Graph

```
School (root tenant)
├── User (global identity)
├── SchoolUser (user ∈ school)
│   └── UserRole (role assignment)
│   └── Enrollment (class membership)
├── AcademicYear → Term → Class
│   ├── Subject
│   └── SubjectClass (class + subject + teacher)
│       ├── Material
│       │   └── MaterialProgress (user progress per material)
│       └── Assignment
│           ├── Submission (student submission)
│           │   └── Assessment (grading)
│           └── AssignmentCategory
├── Feed (class-level communication)
├── Comment (polymorphic: on material, assignment, feed, submission, comment)
├── Media/Attachment (file storage)
├── Notification (activity signals)
└── Log (audit trail)
```

### 5.2 Critical Domain Models

#### School

```go
School {
    ID, Name, Code (unique), Address, Email, Phone, Website
    LogoID (foreign to Media)
    Timestamps: created_at, updated_at, deleted_at (soft delete)
    Schema: edv.schools
}
```

#### User & SchoolUser

```go
User {
    ID, FullName, Email (unique per deleted_at), Password (hidden in JSON)
    IsActive, Timestamps, DeletedAt
    Schema: edv.users
}

SchoolUser {
    ID, UserID (foreign User), SchoolID (foreign School)
    Roles []UserRole (one user can have multiple roles per school)
    Schema: edv.school_users
    Unique: (scu_usr_id, scu_sch_id)
}
```

#### Academic Structure

```go
AcademicYear { ID, SchoolID, Name, IsActive }
Term { ID, AcademicYearID, Name, IsActive }
Subject { ID, SchoolID, Name, Code (unique per school) }
Class { ID, SchoolID, TermID, Code, Title, Description, CreatedBy, IsActive, Timestamps, DeletedAt }
Enrollment { ID, SchoolID, SchoolUserID, ClassID, Role (teacher|student), JoinedAt }
SubjectClass { ID, ClassID, SubjectID, SchoolUserID (teacher) }
```

#### Learning Content

```go
Material {
    ID, SchoolID, SubjectClassID (not ClassID—critical design)
    Title, Description, Type (video|pdf|ppt|other)
    CreatedBy, Timestamps, DeletedAt, Attachments
    Schema: edv.materials
}

MaterialProgress {
    ID, UserID, MaterialID, Status (not_started|completed)
    LastOpenedAt (nullable—passive tracking)
    Unique: (map_usr_id, map_mat_id)
}

Assignment {
    ID, SchoolID, SubjectClassID, CategoryID
    Title, Description, Deadline (nullable)
    AllowLateSubmission, CreatedBy, Timestamps, DeletedAt
    Submissions, Attachments
    Schema: edv.assignments
}

Submission {
    ID, SchoolID, AssignmentID, UserID, SubmittedAt, DeletedAt
    Assessment (nested), Attachments
    Unique: (sbm_asg_id, sbm_usr_id)—1 submission per student per assignment
}

Assessment {
    ID, SubmissionID, Score (decimal), Feedback, AssessedBy, AssessedAt
    Schema: edv.assessments
}

AssignmentCategory { ID, SchoolID, Name }
AssessmentWeight { ID, SubjectID, CategoryID, Weight (decimal) }
```

#### Communication

```go
Feed {
    ID, SchoolID, ClassID (NOT SubjectClassID—class-level only)
    Content, CreatedBy, Timestamps, DeletedAt
    Attachments, Comments (loaded manually)
}

Comment {
    ID, SchoolID, SourceType (material|assignment|feed|submission|comment)
    SourceID, UserID, Content, Timestamps, DeletedAt
    Schema: edv.comments
    Polymorphic: SourceType + SourceID identify target
}

Attachment {
    ID, SchoolID, SourceType (material|assignment|feed|submission|comment)
    SourceID, MediaID, CreatedAt
    Schema: edv.attachments
}

Media {
    ID, SchoolID, Name, FileSize, MimeType
    StoragePath (local path), FileURL (public URL), ThumbnailURL
    IsPublic, OwnerType (user|material|assignment|feed|submission|comment|school|system)
    OwnerID, CreatedAt, DeletedAt
}
```

#### System

```go
Notification {
    ID, UserID, Type (assignment_created|assignment_graded|comment_added|material_added|feed_posted)
    Title, Message, Link (optional), RelatedID (UUID to content)
    IsRead, CreatedAt
    Schema: edv.notifications
}

Log {
    ID, SchoolID, UserID, Action, Metadata (jsonb), CreatedAt
    Schema: edv.logs
}

Role { ID, Name }
UserRole { ID, SchoolUserID, RoleID, CreatedAt }
```

### 5.3 SourceType Constants (Polymorphic)

```go
const (
    SourceMaterial   SourceType = "material"
    SourceAssignment SourceType = "assignment"
    SourceFeed       SourceType = "feed"
    SourceSubmission SourceType = "submission"
    SourceComment    SourceType = "comment"
)
```

### 5.4 Status Constants

```go
MaterialProgress:
    - not_started
    - completed

Assignment:
    - AllowLateSubmission: bool (not enum)
    - Deadline: nullable

Notification Types (constants):
    - assignment_created
    - assignment_graded
    - comment_added
    - material_added
    - feed_posted
```

---

## 6. DATA TRANSFER OBJECTS (DTOs)

### 6.1 DTO Categories

#### Request DTOs (Input Validation)

- **Binding Tags**: `binding:"required,uuid"`, `binding:"omitempty"`, `binding:"oneof=teacher student"`
- **Schema**: Match required fields for creation, optional fields for updates

#### Response DTOs (Output Mapping)

- **Pattern**: Flatten/transform domain models for API consumption
- **Naming**: Consistent snake_case → camelCase conversion

### 6.2 Key DTO Flows

#### Material Flow

```go
CreateMaterialDTO {
    SchoolID (uuid), SubjectClassID (uuid), Title, Description, Type, MediaIDs, Medias
}

CreateMediaInline {
    Name, FileSize, MimeType, FileURL, ThumbnailURL (for video embeds, YouTube, etc.)
}

MaterialResponseDTO {
    ID, Title, Description, Type, Creator, CreatedAt, Attachments
}
```

#### Assignment Flow

```go
CreateAssignmentDTO {
    SchoolID, SubjectClassID, CategoryID, Title, Description, Deadline, AllowLateSubmission, MediaIDs
}

CreateSubmissionDTO {
    SchoolID, MediaIDs
}

SubmissionResponseDTO {
    ID, UserName, SubmittedAt, IsLate, Attachments, Assessment
}

CreateAssessmentDTO {
    Score (required), Feedback
}

AssessmentResponseDTO {
    Score, Feedback, Assessor, AssessedAt
}
```

#### Feed Flow

```go
CreateFeedDTO {
    SchoolID, ClassID, Content, MediaIDs
}

FeedResponseDTO {
    ID, Content, CreatorName, CreatedAt, Attachments, CommentCount
}
```

#### Enrollment Flow

```go
CreateEnrollmentDTO {
    SchoolID, SchoolUserIDs (array), ClassID, Role (teacher|student)
}

EnrollmentResponseDTO {
    ID, SchoolID, SchoolUserID, UserFullName, UserEmail, ClassID, ClassTitle, Role, JoinedAt
}
```

---

## 7. VALIDATION RULES

### 7.1 Handler-Level Validation (via DTO binding)

#### Required Fields

```go
// School context
SchoolID:        required, uuid
SchoolCode:      required, string (URL param)

// Academic structure
AcademicYearID:  required, uuid
TermID:          required, uuid
SubjectID:       required, uuid
ClassID:         required, uuid
SubjectClassID:  required, uuid

// Learning content
Title:           required, string
CategoryID:      required, uuid
UserID:          required, uuid

// Roles & permissions
Role:            required, oneof=teacher student (class enrollment)
                 oneof=admin student teacher super_admin (platform roles)
```

#### Optional Fields

```go
// Nullable in domain
Deadline:        nullable time.Time
AllowLateSubmission: default true (bool)
Description:     optional, string
Feedback:        optional, string
```

#### Special Validations

- **UUID Format**: Binding tag `uuid` validates format
- **Email Unique**: User email unique per deleted_at state (enforced at DB level)
- **School Code Unique**: Global unique constraint
- **Subject Code Unique**: Per school
- **Class Code Unique**: Per school + term
- **Submission Unique**: Per assignment + user (upsert semantics)
- **Enrollment Unique**: Per school user + class
- **SubjectClass Unique**: Per class + subject + teacher
- **MaterialProgress Unique**: Per user + material
- **Attachment Linkage**: SourceType + SourceID valid for source entity

### 7.2 Service-Level Validation

#### Authorization Checks

```go
// Feed creation (teacher)
- Teacher must teach ≥1 subject in class

// Material/Assignment/Feed creation
- User must be school member (checked via middleware first)

// Submission
- Student must be enrolled in class
- 1 submission per student per assignment (upsert)

// Assessment
- Assessor (teacher) must teach the class
```

#### Business Rule Validation

```go
// Feed posting
- Class must belong to school

// Late submission tracking
- IsLate = Deadline != nil && SubmittedAt > Deadline

// Grade calculation
- Weights must sum correctly per subject
```

---

## 8. DATABASE TABLES & RELATIONSHIPS

### 8.1 Core Tables

```
edv.schools (root tenant)
├── edv.users
├── edv.school_users
│   └── edv.user_roles
├── edv.academic_years
│   └── edv.terms
│       └── edv.classes
├── edv.subjects
├── edv.subject_classes (class + subject + teacher)
├── edv.enrollments (class membership)
├── edv.materials (live in subject_classes)
├── edv.material_progress (per student per material)
├── edv.assignments (live in subject_classes)
├── edv.submissions (student work)
├── edv.assessments (grading)
├── edv.feeds (class communication)
├── edv.comments (polymorphic)
├── edv.medias (file storage metadata)
├── edv.attachments (media links)
├── edv.notifications
└── edv.logs
```

### 8.2 Key Relationships

#### School Scoping

```
All entities (except User, Role) have scu_sch_id or equivalent
Enforces data isolation between schools
```

#### Class vs SubjectClass

```
Class: Academic context (XII IPA 1)
    - 1 class : N subject classes
    - Materials/assignments live in subject classes (NOT class)
    - Feed live in class (communication across all subjects)
    - Enrollments link students/teachers to class

SubjectClass: Daily learning workspace
    - Composite: ClassID + SubjectID + TeacherID
    - 1 teacher : 1 subject : 1 class
    - Materials, assignments, grades scoped to subject class
```

#### Soft Deletes

```
Most entities: deleted_at nullable timestamptz
GORM automatically filters out soft-deleted in queries
Hard delete available for super_admin only
```

#### Composite Unique Keys

```
(scu_usr_id, scu_sch_id)          → SchoolUser
(acy_sch_id, acy_name)             → AcademicYear
(trm_acy_id, trm_name)             → Term
(sub_sch_id, sub_code)             → Subject
(cls_sch_id, cls_code, cls_trm_id) → Class
(scl_cls_id, scl_sub_id, scl_scu_id) → SubjectClass
(urol_scu_id, urol_rol_id)         → UserRole
(sbm_asg_id, sbm_usr_id)           → Submission
(map_usr_id, map_mat_id)           → MaterialProgress
(asw_sub_id, asw_asc_id)           → AssessmentWeight
(snt_usr_id, snt_mat_id)           → StudentNote (future)
```

---

## 9. EXISTING BUSINESS RULES

### 9.1 School & Membership

1. **Super Admin Bypass**: Super admins bypass school membership checks
2. **User→SchoolUser→Role Chain**: Roles attach to school membership, not global user
3. **One User Multi-School**: User can belong to multiple schools with different roles
4. **School Soft Delete**: Deleted schools recoverable by super admin

### 9.2 Academic Structure

1. **Active State**: Academic years and terms have `is_active` boolean (not enforced at DB level—advisory)
2. **Term Ownership**: Terms belong to academic years; classes belong to terms
3. **Class-Term Lock**: Class code unique per school+term combo (can't duplicate class in same term)

### 9.3 Learning Hierarchy

1. **Material Ownership**: Materials belong to SubjectClass, NOT Class
2. **Assignment Ownership**: Assignments belong to SubjectClass, NOT Class
3. **Feed Scope**: Feeds belong to Class (cross-subject communication)
4. **Student Enrollment**: Enrollment at CLASS level; extends to all subject classes in class

### 9.4 Assignment Lifecycle

1. **Submission Upsert**: Only 1 submission per student per assignment (upsert on resubmit)
2. **Late Tracking**: IsLate computed via Deadline comparison (not stored)
3. **Assessment Upsert**: Only 1 assessment per submission (upsert on re-grade)
4. **Deadline Optional**: Assignments can have no deadline (nullable)
5. **Late Allowed**: AllowLateSubmission boolean allows submission after deadline

### 9.5 Communication

1. **Feed Authorization**: Teachers can only post if teaching ≥1 subject in class
2. **Comment Polymorphism**: Comments can attach to material, assignment, feed, submission, or another comment
3. **Content Owner Notification**: Content owner notified of comment (except self-comment)

### 9.6 Notifications (Best-Effort)

1. **Assignment Created**: Notify all students in class
2. **Material Added**: Notify all students in class
3. **Feed Posted**: Notify all class members except creator
4. **Comment Added**: Notify content owner (except self-comment)
5. **Assignment Graded**: Notify student
6. **Failure Tolerance**: Notification failures don't cascade; main action succeeds

### 9.7 Material Progress

1. **No Auto-Complete**: Opening material doesn't mark completed
2. **Last Opened Tracking**: LastOpenedAt updated passively (advisory only)
3. **Explicit Completion**: Status='completed' requires explicit action (via UpdateProgress endpoint)

### 9.8 Grade Calculation

1. **Weighted by Category**: Grades weighted per subject + assignment category
2. **Weights Per Subject**: Different subjects can have different category weights
3. **No GPA**: Only per-subject grades tracked

---

## 10. POTENTIAL SIDE EFFECTS & DEPENDENCIES

### 10.1 Cascading Operations

#### Material Creation

```
Material created
  ├─ Files uploaded to storage provider
  ├─ Media records created
  ├─ Attachments linked
  └─ Notifications sent to all class students (best-effort)

If any file upload fails:
  └─ Best-effort cleanup of uploaded objects (may orphan files)
```

#### Assignment Creation

```
Assignment created
  ├─ Attachments linked
  └─ Notifications sent to all class students (best-effort)
```

#### Feed Creation

```
Feed created
  ├─ Authorization check: teacher must teach in class
  ├─ Attachments linked
  └─ Notifications sent to all class members except creator (best-effort)
```

#### Comment Creation

```
Comment created
  └─ Notification sent to content owner (if not self-comment, best-effort)
```

#### Submission Creation

```
Submission created (upsert)
  ├─ Attachments linked
  └─ No notification (student action)
```

#### Assessment Creation

```
Assessment created (upsert on submission)
  └─ Notification sent to student (best-effort)
```

### 10.2 Deletion Side Effects

#### Material Deletion

```
Material soft-deleted
  ├─ Material progress orphaned (no cascade delete)
  ├─ Attachments orphaned
  └─ Comments remain but reference deleted material
```

#### Assignment Deletion

```
Assignment soft-deleted
  ├─ Submissions remain (soft-deleted separately)
  ├─ Assessments remain
  └─ Attachments orphaned
```

#### Class Deletion

```
Class soft-deleted
  ├─ Enrollments remain (soft-deleted separately)
  ├─ SubjectClasses remain
  ├─ Feeds remain
  ├─ Materials remain (via SubjectClass)
  └─ Comments remain
```

#### SubjectClass Deletion

```
SubjectClass deleted (unassign)
  ├─ Materials remain (orphaned, still accessible)
  ├─ Assignments remain (orphaned)
  └─ No students affected (enrollment at CLASS level)
```

#### School Deletion (Soft)

```
School soft-deleted
  └─ Entire school data hidden from queries (GORM soft delete filter)
```

#### School Deletion (Hard/Permanent)

```
School permanently deleted
  └─ ALL school data removed (foreign key constraints cascade if enabled—verify!)
```

### 10.3 Update Side Effects

#### Material Update

```
Material updated
  ├─ Old attachments unlinked
  ├─ New attachments linked
  └─ No notification (teacher action)
```

#### Assignment Update

```
Assignment updated
  ├─ Old attachments unlinked
  ├─ New attachments linked
  └─ No re-notification (already posted)
```

#### Feed Update

```
Feed updated
  ├─ Old attachments unlinked
  ├─ New attachments linked
  └─ No notification (already posted)
```

#### Submission Update

```
Submission updated
  ├─ Old attachments unlinked
  ├─ New attachments linked
  └─ No notification
```

#### Assessment Update

```
Assessment updated
  └─ Score/feedback updated
  └─ Consider: should student be re-notified? (currently not)
```

### 10.4 Authorization Dependencies

#### Feed Creation (Teacher Role)

```
Handler: middleware.RequireRole("teacher")
Service: Verify teacher teaches ≥1 subject in class
  └─ Query: SubjectClass where ClassID = feed.ClassID AND TeacherID = userID

If not authorized:
  └─ Return 403 "teacher does not teach any subject in this class"
```

#### Material Creation (Teacher Role)

```
Handler: middleware.RequireRole("teacher")
Service: (No service-level auth check—only handler-level)
```

#### Assignment Creation (Teacher Role)

```
Handler: middleware.RequireRole("teacher")
Service: (No service-level auth check—only handler-level)
```

#### Submission (Student Role)

```
Handler: middleware.RequireRole("student")
Service: (No student verification—assume upsert is idempotent and allowed)
```

#### Assessment (Teacher Role)

```
Handler: middleware.RequireRole("teacher")
Service: (No teacher-owns-assignment check—any teacher can assess)
  └─ TODO: Should verify teacher owns the assignment's subject class
```

### 10.5 Storage Provider Integration

#### Upload Flow

```
Handler receives multipart files
  ├─ Check file size ≤ 10MB
  ├─ Call storage.Upload(ctx, objectPath, reader, mimeType)
  ├─ Create Media record with returned publicURL
  ├─ Link Attachment
  └─ On error: best-effort cleanup via storage.Delete()
```

#### Storage Provider Types

```
1. Supabase (default in .env)
2. Local filesystem (disabled if provider nil)
3. Custom S3-compatible
```

#### Current Status

```
- Real upload attempted (not mock)
- Cleanup on failure is best-effort (may orphan files)
- No signed URLs or private downloads (TODO)
- No thumbnail generation (TODO)
```

### 10.6 Data Consistency Risks

#### Race Conditions

1. **Submission Upsert**: Two concurrent submissions from same student → last write wins (expected)
2. **Assessment Upsert**: Two concurrent assessments from different teachers → last write wins (unexpected, should lock)
3. **Material Progress**: Concurrent progress updates → GORM Save() overwrites (expected)

#### Orphaned Data

1. **Soft-deleted materials** with attachments: Attachments not automatically cleaned
2. **Orphaned SubjectClass**: Materials/assignments remain accessible
3. **Uploaded files**: If Media creation fails, objects orphaned in storage

#### Missing Constraints

1. **Teacher Ownership**: Assessment doesn't verify teacher created assignment
2. **Student Enrollment**: Submission doesn't verify student enrolled in class
3. **Submission Count**: No max submission limit (expected—unlimited resubmit)

---

## 11. MIDDLEWARE FLOW

### 11.1 Authentication Middleware (`AuthRequired()`)

```
1. Extract Authorization header ("Bearer <token>")
2. Parse JWT with JWT_SECRET from .env
3. Verify HMAC signature
4. Extract claims (user_id, email, etc.)
5. Store claims in context c.Set("user", claims)
6. Allow next() or Abort with 401
```

### 11.2 School Member Middleware (`RequireSchoolMember()`)

```
1. Extract schoolID from:
   a. SchoolId header (priority 1)
   b. schoolCode URL param (priority 2)
2. If super admin: bypass school membership check
3. Otherwise: verify user is member of school
4. Store schoolID in context c.Set("school_id", schoolID)
5. Allow next() or Abort with 403
```

### 11.3 Role Middleware (`RequireRole()`)

```
1. Get schoolID from context (set by RequireSchoolMember) or from header/param
2. Query RBAC: Get user's roles in school
3. Verify user has one of allowed roles
4. Store roles in context c.Set("user_roles", roles)
5. Allow next() or Abort with 403
```

### 11.4 CORS Middleware

```
corsMiddleware() applies to all routes
```

### 11.5 Middleware Chaining Pattern

```
// Example: Create material (protected, school member, teacher only)
materialAPI.POST("",
    middleware.RequireSchoolMember(schoolService),
    middleware.RequireRole(schoolService, "teacher"),
    materialHandler.Create
)

// Execution order:
// 1. AuthRequired() [from api.Use()]
// 2. RequireSchoolMember()
// 3. RequireRole()
// 4. materialHandler.Create
```

---

## 12. ERROR HANDLING

### 12.1 Centralized Error Handler

```go
func HandleError(c *gin.Context, err error) {
    // Maps Go errors to HTTP responses
    // Examples: gorm.ErrRecordNotFound → 404
}

func HandleBindingError(c *gin.Context, err error) {
    // Maps DTO binding errors to 400
    // Extracts validation error details
}
```

### 12.2 Common Error Responses

```
401 Unauthorized       - Missing/invalid JWT
403 Forbidden          - Insufficient permissions, not school member
404 Not Found          - Record doesn't exist (gorm.ErrRecordNotFound)
400 Bad Request        - Invalid input, missing fields, binding errors
422 Unprocessable Entity - Business logic violation (suggested, not verified in code)
500 Internal Server Error - Unexpected DB/system errors
```

### 12.3 Error Propagation

```
Handler
  ├─ Service
  │   ├─ Repository (DB errors propagate up)
  │   └─ Notification service (errors swallowed with _)
  └─ Handler error mapping to HTTP response
```

---

## 13. REQUEST/RESPONSE FLOW EXAMPLES

### Example 1: Create Material (with file upload)

```
Request:
  POST /api/materials
  Headers: Authorization: Bearer <token>, SchoolId: <uuid>
  Body: multipart/form-data
    - schoolId: UUID
    - subjectClassId: UUID
    - materialTitle: string
    - materialDesc: string (optional)
    - materialType: video|pdf|ppt|other
    - files: [File, File, ...]

Handler (MaterialHandler.Create):
  1. Extract userID from JWT context
  2. Parse multipart form
  3. Validate required fields
  4. Create Material domain object
  5. Call service.Create(mat, mediaIDs, uploads, medias)

Service (MaterialService.Create):
  1. Save material to DB
  2. For each uploaded file:
     a. Upload to storage provider
     b. Create Media record
     c. Append to mediaIDs
  3. Link attachments
  4. Notify students (best-effort)

Repository (MaterialRepository.Create):
  1. INSERT into edv.materials

Response:
  201 Created
  { "message": "Material created successfully" }
```

### Example 2: Submit Assignment

```
Request:
  POST /api/assignments/submit/:assignmentId
  Headers: Authorization: Bearer <token>, SchoolId: <uuid>
  Body:
    {
      "schoolId": "...",
      "mediaIds": ["...", "..."]
    }

Handler (AssignmentHandler.Submit):
  1. Extract userID from JWT
  2. Validate input DTO
  3. Call service.Submit(submission, mediaIDs)

Service (AssignmentService.Submit):
  1. Upsert submission (1 per student per assignment)
  2. Link attachments
  3. Calculate isLate (if deadline provided)
  4. Return submission

Repository (AssignmentRepository.UpsertSubmission):
  1. UPSERT edv.submissions WHERE sbm_asg_id = ? AND sbm_usr_id = ?

Response:
  201 Created (or 200 OK if re-submit)
  { "message": "Submission created/updated" }
```

### Example 3: Get Assignments by Subject Class

```
Request:
  GET /api/assignments/subject-class/:subjectClassId?page=1&limit=20&search=math

Handler:
  1. Extract params
  2. Call service.GetAssignmentsBySubjectClass(subjectClassId, search, page, limit)

Service:
  1. Query DB
  2. Attach media/attachments for each result

Repository:
  1. SELECT * FROM edv.assignments
       WHERE asg_scl_id = ? AND asg_title ILIKE ? ...
       LIMIT 20 OFFSET 0
  2. Preload Category, SubjectClass.Subject

Response:
  200 OK
  {
    "subjectClass": { "id": "...", "subjectName": "..." },
    "data": {
      "data": [{ Assignment }, ...],
      "totalItems": 42,
      "page": 1,
      "limit": 20,
      "totalPages": 3
    }
  }
```

---

## 14. ROUTING PRECEDENCE WARNING

### Risky Pattern Detected

```
Route Definition Order (main.go):
1. assignmentAPI.GET("/my-submission/:assignmentId", ...) ✓
2. assignmentAPI.GET("/status/:id", ...)                   ✓
3. assignmentAPI.GET("/:assignmentId", ...)                ⚠️ GREEDY

Problem:
GET /api/assignments/status/:id
  → Matches route 3 with :assignmentId="status"
  → Route 2 never reached

Solution:
Reorder routes: specific → generic
1. GET /status/:id
2. GET /my-submission/:assignmentId
3. GET /:assignmentId
```

---

## 15. SUMMARY TABLE: Layer Responsibilities

| Layer          | Responsibility                                       | Key Files                       | Count  |
| -------------- | ---------------------------------------------------- | ------------------------------- | ------ |
| **Handler**    | HTTP parsing, auth checks, response mapping          | `internal/handler/*_handler.go` | 23     |
| **Service**    | Business logic, notification triggers, orchestration | `internal/service/*_service.go` | 21     |
| **Repository** | Database queries, soft deletes, preloading           | `internal/repository/*_repo.go` | 22     |
| **Domain**     | Entity models, constants, table names                | `internal/domain/*.go`          | 19     |
| **DTO**        | Input validation, response mapping                   | `internal/dto/*.go`             | 19+    |
| **Middleware** | Auth JWT, RBAC roles, school context                 | `internal/middleware/*.go`      | 2      |
| **Storage**    | File uploads, providers                              | `internal/storage/`             | Custom |

---

## 16. CRITICAL FINDINGS & NOTES

### Known Issues (from AGENT.md & review)

1. ⚠️ **Route Ordering**: GET /assignments/status/:id may be swallowed by GET /assignments/:assignmentId
2. ⚠️ **Schema Inconsistency**: Docs reference asg_scl_id but code unclear on class vs subject class
3. ⚠️ **No Unit Tests**: 0 \*\_test.go files exist
4. ⚠️ **gofmt Non-Compliance**: Many files not formatted
5. ⚠️ **Orphaned Files**: /backend/tmp/main binary artifact in repo

### Unimplemented TODOs (deferred to future)

1. ❌ **Real File Storage**: S3/Supabase integration (stub only)
2. ❌ **Chat WebSocket**: Realtime messaging (schema ready, routes pending)
3. ❌ **Student Notes**: Personal notes per material (schema ready, routes pending)
4. ❌ **Email Notifications**: Delivery mechanism (in-app only)
5. ❌ **Signed URLs**: Private/secure file download
6. ❌ **Thumbnail Generation**: Auto-generate from videos
7. ❌ **Nested Comments**: Comment threading (flat only)

### Design Decisions

1. ✅ **Soft Deletes**: All entities except linking tables
2. ✅ **Upsert Semantics**: Submissions and Assessments (last-write-wins)
3. ✅ **Best-Effort Notifications**: Don't cascade on failure
4. ✅ **Manual Attachment Linking**: No auto-cascade, manual link/unlink
5. ✅ **Polymorphic Comments**: SourceType + SourceID, no foreign key constraint
