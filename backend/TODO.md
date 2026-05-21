# 🎓 Eduverse LMS - Development Progress

## ✅ Completed Features

1. ✅ Header pattern untuk bulk list responses (School, Class, SubjectClass)
2. ✅ Endpoint untuk members/enrollments dengan header
3. ✅ Error handling yang user-friendly (tidak expose database details)
4. ✅ Materials connected to SubjectClass dengan header
5. ✅ Complete CRUD operations untuk semua modules
6. ✅ Pagination & search untuk list endpoints yang besar
7. ✅ Assignment status tracking (submission statistics)
8. ✅ File upload support (multipart form & inline media)
9. ✅ Dashboard Statistics (Student, Teacher, Admin)
10. ✅ Authentication & Authorization (JWT middleware)
11. ✅ Implementasi auto get email dan user id dari middleware
12. ✅ Role-based Access Control (RBAC middleware)
13. ✅ Grade Book Implementation (configure weights, calculate final grades, grade reports)
14. ✅ Notification System (CRUD endpoints, unread count, mark as read)
15. ✅ Assignment route param correctness
16. ✅ Assignment ownership docs/schema sync
17. ✅ Auth/RBAC middleware hardening
18. ✅ Class delete protection (blocked when enrollments/subject classes exist)
19. ✅ Subject delete protection (blocked when subject classes exist)
20. ✅ Media storage abstraction + Supabase provider
21. ✅ Media upload endpoint wired to storage
22. ✅ Media delete storage cleanup

## 🚀 High Priority (Critical for Production)

- [ ] **Assignment Extensions**: Student request extension, teacher approve/reject, extended deadline logic
  - Files needed: Update `internal/domain/assignment.go` (add extension fields to Submission), update DTOs, service, handler
  - Endpoints: `POST /assignments/submit/:submissionId/request-extension`, `PATCH /assignments/submit/:submissionId/review-extension`
  - Database: Add extension fields to submissions table
  
- [x] **Notification Triggers Integration**: Auto-create notifications untuk:
  - New assignment created → notify students in class ✅
  - Assignment graded → notify student who submitted ✅
  - New comment added → notify content owner ✅
  - New material added → notify students in class ✅
  - New feed posted → notify class members ✅
  - Best-effort: aksi utama tidak gagal jika notif gagal ✅
  - Self-notification prevention untuk feed dan comment ✅
  - Files updated: `assignment_service.go`, `material_service.go`, `feed_service.go`, `comment_service.go`
  - New repo: `content_owner_repo.go`
  - New repo methods: `enrollment_repo.go` (GetStudentUserIDsByClass, GetMemberUserIDsByClass)

- [~] **File Upload Integration**: S3/Supabase storage untuk media files
  - [x] Real file upload via `POST /media/upload`
  - [x] Supabase storage provider wired by `STORAGE_PROVIDER=supabase`
  - [x] Disabled storage returns 501 when upload is not configured
  - [x] File validation (size limit, safe object path validation)
  - [x] Storage cleanup if metadata save fails
  - [x] Storage cleanup when media is deleted
  - [x] Multipart material upload wired to storage (no more placeholder URL)
  - [ ] Generate signed URLs untuk download
  - [ ] Thumbnail generation

## 📊 Analytics & Reporting (Medium Priority)

- [x] **Dashboard Statistics**:
  - Student: pending assignments, average scores, upcoming deadlines ✅
  - Teacher: pending reviews, submission rates, class performance ✅
  - Admin: school statistics, enrollment trends ✅
- [x] **Grade Book Implementation**:
  - Configure weighted grades using assessment_weights table ✅
  - Auto-calculate final grades per student per subject ✅
  - Generate grade reports per class ✅
  - Letter grade conversion (A, B, C, D, E) ✅
- [ ] **Grade Report / Transcript Export**:
  - Export individual student transcript to PDF
  - Export class grade report to Excel
  - Generate report cards per term/year

- [ ] **Activity Feed / Timeline**:
  - Recent assignments, submissions, grades, comments
  - Per class or per user feed

- [ ] **Auto-create Feed Posts from Teacher Actions**:
  - Create class feed post when teacher creates a material
  - Create class feed post when teacher creates an assignment
  - Feed remains the class-level communication surface; no type badge is required for now
  - Keep notification triggers best-effort and separate from feed creation behavior

- [ ] **Active Class Context for Frontend/API Flow**:
  - Frontend needs a clear active class context/store for student workspace navigation
  - Class is the context selector; subjects are the daily content surface inside that context
  - API calls should consistently derive subject-class/material/assignment/feed context from the active class flow
  - Avoid fake class/chat/notes data while context contracts are incomplete

## 🎓 Academic Features (Medium Priority)

- [ ] **Rich Text Support**: HTML content untuk descriptions (materials, assignments, feeds)
  - Update validation untuk accept HTML
  - Sanitize HTML input (prevent XSS)
  - Frontend rich text editor integration

- [ ] **Nested Comments**: Reply to comments functionality
  - Add parent_comment_id field to comments table
  - Update repository untuk fetch nested structure
  - Response DTO dengan nested comments

- [ ] **Activity Feed Enhancement**: Unified stream untuk class activities
  - Combine materials + assignments + feeds + comments dalam satu feed
  - Pin/unpin posts functionality
  - Filter by type (assignments only, announcements only)

- [ ] **Class Schedule / Timetable**:
  - Weekly schedule per class
  - Teacher schedule view
  - Room management

- [ ] **Material Progress Analytics**:
  - Track completion rates
  - Most viewed materials
  - Student engagement metrics
  - Do not mark material as completed automatically just because it was opened
  - If passive tracking is added, use a separate `last_opened_at`-style signal; completion must be explicit or governed by a clear rule

## 🔧 Enhancement Features (Low Priority)

- [ ] **Bulk Operations**:
  - Bulk grade assignments
  - Bulk enroll students
  - Bulk delete submissions

- [ ] **Export Functionality**:
  - Export grades to Excel/PDF
  - Export class rosters
  - Export submission reports

- [ ] **Leaderboard / Rankings**:
  - Top students per class/subject
  - Most active students
  - Gamification elements

- [ ] **Notification Preferences**:
  - User settings for notification types
  - Email vs in-app preferences
<!-- 
## 🔮 Future Enhancements

- [ ] **Attendance System**: Track student attendance per session
- [ ] **Quiz/Exam Module**: Multiple choice, auto-grading, time limits
- [ ] **Discussion Forum**: Thread-based discussions per class
- [ ] **Parent Portal**: Parent accounts to view child's progress
- [ ] **Real-time Features**: WebSocket for live updates
- [ ] **Email Service**: Password reset, notifications via email
- [ ] **Advanced Search**: Full-text search across materials and assignments -->
