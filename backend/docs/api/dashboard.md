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

## 5. Security Dashboard (Phase 11.5)

Summary widgets for the audit-logging system's authentication/security
signals — failed logins, brute-force detection, password reset activity,
and suspicious activity (token reuse, recovery code use, repeated MFA
failures).

**Phase 11.5.2 — super_admin only.** The school-pinned endpoint
(`GET /admin/:schoolId/security`, `admin` role) shipped in Phase 11.5 was
**removed entirely** — both the route and the handler method
(`SecurityDashboardHandler.GetAdminSecurityDashboard`) — rather than left
in the codebase unregistered. A handler method is a thin routing adapter
with no reuse value once its route is gone: it hard-codes the
`admin`-scoped-to-own-school shape, which isn't what a future super_admin
drill-down (a different auth shape — super_admin viewing an *arbitrary*
school, not an admin viewing *their own*) would reuse anyway. Deleting it
avoids unreachable code sitting in the handler package that looks live but
isn't route-tested by anything.

- **URL:** `/super-admin/security`
- **Method:** `GET`
- **Auth:** `RequireRole("super_admin")` — unrestricted, platform-wide, no school scoping. This is now the **only** way to reach this dashboard.

**What was deliberately kept, and why it's a dead-code candidate today:**
`SecurityDashboardService.GetDashboard(schoolID *string)` and every
`SecurityRepository` method it calls (`CountByActions`,
`GroupFailedLoginsByEmail`/`ByIP`, `GetFailedLoginAttemptTimes`/`ByIP`,
`GetRecentByActions`, and the `scopeToSchool` join helper) still accept and
correctly handle a non-nil `schoolID`. None of that was touched — it's
exactly the reusable piece a future super_admin drill-down endpoint
(e.g. `GET /dashboard/super-admin/security/:schoolId`, still
`RequireSystemSuperAdmin`, just narrowing the same query) would call
directly with no service/repo changes. **However**, as of this phase, the
only live call site (`GetSuperAdminSecurityDashboard`) always passes `nil`
— so the `schoolID != nil` branch in every one of those methods is today
only exercised by `security_dashboard_service_test.go`
(`TestGetDashboard_PassesSchoolIDThroughToRepo` and friends), never by an
actual HTTP request. Flagging this explicitly rather than deleting it
per-instruction — if no super_admin drill-down feature materializes, this
is the dead code to revisit later.

**Response (both endpoints, same shape):**
```json
{
  "windowHours": 24,
  "generatedAt": "2026-07-24T10:00:00Z",
  "failedLoginCount": 12,
  "bruteForceIncidents": [
    { "targetType": "email", "target": "target@example.com", "failureCount": 7, "lastAttemptAt": "2026-07-24T09:45:00Z" },
    { "targetType": "ip", "target": "203.0.113.9", "failureCount": 9, "lastAttemptAt": "2026-07-24T09:50:00Z" }
  ],
  "passwordResetRequestedCount": 4,
  "passwordResetCompletedCount": 3,
  "suspiciousActivities": [
    {
      "logId": "uuid",
      "action": "auth.token.reuse_detected",
      "severity": "HIGH",
      "userId": "uuid",
      "userName": "Budi Santoso",
      "userEmail": "budi@example.com",
      "createdAt": "2026-07-24T08:12:00Z"
    }
  ]
}
```

**Important data-availability note:** every `auth.*` audit action is
`scope=platform` with no `log_sch_id` (see `log.md` §4 — none of these
events ever carry a school ID; login/password-reset are pre-authentication,
and MFA/token events are user-scoped, not school-scoped). The school-pinned
endpoint therefore cannot filter by `log_sch_id` like the rest of the audit
log surface — it resolves school membership indirectly instead, via
`SecurityRepository.scopeToSchool`: for rows with a `user_id` (MFA/token
events), it joins `edv.school_users` on that ID; for rows with no `user_id`
(`auth.login.failed`, `auth.password.reset.requested` — identity is
intentionally unknown at that point), it resolves the target account by
matching `log_metadata->>'email'` against `edv.users.usr_email`, then checks
that user's school membership. Both paths exclude soft-deleted
`school_users` rows.

**Brute-force definition:** a target (an email address OR a source IP) is
flagged as an incident when its `auth.login.failed` attempts include **at
least 5 failures within any 15-minute span** in the lookback window — not
merely 5 failures spread across the whole window, which would misclassify
an ordinary "forgot my password, a few tries over the day" pattern as an
attack. This threshold matches the rate-limit tiers already used elsewhere
in this codebase (`mfa_verify` and change-password both use 5 attempts / 15
minutes), rather than inventing a new one.

**IP capture (Phase 11.5.1):** `AuthService.logLoginFailed` — and every
other `auth.*` action where the request's IP was already available in
scope via `RefreshTokenMetadata` (`auth.login.success`, `member.login`,
`auth.token.refreshed`, `auth.token.reuse_detected`, `auth.mfa.verified`,
`auth.mfa.verify.failed` both call sites, `auth.mfa.recovery_code.used`) —
now populates `domain.ActorContext.IPAddress`, which `LogService.buildLogEntry`
writes into `domain.Log.IPAddress`. Two `auth.*` actions were **deliberately
left out** because doing so would require new handler→service plumbing (a
signature change), not just using a value that was already being passed
through and ignored: `auth.password.reset.requested`/`auth.password.reset.completed`
(`PasswordResetService.Request`/`Reset` take no request-metadata parameter
today) and `auth.logout`/`auth.session.revoked` (`AuthService.Logout`/`RevokeSession`
likewise take none — `auth.session.revoked`'s metadata already carries the
*revoked session's own* stored IP/user-agent, a different concept from the
current request's IP). Recommended as a follow-up if IP-based analysis is
wanted for those too, but out of scope here.

Both groupings run and are reported together — an account can be attacked
from many sources, and one source can attack many accounts, and neither
pattern subsumes the other. Every pre-existing log row (written before this
phase) has `ip_address = NULL` and is correctly excluded from **all**
IP-based queries by an explicit `WHERE ip_address IS NOT NULL` — it's
treated as "IP unknown," never grouped together with other NULLs as if they
shared an IP, and it still counts normally in the email-based grouping
(which never depended on IP).

**Suspicious activity definition:** the union of `auth.token.reuse_detected`
(HIGH — a refresh token was replayed after being rotated, i.e. likely
stolen), `auth.mfa.recovery_code.used` (HIGH — the user's authenticator app
access was lost or bypassed), and `auth.mfa.verify.failed` (MEDIUM — a wrong
TOTP/recovery code) — the only actions in the taxonomy representing a
credential/session integrity concern rather than routine account activity.
Returned newest-first, capped at 50 rows.

**Index added (migration `0010`):** `idx_logs_action_created_at` on
`(log_action, created_at DESC)`. None of the existing composite indexes
(`idx_logs_school_created_at`, `idx_logs_user_created_at`,
`idx_logs_severity`) serve this dashboard's query shape well — every widget
here filters by a fixed `log_action IN (...)` list plus a `created_at`
range, and these auth events carry neither a school ID nor (for
pre-authentication events) a user ID, and don't share one severity tier.

**Index added (migration `0011`, Phase 11.5.1):**
`idx_logs_action_ip_created_at` on `(log_action, ip_address, created_at DESC)
WHERE ip_address IS NOT NULL` — serves the new per-IP grouping/lookup
queries (`SecurityRepository.GroupFailedLoginsByIP`/
`GetFailedLoginAttemptTimesByIP`) directly, and the partial-index condition
means every pre-existing (necessarily NULL-IP) log row is permanently
excluded from this index rather than dead weight in it.

**No new audit action added.** Viewing this dashboard is a read, gated by
the same `RequireSchoolMember + RequireRole`/`RequireSystemSuperAdmin`
checks as every other sensitive read surface in this codebase — consistent
with how viewing `/logs` itself is not separately audited either.

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
