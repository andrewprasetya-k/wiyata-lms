# Wiyata AI Handoff

Last verified against codebase: 2026-07-05.

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
- onboarding `school_registration_requests` and `invitations`.
## 8. Multi-Tenant and Global Identity Model

User accounts are global by email. A user can belong to multiple schools through `school_users`.

Roles are not global for ordinary school users. They attach to school memberships. A user can have different roles across schools and may have multiple roles in one school according to the data model.

`super_admin` is a platform context, not a school role context.

Soft-deleted school memberships must not appear in login/context responses. Active school membership filtering is part of auth correctness.
## 9. Authentication and Authorization

JWT identifies the global user. School-scoped APIs also require school membership context.

Important middleware concepts:

- `AuthRequired` validates JWT.
- `RequireSchoolMember` validates active membership in `SchoolId`.
- `RequireRole` authorizes the selected active role when `Active-Role` is present.
- `RequireSystemSuperAdmin` protects platform routes.

Frontend role/context values are never trusted alone. Backend validates `SchoolId` and `Active-Role` against live DB membership/roles.
## 10. Active School and Active Role Context

Frontend active context is a union:

```ts
type ActiveContext =
  | { type: "school"; schoolId: string; schoolUserId: string; role: "admin" | "teacher" | "student" }
  | { type: "platform"; role: "super_admin" }
```

Runtime behavior uses exactly one active school + one active school role. It no longer uses the union of all roles for route authorization.

`auth.activeRole` is one selected role. `auth.activeRoles` and `auth.allRoles` are compatibility surfaces and should not be used to grant access to role pages.
## 11. Request Context Headers

For school context, frontend sends:

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

- `ensureUserContext()` initializes context once per app session and waits for an in-flight context request.
- `refreshUserContext()` is a forced refresh.
- In-flight context request protection prevents concurrent route guards from racing.
- `switchContext(target)` validates target, persists it, increments `contextVersion`, emits `wiyata:context-changed`, resets active class state, and returns the landing route.

Role layouts key their child `RouterView` by `auth.contextVersion` so same-path switches like `/student/dashboard` remount and reload data.

Singleton state reset/refetch on `wiyata:context-changed` exists for:

- `useFeedUnreadCount`
- `useNotificationUnreadCount`
- `useChatRoomSummary`
## 13. School Registration and Approval

Public visitors submit:

- `POST /api/school-registration-requests`

Super admin can:

- list/detail requests,
- reject pending requests,
- approve pending requests.

Approval creates a school, creates an admin invitation, marks the request approved, and sends invitation email best-effort after the transaction.

The response still includes invitation token/link for manual fallback because email may be disabled or fail.
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
## 15. Email Behavior and Security

SMTP env keys are documented in `.env.example` and docs. Do not include actual env values in documentation.

Security invariants:

- Passwords are never sent by email.
- Existing user passwords are never overwritten when a user is reused.
- Invitation tokens are generated securely and stored only as hashes.
- Raw invitation tokens are returned only once for dev/manual fallback.
- Email failures are best-effort after successful DB work.
- Do not log SMTP passwords or raw invitation tokens.
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

Representative endpoints: `POST /api/login`, `GET /api/me/context`, `POST /api/school-registration-requests`, `/api/super-admin/school-registration-requests`, `GET /api/invitations/:token`, `POST /api/invitations/:token/accept`, `/api/admin/school-member-invitations`, `/api/admin/school-members`, `/api/admin/school-members/import/preview`, `/api/admin/school-members/import/commit`, `/api/enrollments`, `/api/subject-classes/assign`, `/api/subject-classes/my-teaching`, `/api/materials`, `/api/assignments`, `/api/comments`, `/api/feeds`, `/api/notifications`, `/api/chat/rooms`, `/api/ws/chat`.

Use specialized docs in `backend/docs/api/` and route registration in `backend/cmd/api/main.go` for exact contracts.
## 22. Important Frontend Routes

Representative routes: `/login`, `/school-registration`, `/invite/:token`, `/superadmin/dashboard`, `/super-admin/school-registration-requests`, `/admin/dashboard`, `/admin/users`, `/admin/enrollments`, `/admin/subject-classes`, `/teacher/dashboard`, `/teacher/subjects/:subjectClassId`, `/teacher/subjects/:subjectClassId/assignments/:assignmentId`, `/teacher/notifications`, `/student/dashboard`, `/student/subjects/:subjectClassId/materials/:materialId`, `/student/subjects/:subjectClassId/assignments/:assignmentId`, `/student/notifications`.

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
## 25. Recently Completed Work

- Active school + active role backend/frontend foundation, visual ContextSwitcher, `Active-Role` CORS support, and keyed route remounting.
- School registration, super admin approval/reject, admin invitation, public invitation accept, teacher/student member invitations, and best-effort emails.
- Notification Center, material/assignment discussions, Teacher Assignment Detail, AdminEnrollments frontend role inference, and `timestamptz`/RFC3339 timestamp migration.
## 26. Known Technical Debt and Edge Cases

- Backend still accepts enrollment `role` payload and does not authoritatively derive it from school roles.
- Some historical docs remain stale or analysis-only.
- Notification realtime is not implemented.
- Signed/private file delivery and thumbnails are unfinished.
- Multi-role/multi-school QA should continue after context switcher changes.
- Page-local in-flight requests are usually handled by route remount, not a global cancellation manager.
- Some frontend build warnings may be non-blocking but should be rechecked in current output.
## 27. Current Open Work

- Assignment extension request/review flow; protected media download URLs and thumbnails; grade/transcript export; notification preferences and optional realtime notification delivery; rich text and sanitization; nested comments if product decides; backend-authoritative enrollment role derivation.
## 28. Recommended Next Steps

1. Add backend validation/derivation for enrollment role based on school-level roles.
2. Add focused tests around context switching, invitation accept, and discussion notification recipients.
3. Continue admin setup UX polish around prerequisites for subject-class assignment.
4. Implement protected file delivery before expanding media-heavy workflows.
## 29. Validation Commands

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
## 30. Known Non-Blocking Warnings

Historical local runs have surfaced non-blocking warnings such as CSS `@import` order or large Vite chunk warnings. Treat current command output as authoritative; do not assume old warnings still apply.

Shell startup warnings from a developer's local profile are environment issues, not project validation failures.
## 31. AI Development Workflow

- Read the code path before proposing a fix.
- For frontend-to-backend behavior, trace page/component → service → API route → handler → service → repository.
- Respect explicit read-only/audit requests.
- Keep patches scoped to the requested stage.
- Prefer existing patterns over new architecture.
- Run the requested validation commands.
- Report honestly when validation is build-only and not runtime QA.
## 32. Git Safety Rules

- The worktree may be dirty.
- Never revert changes you did not make unless explicitly asked.
- Do not use destructive commands such as `git reset --hard` or `git checkout --` without explicit approval.
- Do not delete, rename, or archive docs unless the user explicitly asks.
- For documentation-only tasks, do not modify runtime source, config, migrations, or tests.
## 33. Detailed Documentation Index

Specialized docs to consult: `README.md`, `README_EN.md`, `TODO.md`, `backend/TODO.md`, `backend/schema.md`, `backend/docs/api/school_registration_requests.md`, `backend/docs/api/enrollment.md`, `backend/docs/api/notification.md`. Older analysis/reference docs in `docs/ANALYSIS_INDEX.md`, `docs/CODEBASE_ANALYSIS.md`, `docs/QUICK_REFERENCE.md`, and `docs/PRODUCT_SCOPE.md` must be verified against current code before relying on details.
## 34. Source-of-Truth Hierarchy

Use this hierarchy when facts conflict:

1. Current code and actual database schema.
2. Tests and migrations/manual schema changes.
3. `docs/AI_HANDOFF.md`.
4. Specialized API/schema documentation.
5. Historical planning/analysis documents.

Do not include secrets in documentation: SMTP passwords, JWTs, database credentials, `.env` values, raw invitation tokens, real user passwords, or private secret-bearing URLs.
