# 📊 Dashboard API Documentation

Base URL: `/api/dashboard`

## 1. Student Dashboard
Get dashboard statistics for a student.

- **URL:** `/student/:userId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("student")`
- **Ownership:** Handler verifies `:userId` equals the JWT-authenticated user's ID. A student can only view their own dashboard.
- **Response:**
```json
{
  "pendingAssignments": 5,
  "upcomingDeadlines": [
    {
      "assignmentId": "uuid",
      "assignmentTitle": "Quiz Chapter 1",
      "subjectName": "Matematika",
      "deadline": "2026-03-01T16:59:00Z",
      "isSubmitted": false
    }
  ],
  "averageScore": 85.5,
  "completedMaterials": 12,
  "totalMaterials": 20
}
```

**Metrics:**
- `pendingAssignments`: Number of assignments not yet submitted
- `upcomingDeadlines`: Next 5 assignments ordered by deadline
- `averageScore`: Average score across all graded submissions
- `completedMaterials`: Number of materials marked as completed
- `totalMaterials`: Total materials available to the student

---

## 2. Teacher Dashboard
Get dashboard statistics for a teacher.

- **URL:** `/teacher/:schoolUserId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("teacher")`
- **Ownership:** Handler verifies `:schoolUserId` equals the `school_user_id` set in context by `RequireSchoolMember`. A teacher can only view their own dashboard.
- **Note:** `schoolUserId` is the `school_user_id` (scu_id), NOT user_id
- **Response:**
```json
{
  "pendingReviews": 15,
  "totalStudents": 120,
  "submissionRate": 78.5,
  "classPerformance": [
    {
      "classId": "uuid",
      "className": "12 IPA 1",
      "subjectName": "Matematika",
      "averageScore": 82.3,
      "submissionRate": 85.0,
      "totalStudents": 30
    }
  ]
}
```

**Metrics:**
- `pendingReviews`: Number of submissions waiting for grading
- `totalStudents`: Total unique students across all teacher's classes
- `submissionRate`: Overall submission rate across all assignments (%), calculated as submitted active-student assignment slots divided by total eligible active-student assignment slots. Active students are class enrollments with `left_at IS NULL`.
- `classPerformance`: Performance breakdown per class/subject
- `classPerformance[].submissionRate`: Per class/subject submission rate using the same eligible active-student assignment slot denominator.

---

## 3. Admin Dashboard
Get dashboard statistics for school admin.

- **URL:** `/admin/:schoolId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("admin")`
- **Ownership:** Handler verifies `:schoolId` equals the active `school_id` from context. An admin can only view their own school's dashboard.
- **Response:**
```json
{
  "totalStudents": 450,
  "totalTeachers": 35,
  "totalClasses": 15,
  "activeClasses": 12,
  "enrollmentTrends": [
    {
      "className": "12 IPA 1",
      "totalEnrolled": 32,
      "teachers": 2,
      "students": 30
    }
  ],
  "recentActivities": [
    {
      "userName": "John Doe",
      "action": "Created new assignment",
      "timestamp": "2026-02-24T03:30:00Z"
    }
  ],
  "classesWithoutTeacher": [
    { "classId": "uuid", "className": "12 IPA 1" }
  ],
  "classesWithoutTeacherTotal": 3,
  "contentLessSubjectClasses": [
    { "subjectClassId": "uuid", "className": "12 IPA 1", "subjectName": "Fisika" }
  ],
  "contentLessSubjectClassesTotal": 2,
  "subjectsWithoutAssessmentWeight": [
    { "subjectId": "uuid", "subjectName": "Kimia" }
  ],
  "subjectsWithoutAssessmentWeightTotal": 1,
  "backlogTotal": 24,
  "backlogClasses": [
    { "classId": "uuid", "className": "12 IPA 1", "backlogCount": 9 }
  ],
  "schoolPerformanceRollup": []
}
```

**Metrics (Phase 6/7 — original admin dashboard):**
- `totalStudents`: Total students enrolled in the school
- `totalTeachers`: Total teachers assigned to classes
- `totalClasses`: Total classes (including inactive)
- `activeClasses`: Only active classes
- `enrollmentTrends`: Enrollment breakdown per class
- `recentActivities`: Last 10 activities from logs

**Metrics (Phase 7 Batch 1 — added `dashboard_repo.GetClassesWithoutTeacher`):**
- `classesWithoutTeacher` / `classesWithoutTeacherTotal`: Top 5 active classes with zero `subject_classes` rows (i.e. no subject+teacher assignment yet), plus the total count. `subject_classes.scl_scu_id` is non-nullable, so "no teacher" means the class has no `subject_classes` row at all, not a null-teacher row.

**Metrics (Phase 7 Batch 2 — added `GetContentLessSubjectClasses`, `GetSubjectsWithoutAssessmentWeight`):**
- `contentLessSubjectClasses` / `contentLessSubjectClassesTotal`: Top 5 subject-classes (within active classes) with neither materials nor assignments, plus total count. Reuses the same "has content" definition as `SubjectClassRepository.HasSubjectClassContent` (materials OR assignments, school-scoped, non-deleted) — not a new business rule.
- `subjectsWithoutAssessmentWeight` / `subjectsWithoutAssessmentWeightTotal`: Top 5 subjects with zero `assessments_weights` rows, plus total count. `grade_service.ConfigureWeights` only ever persists a weight set summing to exactly 100, so a subject is either fully configured or has no rows at all — never partially configured — making a simple zero-rows check correct and complete.

**Metrics (Phase 7 Batch 3 — added `GetGradingBacklog`, `GetSchoolPerformanceRollup`):**
- `backlogTotal` / `backlogClasses`: Total ungraded submissions school-wide (a submission with no matching `assessments` row — the same "waiting for grading" definition already used by `GetPendingReviewsCount` on the teacher dashboard, just re-scoped from one teacher to the whole school), plus the top 3 classes by backlog count.
- `schoolPerformanceRollup`: Backend method (`GetSchoolPerformanceRollup`) and DTO field exist and reuse the same `average_score` formula as the teacher dashboard's `GetClassPerformance` (weakest 5 subject-classes, ranked ascending, excluding subject-classes with zero graded assessments). **Not currently wired** — `dashboard_service.GetAdminDashboard` does not call this method, so the field always serializes as `null`/empty in the live response. See `docs/PROJECT_CONTEXT_HANDOFF.md` §26 for status.

---

## 4. Super Admin Dashboard
Get platform-wide dashboard data for the super admin.

- **URL:** `/super-admin`
- **Method:** `GET`
- **Auth:** `RequireRole("super_admin")` (no `RequireSchoolMember` — this is a platform-level view, not scoped to a school)
- **Response:**
```json
{
  "schoolsWithoutAdmin": [
    { "schoolId": "uuid", "schoolName": "SMA Contoh", "schoolCode": "ABC123", "createdAt": "2026-02-20T10:00:00Z" }
  ],
  "schoolsWithoutAdminTotal": 2,
  "schoolsWithoutSetup": [
    { "schoolId": "uuid", "schoolName": "SMA Lain", "schoolCode": "XYZ789", "createdAt": "2026-02-18T09:00:00Z" }
  ],
  "schoolsWithoutSetupTotal": 1,
  "schoolGrowthTrend": [
    { "period": "2025-09", "count": 2 },
    { "period": "2025-10", "count": 0 },
    { "period": "2025-11", "count": 1 },
    { "period": "2025-12", "count": 3 },
    { "period": "2026-01", "count": 1 },
    { "period": "2026-02", "count": 4 }
  ],
  "userGrowthTrend": [
    { "period": "2025-09", "count": 12 },
    { "period": "2025-10", "count": 8 },
    { "period": "2025-11", "count": 15 },
    { "period": "2025-12", "count": 20 },
    { "period": "2026-01", "count": 9 },
    { "period": "2026-02", "count": 25 }
  ]
}
```

**Metrics (Phase 7 Batch 1 — added `GetSchoolsWithoutAdmin`, `GetSchoolsWithoutSetup`):**
- `schoolsWithoutAdmin` / `schoolsWithoutAdminTotal`: Top 5 active schools with zero active `school_users` holding the `admin` role, plus total count.
- `schoolsWithoutSetup` / `schoolsWithoutSetupTotal`: Top 5 active schools with no active `academic_years` row, plus total count.

**Metrics (Phase 7 Batch 4 — added `GetSchoolGrowthTrend`, `GetUserGrowthTrend`):**
- `schoolGrowthTrend` / `userGrowthTrend`: Exactly 6 monthly points each, oldest → newest, `period` formatted `YYYY-MM`. Built with `generate_series` so a month with zero signups still appears as an explicit `0` rather than being skipped. Counts every school/user *created* in that month regardless of later soft-deletion — this is a historical event count, not a current-active count (`GetSchoolSummary` already covers current-active).

---

## Key Features

### Real-time Calculations
All metrics are calculated in real-time from the database:
- No caching required
- Always up-to-date data
- Efficient SQL queries with proper joins

### Role-based Views
Each dashboard is tailored to the user's role:
- **Student**: Focus on personal progress and deadlines
- **Teacher**: Focus on grading workload and class performance
- **Admin**: Focus on school-wide statistics, work-queue widgets, and grading backlog
- **Super Admin**: Focus on platform-wide tenant health and growth trends

### Performance Optimized
- Uses aggregated queries to minimize database load
- Limits on list results (e.g., top 5 deadlines, last 10 activities, top 3 backlog classes)
- Indexed columns for fast lookups
- All Phase 7 additions ride the existing `/admin/:schoolId` and `/super-admin` endpoints — no new endpoints were introduced across Batches 1–4

---

## Usage Examples

### Student checking their dashboard
```bash
GET /api/dashboard/student/123e4567-e89b-12d3-a456-426614174000
# Parameter: userId (usr_id from users table)
```

### Teacher viewing pending reviews
```bash
GET /api/dashboard/teacher/223e4567-e89b-12d3-a456-426614174000
# Parameter: schoolUserId (scu_id from school_users table)
```

### Admin monitoring school statistics
```bash
GET /api/dashboard/admin/323e4567-e89b-12d3-a456-426614174000
# Parameter: schoolId (sch_id from schools table)
```

### Super admin monitoring platform health
```bash
GET /api/dashboard/super-admin
# No path parameter — platform-wide, not scoped to a school.
# Requires super_admin role; no SchoolId header needed.
```

---

## Notes

- **Authentication Required**: All dashboard endpoints require JWT; student/teacher/admin additionally require school membership + matching role. Super admin requires the `super_admin` role only (no school membership check).
- **Ownership enforced at handler**: The URL parameter (`:userId`, `:schoolUserId`, `:schoolId`) is validated against the caller's JWT context on every request. Sending another user's ID returns `403 Forbidden`. Not applicable to `/super-admin` (no URL parameter).
- **Caching**: Consider adding Redis caching for admin dashboard if school is large.
- **Pagination**: Currently returns fixed limits (5 deadlines, 10 activities, 5 work-queue rows, 3 backlog classes, 6 trend months). Can be made configurable.
- **Single bundled response per role**: Each dashboard (student/teacher/admin/super-admin) returns everything the page needs in one response, rather than one endpoint per widget. Phase 7 (Batches 1–4) extended the existing `admin` and `super-admin` DTOs rather than adding new endpoints, per this convention.
