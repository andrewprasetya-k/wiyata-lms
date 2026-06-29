# 📚 Wiyata LMS - API Endpoints Summary

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
- `POST /super-admin/school-bootstrap` - Atomically create school tenant and assign initial school admin (system super_admin only)
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

## 👥 Admin School Members

- `GET /admin/school-members` - List active-school members only (admin only)
- `POST /admin/school-members` - Add/reuse/restore one active-school member (admin only)
- `DELETE /admin/school-members/:schoolUserId` - Soft-delete membership from active school (admin only)
- `PATCH /admin/school-members/:schoolUserId/restore` - Restore soft-deleted active-school membership (admin only)
- `POST /admin/school-members/import/preview` - Validate CSV import rows for active-school members (admin only)
- `POST /admin/school-members/import/commit` - Import active-school members from validated rows (admin only)

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

Subject payloads support optional `color` for visual identity. Accepted values are hex colors in `#RGB`, `#RRGGBB`, or `#RRGGBBAA` format; empty color remains valid and lets the frontend use fallback colors.

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

- `GET /notes` - List the current student's accessible material notes across active enrolled classes
- `GET /notes/material/:materialId` - Get the current student's private note, or `{ "note": null }`
- `PUT /notes/material/:materialId` - Create or update one plain-text note for the accessible material
- `DELETE /notes/material/:materialId` - Hard-delete the current student's note for the accessible material
- `GET /notes/subject-class/:subjectClassId` - List the current student's material notes for an actively enrolled subject class

Notes are material-only for MVP. They are scoped to the JWT user and active `SchoolId`, require active student enrollment in the material's class, and are not accessible to teacher/admin roles.

## 🗓️ Academic Activity

- `GET /academic-activity?from=YYYY-MM-DD&to=YYYY-MM-DD` - Get normalized, actionable academic activity for the current active-school student or teacher. Defaults to today through today + 7 days and supports a maximum 60-day range.

Academic Activity is not notification/feed/calendar storage. It merges academic sources into one response for future My Day, Activity, and calendar marker UI. Student activity currently includes assignment deadlines, new materials, class feed posts, and graded assignments. Teacher activity currently includes received submissions, pending review submissions, assignment deadlines, and feed comments. Access follows active school membership plus active enrollment/teaching rules.

## 📰 Feeds (Announcements)

- `POST /feeds` - Create active-school class feed (admin or teacher who teaches the class)
- `GET /feeds/unread-count` - Get current user's active-school unread feed notification count (`feed_posted` and `comment_added`)
- `PATCH /feeds/read` - Mark current user's active-school feed-related notifications as read
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

## 💬 Chat

- `GET /ws/chat?token=&schoolId=` - Connect WebSocket realtime transport for chat `new_message`, `message_read`, and `room_updated` events
- `GET /chat/rooms?search=` - List room sekolah, grup kustom yang bisa diakses, dan direct message aktif; `search` juga mencocokkan nama/email target DM
- `GET /chat/members?search=&excludeRoomId=` - Search active school members for chat picker grup/DM, optionally excluding active members of a room
- `POST /chat/school/open` - Open or create the active school's main chat room
- `POST /chat/dm/open` - Open or reuse a direct message room with one active member in the same school
- `POST /chat/groups` - Create custom group room with active school members
- `GET /chat/groups/:roomId` - Get group room info, creator, admins, and active members
- `PATCH /chat/groups/:roomId` - Rename custom group room as a group admin
- `POST /chat/groups/:roomId/leave` - Leave a custom group room
- `POST /chat/groups/:roomId/members` - Add or restore active school members into a custom group room
- `DELETE /chat/groups/:roomId/members/:userId` - Remove a member from a custom group room
- `GET /chat/rooms/:roomId/read-summary` - Get per-member read receipt summary for an accessible room
- `GET /chat/rooms/:roomId/messages` - List text/file messages with `limit` and `before` pagination
- `POST /chat/rooms/:roomId/messages` - Create message with optional upload-first `mediaIds` and return canonical message DTO
- `PATCH /chat/rooms/:roomId/read` - Mark accessible room as read with optional validated `lastReadMessageId`

Chat MVP supports text messages and upload-first file/image attachments. Active school admins, teachers, and
students can participate in the school-wide room if their school membership is
active. Custom group rooms are limited to selected active school members through
`chat_room_members.left_at IS NULL`, with admin-only rename/member management and
automatic ownership transfer when the creator leaves. Direct message rooms are
limited to exactly two active members in the same active school and are reused
idempotently when the same pair opens DM again. Unread counts exclude messages
sent by the current user and are based on `chat_read_receipts.last_read_msg_id`
or `last_read_at`. WebSocket in Sprint 18B is event transport only for
`new_message`, `message_read`, and `room_updated`; message creation still uses
REST and polling remains as fallback. Attachment messages upload files through
`POST /api/medias/upload`, then send `mediaIds` to chat; `new_message` includes
attachment metadata. Current storage URLs may be public depending on provider,
with signed/protected downloads deferred.
It does not enable subject/class rooms, typing indicators, message delete, or
notifications.

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
