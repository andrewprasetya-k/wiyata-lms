# 📚 Eduverse LMS - API Endpoints Summary

Base URL: `http://localhost:8080/api`

## 🔐 Authentication

**Public Endpoints (No Auth Required):**
- `POST /login` - User login
- `POST /register` - User registration

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
- `POST /users` - Create user
- `GET /users` - List all users (with pagination & search)
- `GET /users/:id` - Get user by ID
- `PATCH /users/:id` - Update user
- `PATCH /users/change-password/:id` - Change password
- `DELETE /users/:id` - Delete user

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
- `POST /subject-classes/assign` - Assign subject & teacher to class
- `GET /subject-classes/class/:classId` - Get subjects by class
- `GET /subject-classes/:id` - Get assignment detail
- `PATCH /subject-classes/:id` - Update assignment
- `DELETE /subject-classes/:id` - Unassign subject from class

## 👥 Enrollments (Class Members)
- `POST /enrollments` - Enroll members to class (bulk)
- `GET /enrollments/class/:classId` - Get members by class
- `GET /enrollments/member/:schoolUserId` - Get classes by member
- `GET /enrollments/:id` - Get enrollment by ID
- `PATCH /enrollments/:id` - Update enrollment role
- `DELETE /enrollments/:id` - Unenroll member

## 📁 Media & Files
- `POST /medias/upload` - Upload file (multipart form)
- `POST /medias/metadata` - Record media metadata
- `GET /medias/:id` - Get media by ID
- `DELETE /medias/:id` - Delete media record

## 📖 Materials (Learning Content)
- `POST /materials` - Create material (JSON or multipart form)
- `GET /materials` - List materials (with pagination & search)
- `GET /materials/:id` - Get material by ID
- `PATCH /materials/:id` - Update material
- `DELETE /materials/:id` - Delete material
- `POST /materials/progress` - Update material progress

## 📰 Feeds (Announcements)
- `POST /feeds` - Create feed
- `GET /feeds/class/:classId` - Get feeds by class
- `GET /feeds/:id` - Get feed by ID
- `PATCH /feeds/:id` - Update feed
- `DELETE /feeds/:id` - Delete feed

## 💬 Comments
- `POST /comments` - Create comment
- `GET /comments?type=&id=` - Get comments by source
- `GET /comments/:id` - Get comment by ID
- `PATCH /comments/:id` - Update comment
- `DELETE /comments/:id` - Delete comment

## 📝 Assignments & Grading
### Categories
- `POST /assignments/categories` - Create category
- `GET /assignments/categories/school/:schoolCode` - Get categories by school

### Assignments
- `POST /assignments` - Create assignment
- `GET /assignments/subject-class/:subjectClassId` - Get by subject class
- `GET /assignments/:assignmentId` - Get assignment with submissions
- `GET /assignments/status/:id` - Get assignment status
- `PATCH /assignments/:id` - Update assignment
- `DELETE /assignments/:id` - Delete assignment

### Submissions
- `POST /assignments/submit/:assignmentId` - Submit assignment
- `GET /assignments/submit/:submissionId` - Get submission by ID
- `PATCH /assignments/submit/:submissionId` - Update submission
- `DELETE /assignments/submit/:submissionId` - Delete submission

### Assessments (Grading)
- `POST /assignments/assess/:submissionId` - Grade submission
- `PATCH /assignments/assess/:submissionId` - Update assessment
- `DELETE /assignments/assess/:submissionId` - Delete assessment

## 📊 Logs
- `GET /logs/school/:schoolId` - Get logs by school

## 📊 Grade Book
- `POST /grades/weights` - Configure assessment weights
- `GET /grades/weights/subject/:subjectId` - Get weights by subject
- `GET /grades/student/:userId/subject/:subjectId` - Get student final grade
- `GET /grades/class/:classId/subject/:subjectId` - Get class grade report

## 🔔 Notifications
- `GET /notifications` - Get user notifications (with pagination)
- `GET /notifications/unread-count` - Get unread count
- `PATCH /notifications/read/:id` - Mark notification as read
- `PATCH /notifications/read-all` - Mark all notifications as read
- `DELETE /notifications/:id` - Delete notification

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

**Last Updated:** 2026-03-12
**Version:** 1.2
