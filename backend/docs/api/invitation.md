# Invitation Accept API (Public)

Base URL: `/api/invitations`

These are the **public**, unauthenticated-by-default endpoints a person uses when they open an invitation link (`/invite/:token` on the frontend). They are separate from `backend/docs/api/school_member_invitations.md`, which documents the **admin-facing** creation/list/revoke endpoints. An invitation created via that API is accepted here.

Both `GET /invitations/:token` and `POST /invitations/:token/accept` are registered before the global `AuthRequired()` middleware in `cmd/api/main.go`, so they work with no `Authorization` header. `POST /invitations/:token/accept-authenticated` is registered after it and requires a valid JWT — see below.

## 1. Get Invitation Metadata

`GET /api/invitations/:token`

Public. Not rate-limited (token-gated: guessing is infeasible given the token's length and hashing).

Response:

```json
{
  "invitationId": "8fc22388-90ee-4123-aad4-138c5c51e8c8",
  "email": "budi@example.com",
  "role": "student",
  "school": {
    "schoolId": "...",
    "schoolCode": "ABC123",
    "schoolName": "SMA Contoh"
  },
  "expiresAt": "2026-07-09T05:00:00Z",
  "status": "valid",
  "existingUser": true
}
```

- `existingUser` (added for Phase 8's existing-user flow): `true` if `email` already belongs to a `User` row, `false` otherwise. Computed via `UserRepository.CheckEmailExists` — the same check `auth_service.Register` already uses to reject duplicate signups, not a new query. This is what the frontend uses to decide which of the three `AcceptInvitation.vue` states to render (new-user form / "login to accept" / "accept" button).
- Not user-enumeration-free, but narrower than `/register`, which already leaks the same fact publicly with no token required at all: this only reveals the fact about the one specific email an admin already chose when creating the invitation, and only to someone who possesses the token.

## 2. Accept — New User (Registration)

`POST /api/invitations/:token/accept`

Public. Used when `existingUser` is `false`. **Unchanged by Phase 8** — this is the original registration-style accept path.

Request:

```json
{
  "name": "Budi Santoso",
  "password": "minimal6karakter",
  "confirmPassword": "minimal6karakter"
}
```

`name`, `password`, and `confirmPassword` are all required (`password` min 6 characters, must match `confirmPassword`). This is where a brand-new user's `fullName` actually comes from — **not** from the `fullName` an admin optionally typed when creating the invitation (see `school_member_invitations.md` — that field is a separate, optional, display-only value).

If the email already belongs to an existing user, the submitted `name`/`password` are only applied to fields that are currently empty on that account (`resolveInvitationUser` in `invitation_repo.go`) — an existing password is never overwritten. In practice, existing users should use endpoint 3 instead; this fallback exists at the data layer for safety, not as the intended UX.

Response:

```json
{
  "message": "Invitation accepted",
  "user": { "userId": "...", "fullName": "Budi Santoso", "email": "budi@example.com" },
  "school": { "schoolId": "...", "schoolCode": "ABC123", "schoolName": "SMA Contoh" },
  "role": "student"
}
```

## 3. Accept — Existing User (Authenticated)

`POST /api/invitations/:token/accept-authenticated`

**Requires a valid JWT** (`Authorization: Bearer <token>`). Registered after `api.Use(middleware.AuthRequired())` in `cmd/api/main.go` — no new middleware was introduced, this reuses the same `AuthRequired()` every other protected route uses.

No request body. Identity comes entirely from the JWT (`middleware.GetUserID(c)`), not from any submitted field — there is no name/password form for this path.

**Server-side verification:** the handler loads the authenticated user by ID and checks `user.Email == invitation.Email` (case-insensitive) before accepting anything. A mismatch returns `403` with `"error": "Invitation email does not match the authenticated account"` — the caller being logged in as a *different* account is rejected, not silently ignored. This check happens in the backend (`invitation_repo.AcceptAuthenticated`); the frontend never being trusted to enforce it is intentional.

Response: identical shape to endpoint 2.

**School-role combination rule (both endpoints).** Before assigning the invitation's role, `finalizeInvitationAcceptance` (`invitation_repo.go`) fetches the accepting `school_users`' existing role names and validates the combined set via `domain.ValidateSchoolRoleCombination` — the same shared validator used by the `AdminUsers.vue` role editor, `POST/PATCH /rbac/user-roles*`, and CSV import/direct member creation. This matters specifically because an *existing* user can already hold a school role before accepting an invitation for a different one: e.g. a user who is already `student` at a school accepting a `teacher` invitation to the same school is rejected, not silently combined. `admin`+`teacher` remains the only school-role combination invitations can legally produce. See `backend/docs/api/rbac.md` §2 for the full rule.

Errors (same for both accept endpoints, via `handleInvitationError`):

| Condition | Status | Body |
|---|---|---|
| Invalid/expired/already-accepted/revoked token | 400 | `{"error": "Invitation is invalid or expired"}` |
| Invitation's stored class no longer exists | 400 | `{"error": "Invitation class is no longer available"}` |
| Authenticated as a different email than the invitation (endpoint 3 only) | 403 | `{"error": "Invitation email does not match the authenticated account"}` |
| No/invalid JWT (endpoint 3 only) | 401 | `{"error": "Unauthorized"}` |
| Accepting would combine `student` with `teacher`/`admin` on the same school membership | 400 | `{"error": "Kombinasi peran tidak diperbolehkan: ..."}` |

### Why two endpoints instead of one "smart" endpoint

Endpoints 2 and 3 share their entire post-identity-resolution logic (attach `SchoolUser`, assign role, enroll in class if applicable, mark the invitation accepted) via two private helpers in `invitation_repo.go` — `lockUsableInvitationWithSchool` and `finalizeInvitationAcceptance`. Neither endpoint calls the other; there's no endpoint "pretending" to be the other one. This was a deliberate choice over branching a single endpoint on whether an `Authorization` header happens to be present, which would have made one endpoint's behavior implicit and harder to reason about. Two endpoints, two clear contracts, one shared implementation.

## Frontend flow (`frontend/src/pages/public/AcceptInvitation.vue`)

1. New user (`existingUser: false`) → the original name/password/confirm form, unchanged, posts to endpoint 2.
2. Existing user, not logged in → an info card ("Email ini sudah terdaftar") and a **"Login untuk menerima undangan"** button. This navigates to `{ name: "login", query: { redirect: route.fullPath } }` — the exact same redirect mechanism the router guard already uses for every `requiresAuth` route; `LoginPage.vue` already reads `route.query.redirect` on success. No new redirect framework was introduced.
3. After login, the user lands back on the same invitation URL, metadata reloads, and — if the now-authenticated account's email matches the invitation — a **"Terima Undangan"** button appears and posts to endpoint 3 (no password field, ever).
4. Authenticated as a *different* email than the invitation: a distinct message with a "Keluar dan masuk dengan akun lain" action (`auth.logout()` then the same login redirect) — sending an already-authenticated user straight to `/login` would silently bounce them back to their own dashboard (the router guard redirects an authenticated visitor away from `/login`), so logout has to happen first.
