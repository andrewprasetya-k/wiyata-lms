# School Member Invitations API

Base URL: `/api/admin/school-member-invitations`

These endpoints let an active-school admin invite teachers and students by email. They create invitation tokens only. They do not create users, school memberships, roles, or enrollments when the invitation is created.

When the public invitation accept endpoint is used:

- Teacher invitations create or reuse the user, create or restore school membership, and assign the teacher role.
- Student invitations create or reuse the user, create or restore school membership, assign the student role, and create or reactivate the stored class enrollment.
- Admin onboarding invitations remain compatible and do not create enrollment.
- If the stored student class no longer exists in the invitation school, accept fails and the invitation remains unaccepted.

Direct-create school member endpoints remain available as a fallback and are not changed by this API.

## Schema Update

Manual SQL required:

```sql
ALTER TABLE edv.invitations
  ADD COLUMN IF NOT EXISTS inv_full_name text,
  ADD COLUMN IF NOT EXISTS inv_class_id uuid REFERENCES edv.classes(cls_id);
```

## Create Invitation

`POST /api/admin/school-member-invitations`

Protected by active school membership and `admin` role.

Request:

```json
{
  "fullName": "Budi Santoso",
  "email": "budi@example.com",
  "role": "student",
  "classCode": "X-IPA-1"
}
```

Rules:

- `role` must be `student` or `teacher`.
- `fullName`, `email`, and `role` are required.
- `classCode` is required for `student`.
- `classCode` must belong to the active school.
- `classCode` is rejected for `teacher`.
- Existing direct-create/import flows are not affected.
- A pending invitation for the same active school, email, and role is rejected.
- The raw token is returned only once for manual fallback. Only the token hash is stored.
- After the invitation row is created, Wiyata sends an invitation email best-effort when SMTP is configured.
- SMTP disabled or email delivery failure does not fail the invitation; use `acceptUrl` or `token` as the manual fallback.

Response:

```json
{
  "message": "School member invitation created",
  "invitation": {
    "invitationId": "8fc22388-90ee-4123-aad4-138c5c51e8c8",
    "fullName": "Budi Santoso",
    "email": "budi@example.com",
    "role": "student",
    "class": {
      "classId": "2b747f6f-1c57-4eb8-8b59-03b11a744463",
      "classCode": "X-IPA-1",
      "classTitle": "X IPA 1"
    },
    "status": "pending",
    "expiresAt": "2026-07-09T05:00:00Z",
    "createdAt": "2026-07-02T05:00:00Z"
  },
  "acceptUrl": "/invite/<rawToken>",
  "token": "<rawToken>"
}
```

Email behavior:

- Teacher email copy says the recipient is invited as `Guru`.
- Student email copy says the recipient is invited as `Siswa`.
- The email includes the full accept URL built from `APP_PUBLIC_URL` and `/invite/:token`.
- Email failures are logged without SMTP secrets or raw tokens.

## List Invitations

`GET /api/admin/school-member-invitations?status=pending&page=1&limit=20`

Status values:

- `pending`
- `accepted`
- `revoked`
- `expired`

Default status is `pending`.

Response:

```json
{
  "data": [],
  "totalItems": 0,
  "page": 1,
  "limit": 20,
  "totalPages": 0
}
```

## Revoke Invitation

`PATCH /api/admin/school-member-invitations/:id/revoke`

Only pending, unexpired invitations can be revoked. Accepted, revoked, or expired invitations return a conflict.

Response:

```json
{
  "message": "School member invitation revoked",
  "invitation": {
    "invitationId": "8fc22388-90ee-4123-aad4-138c5c51e8c8",
    "fullName": "Budi Santoso",
    "email": "budi@example.com",
    "role": "student",
    "status": "revoked",
    "expiresAt": "2026-07-09T05:00:00Z",
    "revokedAt": "2026-07-02T06:00:00Z",
    "createdAt": "2026-07-02T05:00:00Z"
  }
}
```
