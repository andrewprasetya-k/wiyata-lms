# 📚 Eduverse LMS - API Endpoints Summary

Base URL: `http://localhost:8080/api`

## 🔐 Authentication

**Public Endpoints (No Auth Required):**
- `POST /login` - User login
- `POST /register` - Public user self-registration (plain global account only)

**All other endpoints require JWT authentication.**

**Authentication Header:**
```
Authorization: Bearer <your-jwt-token>
```

---

## 🏫 Schools
- `POST /schools` - Create school
- `GET /schools` - List all schools (with pagination)
- `GET /schools/summary` - Get schools summary
- `GET /schools/check-code/:schoolCode` - Check code availability
- `GET /schools/:schoolCode` - Get school by code
- `PATCH /schools/:schoolCode` - Update school
- `PATCH /schools/restore/:schoolCode` - Restore deleted school
- `DELETE /schools/:schoolCode` - Soft delete school
- `DELETE /schools/permanent/:schoolCode` - Hard delete school

## 📅 Academic Years
- `POST /academic-years` - Create academic year
- `GET /academic-years` - List all academic years
- `GET /academic-years/:id` - Get by ID
- `GET /academic-years/school/:schoolCode` - Get by school
- `PATCH /academic-years/:id` - Update academic year
- `PATCH /academic-years/activate/:id` - Activate academic year
- `PATCH /academic-years/deactivate/:id` - Deactivate academic year
- `DELETE /academic-years/:id` - Delete academic year

## 📆 Terms (Semester)
- `POST /terms` - Create term
- `GET /terms` - List all terms
- `GET /terms/:id` - Get by ID
- `GET /terms/academic-year/:academicYearId` - Get by academic year
- `PATCH /terms/:id` - Update term
- `PATCH /terms/activate/:id` - Activate term
- `PATCH /terms/deactivate/:id` - Deactivate term
- `DELETE /terms/:id` - Delete term

## 👤 Users
- `POST /users` - Create global user (system super_admin only)
- `GET /users` - List/search global users (admin/super_admin)
- `GET /users/:id` - Get user by ID (system super_admin only)
- `PATCH /users/:id` - Update user (system super_admin only)
- `PATCH /users/change-password/:id` - Change password by ID (system super_admin only; future `/me/change-password` recommended)
- `DELETE /users/:id` - Delete user (system super_admin only)

## 🏢 School Users (User-School Relationship)
- `POST /school-users/enroll` - Enroll user to school
- `GET /school-users/school/:schoolCode` - Get members by school
- `GET /school-users/user/:userId` - Get schools by user
- `DELETE /school-users/:userId` - Unenroll user from school

## 📚 Subjects
- `POST /subjects` - Create subject
- `GET /subjects` - List all subjects
- `GET /subjects/:id` - Get by ID
- `GET /subjects/school/:schoolCode` - Get by school
- `GET /subjects/school/:schoolCode/:subjectCode` - Get by code
- `PATCH /subjects/:id` - Update subject
- `DELETE /subjects/:id` - Delete subject

## 🔐 RBAC (Roles & Permissions)
### Roles
- `POST /rbac/roles` - Create role
- `GET /rbac/roles` - List all roles
- `GET /rbac/roles/:id` - Get role by ID
- `PATCH /rbac/roles/:id` - Update role
- `DELETE /rbac/roles/:id` - Delete role

### User Roles
- `POST /rbac/user-roles` - Assign role to user
- `DELETE /rbac/user-roles` - Remove role from user
- `GET /rbac/user-roles/:schoolUserId` - Get user roles
- `PATCH /rbac/user-roles/:schoolUserId` - Update user roles

## 🎓 Classes
- `POST /classes` - Create class
- `GET /classes` - List all classes (with pagination & search)
- `GET /classes/:id` - Get class by ID
- `PATCH /classes/:id` - Update class
- `DELETE /classes/:id` - Delete class

## 📖 Subject Classes (Teacher Assignment)
- `POST /subject-classes/assign` - Assign active-school subject and eligible teacher school_user to active-school class (admin)
- `GET /subject-classes/my-teaching` - Get active subject classes taught by current teacher with active teacher class enrollment
- `GET /subject-classes/class/:classId` - Get subject classes by active-school class
- `GET /subject-classes/:id` - Get subject class detail within active school
- `PATCH /subject-classes/:id` - Update subject class assignment within active school (admin)
- `DELETE /subject-classes/:id` - Unassign empty subject class within active school (admin; blocked if materials or assignments exist)

## 👥 Enrollments (Class Members)
- `POST /enrollments` - Enroll or reactivate active-school members to an active-school class (admin, bulk; reactivation preserves original joined_at)
- `GET /enrollments/class/:classId` - Get active class members by class
- `GET /enrollments/member/:schoolUserId` - Get active classes by member
- `GET /enrollments/:id` - Get enrollment by ID within active school
- `PATCH /enrollments/:id` - Update enrollment role within active school (admin)
- `DELETE /enrollments/:id` - Soft-unenroll member within active school by setting `left_at` (admin; blocks teacher if still assigned to subject_class)

## 📁 Media & Files
- `POST /medias/upload` - Upload active-school file (multipart form; owner from JWT)
- `POST /medias/metadata` - Record active-school media metadata
- `GET /medias/:id` - Get media by ID
- `DELETE /medias/:id` - Delete active-school media record (admin or uploader)
- Media attached through `mediaIds` must exist, belong to the active school, and be attachable by the current actor

## 📖 Materials (Learning Content)
- `POST /materials` - Create material for current teacher-owned subject class (JSON or multipart form)
- `GET /materials` - List materials for accessible `subjectClassId`
- `GET /materials/:id` - Get accessible material by ID
- `PATCH /materials/:id` - Update active-school material (admin or owning teacher)
- `DELETE /materials/:id` - Delete active-school material (admin or owning teacher)
- `POST /materials/progress` - Update material progress

## 🗒️ Student Material Notes
- `GET /notes/material/:materialId` - Get the current student's private note, or `{ "note": null }`
- `PUT /notes/material/:materialId` - Create or update one plain-text note for the accessible material
- `DELETE /notes/material/:materialId` - Hard-delete the current student's note for the accessible material
- `GET /notes/subject-class/:subjectClassId` - List the current student's material notes for an actively enrolled subject class

Notes are material-only for MVP. They are scoped to the JWT user and active `SchoolId`, require active student enrollment in the material's class, and are not accessible to teacher/admin roles.

## 📰 Feeds (Announcements)
- `POST /feeds` - Create active-school class feed (admin or teacher who teaches the class)
- `GET /feeds/class/:classId` - Get active-school class feed (admin, teacher who teaches the class, or active enrolled student)
- `GET /feeds/:id` - Get accessible active-school feed by ID
- `PATCH /feeds/:id` - Update active-school feed (admin or owning teacher who teaches the class)
- `DELETE /feeds/:id` - Soft-delete active-school feed (admin or owning teacher who teaches the class)

## 💬 Comments
- `POST /comments` - Create feed-only comment in active school
- `GET /comments?type=feed&id=` - Get accessible feed comments
- `GET /comments/:id` - Get accessible active-school feed comment by ID
- `PATCH /comments/:id` - Update own active-school feed comment
- `DELETE /comments/:id` - Delete own feed comment, or admin-delete active-school comment

## 📝 Assignments & Grading
### Categories
- `POST /assignments/categories` - Create active-school category
- `GET /assignments/categories/school/:schoolCode` - Get categories by school

### Assignments
- `POST /assignments` - Create assignment for current teacher-owned subject class with active-school category
- `GET /assignments/subject-class/:subjectClassId` - Get assignments for accessible subject class
- `GET /assignments/subject-class/submissions/:subjectClassId` - Get submissions grouped by assignment for current teacher subject class
- `GET /assignments/teacher-assignments` - Get teacher global assignments overview across current teacher-owned subject classes with active teacher enrollment
- `GET /assignments/teacher-submissions` - Get teacher global submissions inbox across current teacher-owned subject classes
- `GET /assignments/student-assignments` - Get student global assignments list across active enrolled classes
- `GET /assignments/student/:assignmentId` - Get one student-safe assignment detail from an active enrolled subject class
- `GET /assignments/my-submission/:assignmentId` - Get the current active student's own submission status and attachments
- `GET /assignments/:assignmentId` - Get assignment with submissions for current teacher-owned subject class
- `GET /assignments/status/:id` - Get assignment status
- `PATCH /assignments/:id` - Update active-school assignment (admin or owning teacher)
- `DELETE /assignments/:id` - Delete active-school assignment (admin or owning teacher)

### Submissions
- `POST /assignments/submit/:assignmentId` - Submit assignment as current enrolled student
- `GET /assignments/submit/:submissionId` - Get submission by ID for current teacher-owned subject class
- `PATCH /assignments/submit/:submissionId` - Update current student's own submission
- `DELETE /assignments/submit/:submissionId` - Delete current student's own submission

### Assessments (Grading)
- `POST /assignments/assess/:submissionId` - Grade submission for current teacher-owned subject class
- `PATCH /assignments/assess/:submissionId` - Update assessment for current teacher-owned subject class
- `DELETE /assignments/assess/:submissionId` - Delete assessment for current teacher-owned subject class

## 📊 Logs
- `GET /logs/school/:schoolId` - Get logs by school

## 📊 Grade Book
- `GET /grades/my-grades/:classId` - Get current student's gradebook by active class, including provisional weighted grade when weights and graded assignments exist
- `POST /grades/weights` - Admin-only configure active-school subject-level assessment weights
- `GET /grades/weights/subject/:subjectId` - Get active-school subject weights
- `GET /grades/class/:classId/subject/:subjectId` - Get class grade report

## 🔔 Notifications
- `GET /notifications` - Get current user's notifications (with pagination)
- `GET /notifications/unread-count` - Get current user's unread count
- `PATCH /notifications/read/:id` - Mark current user's notification as read
- `PATCH /notifications/read-all` - Mark all current user's notifications as read
- `DELETE /notifications/:id` - Delete current user's notification

## 📈 Dashboard
- `GET /dashboard/student/:userId` - Get student dashboard (userId = usr_id)
- `GET /dashboard/teacher/:schoolUserId` - Get teacher dashboard (schoolUserId = scu_id)
- `GET /dashboard/admin/:schoolId` - Get admin dashboard

---

## 🔑 Key Features

### Authentication
- JWT-based authentication
- Token expiry: 24 hours
- Public endpoints: login, register
- All other endpoints protected

### Pagination
Most list endpoints support:
- `?page=1` - Page number (default: 1)
- `?limit=20` - Items per page (default: varies)
- `?search=keyword` - Search filter

### Response Patterns
- **Header Pattern**: List responses include parent context (School, Class, SubjectClass)
- **Soft Delete**: Most delete operations are soft deletes (can be restored)
- **Student Material Notes**: Note deletion is an intentional hard delete for the material-only MVP
- **Upsert Logic**: Submissions and assessments auto-update if already exist

### File Handling
- **Multipart Upload**: Direct file upload with auto-detection
- **Metadata Only**: Record already-uploaded files (Supabase/S3)
- **Inline Media**: Create media records within material/assignment creation

### Error Handling
- Standardized error responses (no raw DB errors exposed)
- Validation errors with field-specific messages
- Proper HTTP status codes

---

**Last Updated:** 2026-06-24
**Version:** 1.3
