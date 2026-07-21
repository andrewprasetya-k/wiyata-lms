# Wiyata AI Handoff

Last verified against codebase: 2026-07-21.

This document is a curated read-first guide for future AI coding agents working on Wiyata. It is not a raw merge of all existing documentation. Use it to orient quickly, then verify implementation details in code and tests before changing behavior.

## 1. Start Here / Purpose

Wiyata has changed quickly across onboarding, multi-school identity, active role context, notifications, discussions, chat, and admin setup. Some historical docs are stale. This handoff captures current working contracts and the risks that matter most.

When documentation conflicts with implementation, inspect code and tests. Do not change working behavior merely to match stale documentation.

## 2. Project Overview

Wiyata is a multi-tenant academic workspace / LMS. It supports school onboarding, academic setup, class placement, subject-class workspaces, materials, assignments, submissions, assessment, feed, discussions, chat, notifications, and role-specific dashboards.

The product model is:

- Main dashboard content = workspace / things to do.
- Right rail = awareness / updates to monitor.
- School is the tenant root.
- User identity is global.
- Academic access is scoped through school memberships and roles.

## 3. Repository Structure

- `backend/` - Go/Gin/GORM API.
- `backend/cmd/api/main.go` - app entrypoint, routes, centralized CORS.
- `backend/internal/domain/` - GORM entities.
- `backend/internal/dto/` - request/response contracts.
- `backend/internal/handler/` - HTTP handlers.
- `backend/internal/service/` - business workflows.
- `backend/internal/repository/` - data access.
- `backend/internal/middleware/` - JWT, school context, RBAC.
- `backend/docs/api/` - focused API docs.
- `backend/schema.md` - DBML schema reference.
- `frontend/` - Vue 3 application.
- `frontend/src/services/` - API service wrappers.
- `frontend/src/stores/` - Pinia stores, especially auth context.
- `frontend/src/composables/` - shared UI/data composables.
- `frontend/src/pages/` - role-specific pages.
- `frontend/src/layouts/` and `frontend/src/components/layout/` - role shells, sidebar, context switcher.
- `docs/` - high-level analysis/reference docs.

## 4. Technology Stack

Backend:

- Go, Gin, GORM.
- PostgreSQL/Supabase.
- JWT authentication.
- bcrypt password hashing.
- SMTP via Go standard library with no-op fallback.
- Supabase-compatible media storage provider.

Frontend:

- Vue 3.
- TypeScript.
- Vite.
- Tailwind CSS.
- Pinia.
- Axios-style API client.

## 5. Backend Architecture

Backend code follows Handler → Service → Repository → Domain.

- Handlers parse input, bind DTOs, get auth context, and map responses.
- Services enforce business rules and orchestrate repositories, notifications, storage, and email.
- Repositories own GORM/SQL access.
- Domain models define persisted structure and table names.

Transaction ownership follows the same split: an operation confined to a single repository keeps its transaction inside that repository. An operation that must orchestrate several repositories atomically instead owns its transaction in the service layer, binding each repository to it with `repo.WithTx(tx)`. Services avoid raw GORM calls when an equivalent repository method already exists.

Notifications and email are generally best-effort. The main DB action should succeed even when notification/email delivery fails, unless a specific workflow says otherwise.

## 6. Frontend Architecture

Frontend pages call typed service wrappers. Shared cross-page state is kept in Pinia stores or singleton composables.

The auth store is the source of current active context. Role layouts use keyed `RouterView` so switching school/role remounts pages whose loaders are only in `onMounted`.

Avoid broad redesign. Match existing Wiyata visual style: warm academic workspace, white cards, subtle borders, restrained Tailwind.

## 7. Core Data Model

Key entities:

- `users` - global account identity.
- `school_users` - user membership in a school.
- `roles` / `user_roles` - roles attached to school membership or platform context.
- `schools` - tenant root.
- `academic_years`, `terms`, `classes`, `subjects`.
- `enrollments` - active/historical class placement, with `enr_role`.
- `subject_classes` - class + subject + teacher workspace.
- `materials`, `material_progress`.
- `assignments`, `submissions`, `assessments`.
- `feeds`, `comments`, `attachments`, `medias`.
- `notifications`.
- chat rooms/messages/attachments/read receipts.
- onboarding `email_verifications` and `invitations`. (`school_registration_requests` table still exists in the database but is no longer used by the application — see section 13.)

## 8. Multi-Tenant and Global Identity Model

User accounts are global by email. A user can belong to multiple schools through `school_users`.

Roles are not global for ordinary school users. They attach to school memberships. A user can have different roles across schools and may have multiple roles in one school according to the data model.

`super_admin` is a platform context, not a school role context.

Soft-deleted school memberships must not appear in login/context responses. Active school membership filtering is part of auth correctness.

## 9. Authentication and Authorization

JWT identifies the global user. School-scoped APIs also require school membership context.

Important middleware concepts:

- `AuthRequired` validates JWT.
- `RequireSchoolMember` validates active membership in `SchoolId`; sets `school_id` and `school_user_id` in gin context.
- `RequireRole` authorizes the selected active role when `Active-Role` is present.
- `RequireSystemSuperAdmin` protects platform routes.

Frontend role/context values are never trusted alone. Backend validates `SchoolId` and `Active-Role` against live DB membership/roles.

**Handler-level ownership check:** For resource mutations (PATCH/DELETE), after fetching the resource by ID the handler also verifies `resource.SchoolID == activeSchoolID` from gin context. This prevents cross-school mutations by users who hold a valid role in multiple schools. Pattern is applied on: academic year, term, subject, class, dashboard, and log endpoints. A mismatch returns `403 Forbidden`.

## 10. Active School and Active Role Context

Frontend active context is a union:

```ts
type ActiveContext =
  | {
      type: "school";
      schoolId: string;
      schoolUserId: string;
      role: "admin" | "teacher" | "student";
    }
  | { type: "platform"; role: "super_admin" };
```

Runtime behavior uses exactly one active school + one active school role. It no longer uses the union of all roles for route authorization.

Sidebar awareness now uses a shared SSE stream at `/api/events/sidebar?token=&schoolId=` to invalidate feed and notification badge counts without periodic polling. Chat room summary remains websocket-driven with visibility/context refresh fallback.
`auth.activeRole` is one selected role. `auth.activeRoles` and `auth.allRoles` are compatibility surfaces and should not be used to grant access to role pages.

## 11. Request Context Headers

For school context, frontend sends:

- `frontend/src/services/sidebarStream.ts` - shared SSE sidebar invalidation stream.

- `SchoolId: <schoolId>`
- `Active-Role: admin|teacher|student`

For platform `super_admin`, frontend omits school headers.

Backend CORS allowlist includes:

- `Authorization`
- `Content-Type`
- `SchoolId`
- `schoolid`
- `Active-Role`
- `active-role`

## 12. Context Initialization, Persistence, and Recovery

Authoritative context endpoint:

- `GET /api/me/context`

Frontend session persistence uses:

- `edv_active_context` for the selected context.
- compatibility keys such as `edv_active_school_id`, `edv_active_roles`, and `edv_active_class_id` where old code still needs them.

Important auth-store methods:

- `ensureUserContext()` initializes context once per app session and waits for an in-flight context request. Once `isContextInitialized` is true, it never fetches again on its own for the rest of that session.
- `refreshUserContext()` is a forced refresh — the only way to re-pull `GET /me/context` after the first automatic fetch. Used by `CreateSchool.vue` after school creation, by `VerifyEmail.vue` after a successful verify (if a session is already active in that browser), and by `useAuthContextSync()` (`frontend/src/composables/useAuthContextSync.ts`) on `visibilitychange`/`focus` with a 3s cooldown — the latter exists specifically so `auth.emailVerified` (and memberships/roles) catch up when the user verifies email or otherwise changes server-side state in a different tab/device and comes back.
- In-flight context request protection prevents concurrent route guards from racing.
- `switchContext(target)` validates target, persists it, increments `contextVersion`, emits `wiyata:context-changed`, resets active class state, and returns the landing route.

`emailVerified`/`emailVerifiedAt` on the auth store come exclusively from `GET /me/context`. Never assign them directly — `LoginResponseDTO`/`VerifyEmailResponseDTO` intentionally do not carry this field; always go through `refreshUserContext()`.

Role layouts key their child `RouterView` by `auth.contextVersion` so same-path switches like `/student/dashboard` remount and reload data.

Singleton state reset/refetch on `wiyata:context-changed` exists for:

- `useFeedUnreadCount`
- `useNotificationUnreadCount`
- `useChatRoomSummary`

## 13. Self-Service School Creation and Email Verification

The old School Registration Request / super admin approval flow has been removed (application layer only — the `school_registration_requests` table itself is still present in the database, pending a separate drop). Any authenticated user with a verified email can create a school directly:

- `POST /register` — creates the user, auto-logs in, and issues a verification token/email (best-effort; failure never blocks registration).
- `POST /verify-email` — consumes a single-use, hashed, expiring token and stamps `usr_email_verified_at`.
- `POST /me/resend-verification` — reissues a token for an unverified user.
- `POST /schools` — gated by `RequireVerifiedUser()` (not a role check). Creates the school, enrolls the caller as `SchoolUser`, and assigns the `admin` role, all in one DB transaction. No approval step, no admin-in-the-middle.

Frontend then calls `refreshUserContext()` followed by `switchContext()` before redirecting to the new school's dashboard, so the membership is guaranteed to be visible before navigation.

`POST /super-admin/school-bootstrap` (manual provisioning by a super admin, including creating the first admin user) is unrelated to this flow and remains unchanged.

## 14. Member Invitation, Direct-Create, and Import

School admin can invite teacher/student by email:

- `POST /api/admin/school-member-invitations`
- `GET /api/admin/school-member-invitations`
- `PATCH /api/admin/school-member-invitations/:id/revoke`

Student invitations include class placement metadata. Enrollment is created when the invitation is accepted, not when it is created.

Direct-create/import remains available as fallback:

- manual direct-create,
- CSV import,
- XLSX template/upload parsed in browser and converted to existing import flow.

Direct-create/import now distinguishes created vs reused user metadata and sends account-created or added-to-school email best-effort without sending passwords.

**Phase 8 — existing-user invitation accept.** The admin-facing create form (`AdminUsers.vue`) no longer asks for a name — only email, role, and class (if student). `fullName` on `Invitation` is optional and, when omitted, is now stored as `nil` rather than a pointer to an empty string. It is a legacy/display-only field: no accept path (new-user or existing-user) reads it. Its only remaining consumer is `AdminDashboard.vue`'s pending-invitations widget, which falls back to `invitation.email` when it's empty.

`GET /api/invitations/:token` (public metadata) now returns `existingUser: bool` — whether the invited email already belongs to a `User`, computed via the same `UserRepository.CheckEmailExists` the public register flow already uses. This drives which of three states `AcceptInvitation.vue` renders:

1. New email → the original name/password/confirm form, posting to `POST /api/invitations/:token/accept` (public, unchanged).
2. Existing email, not logged in → an info card + "Login untuk menerima undangan" button, which navigates to `{ name: "login", query: { redirect: route.fullPath } }` — the same redirect mechanism the router guard already uses for every `requiresAuth` route, and that `LoginPage.vue` already reads on success. No new redirect framework was added.
3. Existing email, authenticated as that email → a "Terima Undangan" button posting to `POST /api/invitations/:token/accept-authenticated` — **requires a valid JWT** (registered after `api.Use(middleware.AuthRequired())` in `main.go`, reusing that middleware as-is). No name/password fields. The backend independently verifies `authenticated user's email == invitation.email` before accepting (`invitation_repo.AcceptAuthenticated`) — a caller authenticated as a *different* account is rejected with 403, not silently trusted from the frontend. If authenticated as a different email, the UI instead shows a mismatch message with a logout-and-retry action, since the router guard would otherwise bounce an already-authenticated visitor away from `/login` before it renders.

The public `Accept` endpoint and its DTO are untouched by this — the new-user registration path is exactly what it was before Phase 8. Full contract: `backend/docs/api/invitation.md` (public accept flow) and `backend/docs/api/school_member_invitations.md` (admin create/list/revoke).

**Phase 9.3 — school-role combination validation now covers invitation accept too.** `finalizeInvitationAcceptance` (the private helper both accept endpoints share) validates the accepting membership's resulting role set via `domain.ValidateSchoolRoleCombination` before assigning the invitation's role — the same shared validator used by the RBAC role editor and CSV import/direct member creation (see §24 and §26). An existing user who already holds `student` at a school is rejected (400) if they try to accept a `teacher`/`admin` invitation to that same school, instead of silently ending up with both roles.

## 15. Email Behavior and Security

SMTP env keys are documented in `.env.example` and docs. Do not include actual env values in documentation.

AI material summary env keys are backend-only:

- `AI_SUMMARY_ENABLED`
- `AI_SUMMARY_PROVIDER` (`openai` for OpenAI-compatible chat completions, `gemini` for Gemini REST)
- `AI_SUMMARY_API_KEY`
- `AI_SUMMARY_BASE_URL`
- `AI_SUMMARY_MODEL`
- `AI_SUMMARY_TIMEOUT_SECONDS`

Security invariants:

- Passwords are never sent by email.
- Existing user passwords are never overwritten when a user is reused.
- Invitation tokens are generated securely and stored only as hashes.
- Raw invitation tokens are returned only once for dev/manual fallback.
- Email failures are best-effort after successful DB work.
- Do not log SMTP passwords or raw invitation tokens.
- AI provider API keys stay on the backend. Do not send provider keys to the browser.
- AI prompt text, extracted document content, provider API keys, and raw provider responses must not be logged.
- Document contents are untrusted data. The summary prompt must treat them as data to summarize, not instructions to follow.

## 16. Academic Years, Terms, Classes, and Enrollments

Academic setup is school-scoped. Admins manage years, terms, classes, subjects, enrollments, and subject-class assignments.

`edv.enrollments.enr_role` remains `student|teacher` and is essential downstream. It controls access, teacher assignment validation, student workspace context, and class membership queries.

Current AdminEnrollments behavior:

- no global class-role dropdown;
- frontend infers placement role from selected school member roles;
- admin-only/no eligible role members are not selectable;
- mixed student/teacher selections are split into role-based API requests.

Caveat: backend still receives and writes the `role` payload. It does not yet authoritatively derive enrollment role from school-level roles.

## 17. Subject Classes and Teacher Workspace

`subject_classes` connect class + subject + teacher.

Teacher dashboard/workspace depends on subject-class assignment. Before assigning a teacher to a subject class, that teacher must have:

- school role `teacher`;
- active teacher placement in the target class.

Admin UI label is "Penugasan Mengajar".

## 18. Student Learning Flows

Student dashboard uses active class context and right rail awareness:

- My Day / Hari Ini.
- assignment preview.
- notification/chat/feed tabs.
- mini calendar.

Academic Activity `date` is date-only `YYYY-MM-DD`. Do not parse it as timestamp. Calendar dots must remain assignment_due/deadline-only.

## 19. Materials, Assignments, Submissions, and Assessment

Materials and assignments live under subject class.

Material progress exists. Opening material can update progress signals.

Material PDF attachments can be summarized through:

- `POST /api/materials/:materialId/media/:mediaId/summary`

This MVP summarizes one attached PDF at a time from the file contents, not from `mat_desc`. The backend verifies material access, verifies the `mediaId` is attached to the material, resolves `material -> attachment -> media -> storagePath`, downloads internally from storage, extracts text from text-based PDF only, and calls the configured AI provider. It does not accept arbitrary file URLs from the frontend.

AI summary MVP limitations:

- PDF text layer only.
- No OCR.
- No DOCX/TXT/PPT/PPTX support yet.
- No database persistence, cache, queue, or worker.
- Provider switch supports `AI_SUMMARY_PROVIDER=openai` and `AI_SUMMARY_PROVIDER=gemini`.

Student and teacher material detail pages show a "Rangkum dokumen" action on PDF attachments. Protected user-facing media delivery remains separate and deferred; internal backend storage download for summary does not make public media delivery protected.

Assignment submit flow is not fully optimistic. After upload and submit success, frontend can patch safe local status but backend remains source of truth.

Teacher assessment flow waits for backend success, then patches local submission state without blocking full reload where possible.

Timestamps are now `timestamptz` in DB and API responses should be RFC3339 for timestamp fields. Academic Activity `date` remains date-only.

## 20. Feed, Comments, Notifications, and Chat

Feed posts and comments exist. Feed create supports canonical response and teacher optimistic placeholder UI.

Comments/discussions are polymorphic for:

- feed,
- material,
- assignment.

`DiscussionComments.vue` is generic for material/assignment discussions. `FeedComments.vue` remains feed-specific.

Notification Center exists for student and teacher. Notification unread state is REST/refresh based. There is no general notification WebSocket/SSE.

Chat has DM/group/class-style room flows, unread summary, polling fallback, and WebSocket behavior. Chat realtime is separate from notifications.

## 21. Important Backend Endpoints

Representative endpoints: `POST /api/login`, `POST /api/register`, `POST /api/verify-email`, `GET /api/me/context`, `POST /api/me/resend-verification`, `POST /api/schools` (self-service Create School, see section 13), `GET /api/invitations/:token`, `POST /api/invitations/:token/accept`, `/api/admin/school-member-invitations`, `/api/admin/school-members`, `/api/admin/school-members/import/preview`, `/api/admin/school-members/import/commit`, `/api/enrollments`, `/api/subject-classes/assign`, `/api/subject-classes/my-teaching`, `/api/materials`, `POST /api/materials/:materialId/media/:mediaId/summary`, `/api/assignments`, `/api/comments`, `/api/feeds`, `/api/notifications`, `/api/chat/rooms`, `/api/ws/chat`.

Use specialized docs in `backend/docs/api/` and route registration in `backend/cmd/api/main.go` for exact contracts.

## 22. Important Frontend Routes

Representative routes: `/login`, `/register`, `/verify-email`, `/onboarding`, `/create-school`, `/invite/:token`, `/superadmin/dashboard`, `/admin/dashboard`, `/admin/users`, `/admin/enrollments`, `/admin/subject-classes`, `/teacher/dashboard`, `/teacher/subjects/:subjectClassId`, `/teacher/subjects/:subjectClassId/assignments/:assignmentId`, `/teacher/notifications`, `/student/dashboard`, `/student/subjects/:subjectClassId/materials/:materialId`, `/student/subjects/:subjectClassId/assignments/:assignmentId`, `/student/notifications`. (`/school-registration` still exists but is unlinked from navigation — see section 13; `/superadmin/school-registration-requests` was removed entirely.)

Check `frontend/src/router/index.ts` before adding or linking routes.

## 23. Important Stores and Composables

- `frontend/src/stores/auth.ts` - identity, memberships, active context, context reconciliation.
- `frontend/src/stores/activeClass.ts` - student active class context.
- `frontend/src/services/api.ts` - request headers and API client.
- `frontend/src/components/layout/ContextSwitcher.vue` - visual school/role switcher.
- `frontend/src/composables/useFeedUnreadCount.ts` - feed unread singleton.
- `frontend/src/composables/useNotificationUnreadCount.ts` - notification unread singleton.
- `frontend/src/composables/useChatRoomSummary.ts` - shared chat room summary and unread state.
- `frontend/src/components/discussion/DiscussionComments.vue` - material/assignment discussion.
- `frontend/src/components/feed/FeedComments.vue` - feed comments only.

## 24. Product Decisions and Security Invariants

- Current code and tests override stale documentation.
- School is tenant root.
- Global user identity is separate from school membership.
- One active school role is used at runtime.
- Platform super_admin is not a school context.
- Passwords are never emailed.
- Invitation tokens are stored hashed.
- Existing passwords are not overwritten during reused-user flows.
- Email and notification failures are usually best-effort.
- Assignment deadline is an instant; teacher form sends Jakarta offset for MVP.
- Academic Activity `date` remains date-only.
- Enrollment `enr_role` remains essential until backend-authoritative derivation is implemented.
- School-role combination is restricted: `admin`+`teacher` is the only combination allowed on one `school_users` membership; `student` can never be combined with `teacher` or `admin`. Enforced backend-side as the single source of truth (`domain.ValidateSchoolRoleCombination`, `backend/internal/domain/role_validation.go`) at every mutation path — role sync editor and single-role assign (`rbacService`), CSV import and direct member creation (`adminSchoolMemberImportService`), and invitation accept, both the public new-user path and the authenticated existing-user path (`invitationRepository`) — not just in the frontend. `super_admin` is out of scope for this rule: it is a platform role never assigned through any of these flows. See `backend/docs/api/rbac.md` §2.

## 25. Dashboards (Phase 7 — School Admin & Super Admin)

Phase 7 (Batches 1–4) extended the School Admin and Super Admin dashboards using only the existing bundled endpoints — `GET /dashboard/admin/:schoolId` and `GET /dashboard/super-admin`. No new endpoints, no new tables/migrations, no chart dependency were introduced. Student and Teacher dashboards were explicitly out of scope for this initiative and were not touched. Full response contracts: `backend/docs/api/dashboard.md`.

### School Admin Dashboard (`frontend/src/pages/admin/AdminDashboard.vue`)

Current section order, top to bottom:

1. **Needs Attention** — persistent shell; alerts for incomplete academic setup, pending invitations, and classes without a teacher. Shows a positive "Semua beres" state when nothing needs attention, or a neutral "data belum lengkap" state (not a false positive) when an underlying fetch failed.
2. **Work Queue** — five widgets, each its own persistent-shell card:
   - Undangan Tertunda (pending invitations)
   - Kelas Tanpa Guru (classes without teacher)
   - Subject-Class Tanpa Konten (subject-classes without materials/assignments)
   - Mata Pelajaran Belum Dikonfigurasi (subjects without assessment-weight configuration)
   - Antrean Penilaian (grading backlog — summary only, no drill-down; Admin has no grading UI of its own)
3. **Overview** — 4 stat tiles (total students/teachers/classes/active classes)
4. **Recent Activity** (Aktivitas Terbaru)
5. **Sidebar** — chat panel, then Distribusi Kelas (enrollment distribution bars)

An earlier "Setup Progress" section and a "School Performance Rollup" table were built during this initiative but are **not present in the current file** — see §27 (Known Technical Debt) for exact status. This list reflects what is actually on disk today.

### Super Admin Dashboard (`frontend/src/pages/superadmin/SuperAdminDashboard.vue`)

Current section order, top to bottom:

1. **Needs Attention** — schools without an admin, schools without academic setup; same persistent-shell + positive-empty-state pattern as Admin.
2. **Work Queue** — Sekolah Tanpa Admin, Sekolah Tanpa Setup Akademik.
3. **Overview** — 2 stat tiles (total schools, total platform users) + Sekolah Terbaru (recently created schools).
4. **Platform Trends** — School Growth and User Growth (see below).
5. **Reference** — static descriptive cards (`overviewCards`) and the "Alur pengaturan awal tenant" onboarding-flow article. Not data-driven, unchanged by Phase 7.
6. **Sidebar** — Aksi cepat (Quick Actions).

**Platform Trends widgets:** both read `schoolGrowthTrend` / `userGrowthTrend` from the same bundled `/dashboard/super-admin` response — 6 monthly points each, oldest → newest. They render as **plain HTML/Tailwind bars** (one `<div>` per month, inline `height: %` style, inside a fixed `h-[180px]` container, colored with theme tokens `bg-brand`/`bg-info`) — no charting library. See the architectural decisions below for why.

### Backend repository methods added (Batches 1–4)

All in `internal/repository/dashboard_repo.go`, called from `dashboard_service.GetAdminDashboard` / `GetSuperAdminDashboard`:

- `GetClassesWithoutTeacher` — active classes with zero `subject_classes` rows.
- `GetContentLessSubjectClasses` — subject-classes with neither materials nor assignments (reuses `SubjectClassRepository.HasSubjectClassContent`'s definition, not a new rule).
- `GetSubjectsWithoutAssessmentWeight` — subjects with zero `assessments_weights` rows (safe because `grade_service.ConfigureWeights` only ever persists a set summing to exactly 100 — never partial).
- `GetGradingBacklog` — total ungraded submissions school-wide + top 3 classes by backlog count (reuses the "no matching `assessments` row" definition from `GetPendingReviewsCount`).
- `GetSchoolPerformanceRollup` — weakest 5 subject-classes by average score (reuses `GetClassPerformance`'s formula); **implemented but not currently called** — see §27.
- `GetSchoolsWithoutAdmin` — active schools with zero `school_users` holding the `admin` role.
- `GetSchoolsWithoutSetup` — active schools with no active `academic_years` row.
- `GetSchoolGrowthTrend` / `GetUserGrowthTrend` — 6-month `generate_series`-based counts of schools/users created per month, oldest → newest.

### Architectural decisions (Phase 7)

- Every widget renders a persistent shell across loading/empty/error states — no widget fully disappears from the page.
- Empty states are positive/neutral ("Semua beres", "Tidak ada...", "Belum ada...") rather than blank.
- Widget-level error messages are shown inline per widget rather than one generic page-level failure; widgets fed by the same bundled response fail together (see §27), but each still surfaces its own inline text.
- Loading uses `animate-pulse` skeletons matching each widget's eventual shape (row skeletons, stat-tile skeletons, chart-area skeletons) — no new animation/shimmer style was introduced.
- No chart dependency was introduced. Two 6-point trend charts were judged too simple to justify a charting library — no interactivity/tooltips/axes/zoom required, fixed height, no gradients/3D/shadows/animation per the design brief — so they reuse the hand-rolled proportional-bar idiom `AdminDashboard.vue`'s "Distribusi Kelas" widget already established, extended to a time axis.

## 26. Recently Completed Work

- Active school + active role backend/frontend foundation, visual ContextSwitcher, `Active-Role` CORS support, and keyed route remounting.
- Self-service school creation (`POST /schools`, gated by email verification — creator becomes Admin atomically), admin invitation, public invitation accept, teacher/student member invitations, and best-effort emails. The old School Registration Request / super admin approval flow has been removed from the application layer (Phase 4A); the `school_registration_requests` table itself is still present in the database, pending a separate drop.
- Phase 4B: fixed `GET /academic-years/school/:schoolCode` and `GET /terms/academic-year/:id` returning `data: null` (nil Go slice) instead of `[]` for a brand-new school with zero academic years — this was crashing Admin Dashboard's `loadDashboard()` immediately after self-service Create School. Dashboard now loads correctly with a friendly empty state for schools with no academic year/term/class/student yet.
- Phase 4C/4D: fixed the email-verification banner staying visible after successful verification. `VerifyEmail.vue` now calls `refreshUserContext()` when a session is active, and `useAuthContextSync()` refreshes auth context on tab focus/visibility (throttled) so verification done in another tab/device is picked up without a manual reload.
- Notification Center, material/assignment discussions, Teacher Assignment Detail, AdminEnrollments frontend role inference, and `timestamptz`/RFC3339 timestamp migration.
- **Hardening Phase 1 — Authorization:** Added `RequireSchoolMember + RequireRole` to previously unprotected endpoints (`GET /assignments/status/:id`, `GET /grades/class/:classId/subject/:subjectId`). Added handler-level ownership checks (`resource.SchoolID == activeSchoolID`) to academic year, term, subject, class, dashboard, and log mutation/read endpoints.
- **Hardening Phase 2 — Assessment Weight Transaction:** `ConfigureWeights` now uses `ReplaceBySubject()` — a single atomic DB transaction — instead of separate delete + create calls.
- **Hardening Phase 3 — Resource Ownership Audit:** Dashboard endpoints now require school membership + matching role; params are validated against JWT context. Log endpoint restricted to admin/super_admin of the active school.
- **Hardening Phase 4 — Async Consistency:** Stale response guards added to `ChatWorkspace.vue` (room switch) and `TeacherFeed.vue` (class switch). Pattern: capture resource ID before await, discard response if ID changed.
- **Hardening Phase 5 — Error Handling:** Created `frontend/src/utils/error.ts` with `getApiError(error: unknown)`. Removed 11 duplicate local error helpers; all replaced with shared utility.
- **Hardening Phase 6 — Backend Unit Tests:** Added `grade_service_test.go` (10 tests) and `assignment_service_test.go` (6 tests) covering weight validation, duplicate category, atomic replace, deadline enforcement, and submission integrity.
- **Phase 7 (Batches 1–4) — School Admin & Super Admin Dashboards:** Extended the existing bundled dashboard endpoints with work-queue widgets (classes without teacher, content-less subject-classes, subjects missing assessment-weight config, grading backlog), a school-wide performance rollup, super-admin work-queue widgets (schools without admin/setup), and two platform growth-trend charts (school growth, user growth) rendered with plain HTML/Tailwind bars — no chart library added. See §25 for full detail.
- **Phase 8 (Batches 1–3 + follow-up) — RBAC and Invitation Correctness:** (1) Fixed `RequireSystemSuperAdmin` middleware, which always returned 403 because it resolved a "system school" by code `000000` that nothing in the app ever creates — now checks `rbacRepo.IsSuperAdmin` directly, and 9 super-admin-only routes were moved onto it from the broken `RequireRole(schoolService, "super_admin")`-alone pattern. (2) Replaced `AdminUsers.vue`'s single-role dropdown with a multi-checkbox editor, closing a real data-loss bug where saving a multi-role member's role silently dropped their other roles. (3) Added an authenticated existing-user invitation-accept path (`POST /invitations/:token/accept-authenticated`, JWT-gated, email-matched server-side) alongside the unchanged public registration accept endpoint, plus made the admin invitation-creation form's `fullName` field optional (no accept path ever read it). See §14 for the invitation flow detail and `backend/docs/api/invitation.md` / `school_member_invitations.md` for the API contracts.
- **Phase 9 — Multi-Role Evaluation and School-Role Combination Validation:** (1) *9.1 (read-only audit)* concluded invitation and enrollment should stay single-role-per-action: the `enrollments` table's `(enr_scu_id, enr_cls_id)` unique constraint (no `enr_role` column) makes per-class dual-role structurally impossible regardless of invitation design, and teacher invitations never auto-create enrollment anyway, so multi-role invitations wouldn't remove the always-necessary follow-up admin step. This closes the "evaluate multi-role invitation/enrollment" TODO with **no implementation** (Option A). (2) *9.2 (read-only audit)* traced every remaining single-role-shaped pattern across the frontend and found exactly one real bug: `AdminEnrollments.vue`'s `inferPlacementRole()` silently returns `null` (blocking enrollment UI) for a member holding both `student` and `teacher` — a state only reachable since Batch 2's multi-role editor. (3) *9.3 (implementation)* resolved that bug **at its source instead of patching `AdminEnrollments.vue`**: after a business-rule decision that `student`+`teacher` and `student`+`admin` should never be valid combinations in the first place (only `admin`+`teacher` is), added `domain.ValidateSchoolRoleCombination` (`backend/internal/domain/role_validation.go`) as the single backend source of truth, called from every mutation path that can change a school membership's role set — `rbacService.SyncUserRoles`/`AssignRoleToUser`, `adminSchoolMemberImportService`'s CSV commit and direct member creation, and `invitationRepository.finalizeInvitationAcceptance` (both accept endpoints) — plus a matching frontend inline-validation guard in `AdminUsers.vue`'s role editor. See §24 for the rule statement and `backend/docs/api/rbac.md` §2 for the full contract.

## 27. Known Technical Debt and Edge Cases

- Backend still accepts enrollment `role` payload and does not authoritatively derive it from school roles.
- Some historical docs remain stale or analysis-only.
- Notification realtime is not implemented.
- Signed/private file delivery and thumbnails are unfinished.
- Multi-role/multi-school QA should continue after context switcher changes.
- Page-local in-flight requests are usually handled by route remount, not a global cancellation manager; `TeacherFeed.vue` and `ChatWorkspace.vue` now also have explicit stale guards.
- Some frontend build warnings may be non-blocking but should be rechecked in current output.
- `POST /academic-years`, `POST /terms`, `POST /subjects`, `POST /classes` accept `schoolId` in request body without validating it against the caller's active school — known LOW risk (admin role required, create-only).
- gofmt non-compliance across most Go source files.
- **Dashboard endpoints are bundled per role** (`/dashboard/admin/:schoolId`, `/dashboard/super-admin` each return everything that role's page needs in one response). A failure of that one call fails every widget on the page together, though each widget still shows its own inline error text rather than a generic crash.
- **Several Phase 7 widgets link to the nearest existing page**, not a dedicated one, since no dedicated page exists yet: pending invitations → `/admin/users`; classes-without-teacher → `/admin/classes`; content-less subject-classes and subjects-without-weight-config → `/admin/subject-classes`; schools-without-admin/setup → `/superadmin/schools`.
- **Grading Backlog (Admin Dashboard) is monitoring-only** — no drill-down link, since grading is a teacher-role feature Admin has no UI for.
- **Platform Trends charts (Super Admin) are non-interactive** — no tooltips, hover values, drill-down, export, or filters, by design.
- **`GetSchoolPerformanceRollup` (repository method, `internal/repository/dashboard_repo.go`) and `SchoolPerformanceRollup` (DTO field on `AdminDashboardDTO`) are dead code as of this writing.** `dashboard_service.GetAdminDashboard` does not call the method, and `AdminDashboard.vue` has no template section rendering the field — it always serializes as `null`. A future pass should either finish wiring it (add the service call + a rendering section) or remove the orphaned method/field.
- **An earlier "Setup Progress" section is no longer present in `AdminDashboard.vue`.** The section, its `setupSteps`/`isSetupComplete` script logic, and its `PhCheckCircle`/`PhCircleDashed` icon usage were removed at some point during this initiative and were not restored. Confirm with the team whether this was intentional before assuming it should come back.
- **School-member invitations still carry a single `role` per invitation**, and `CreateSchoolMemberInvitationDTO` has no array field — this is now a confirmed **design decision, not open debt** (Phase 9.1): the `enrollments` table's unique constraint has no `enr_role` column, making per-class dual-role structurally impossible at the database level regardless of invitation design, so multi-role invitations were rejected (Option A, no implementation). See §26.
- **The frontend single-role-assumption audit (Phase 9.2) is now complete**, not scoped to `AdminUsers.vue` only. It found one real bug (`AdminEnrollments.vue`'s `inferPlacementRole()` returning `null` for `student`+`teacher` members) which Phase 9.3 fixed at the source via backend+frontend role-combination validation rather than patching that page — see §24 and §26. No other page-level single-role bugs were found; `ReadProfile.vue` and `AdminSubjectClasses.vue` were confirmed as already-correct multi-role reference patterns.

## 28. Current Open Work

- Assignment extension request/review flow; protected media download URLs and thumbnails; grade/transcript export; notification preferences and optional realtime notification delivery; rich text and sanitization; nested comments if product decides; backend-authoritative enrollment role derivation.

## 29. Recommended Next Steps

1. Add backend validation/derivation for enrollment role based on school-level roles.
2. Add focused tests around context switching, invitation accept flow, and discussion notification recipients (see `backend/TODO.md` — Test Coverage Follow-Up).
3. Continue admin setup UX polish around prerequisites for subject-class assignment.
4. Implement protected file delivery before expanding media-heavy workflows.
5. Apply gofmt to all Go source files.

## 30. Validation Commands

Backend:

```bash
cd backend
GOCACHE=/private/tmp/wiyata-go-build-cache go test ./...
GOCACHE=/private/tmp/wiyata-go-build-cache go build ./...
```

Frontend:

```bash
cd frontend
npm run build
```

Repo hygiene:

```bash
git diff --check
git status --short
```

## 31. Known Non-Blocking Warnings

Historical local runs have surfaced non-blocking warnings such as CSS `@import` order or large Vite chunk warnings. Treat current command output as authoritative; do not assume old warnings still apply.

Shell startup warnings from a developer's local profile are environment issues, not project validation failures.

## 32. AI Development Workflow

- Read the code path before proposing a fix.
- For frontend-to-backend behavior, trace page/component → service → API route → handler → service → repository.
- Respect explicit read-only/audit requests.
- Keep patches scoped to the requested stage.
- Prefer existing patterns over new architecture.
- Run the requested validation commands.
- Report honestly when validation is build-only and not runtime QA.

## 33. Git Safety Rules

- The worktree may be dirty.
- Never revert changes you did not make unless explicitly asked.
- Do not use destructive commands such as `git reset --hard` or `git checkout --` without explicit approval.
- Do not delete, rename, or archive docs unless the user explicitly asks.
- For documentation-only tasks, do not modify runtime source, config, migrations, or tests.

## 34. Detailed Documentation Index

Specialized docs to consult: `README.md`, `README_EN.md`, `TODO.md`, `backend/TODO.md`, `backend/schema.md`, `backend/docs/API_SUMMARY.md`, `backend/docs/api/enrollment.md`, `backend/docs/api/notification.md`, `backend/docs/api/dashboard.md` (student/teacher/admin/super-admin dashboard contracts, including Phase 7 fields — see §25), `backend/docs/api/invitation.md` (public invitation accept flow, including the Phase 8 existing-user path — see §14), `backend/docs/api/school_member_invitations.md` (admin-facing invitation create/list/revoke), `backend/docs/api/rbac.md` (role/user-role management, including the Phase 9.3 school-role combination rule — see §24), `backend/docs/api/school_member_import.md` (direct member creation/CSV import, same combination rule applies). Older analysis/reference docs in `docs/ANALYSIS_INDEX.md`, `docs/CODEBASE_ANALYSIS.md`, `docs/QUICK_REFERENCE.md`, and `docs/PRODUCT_SCOPE.md` must be verified against current code before relying on details. (`backend/docs/api/school_registration_requests.md`, which documented the removed flow, has been deleted — see section 13 for the current self-service Create School flow.)

## 35. Source-of-Truth Hierarchy

Use this hierarchy when facts conflict:

1. Current code and actual database schema.
2. Tests and migrations/manual schema changes.
3. `docs/AI_HANDOFF.md`.
4. Specialized API/schema documentation.
5. Historical planning/analysis documents.

Do not include secrets in documentation: SMTP passwords, JWTs, database credentials, `.env` values, raw invitation tokens, real user passwords, or private secret-bearing URLs.
