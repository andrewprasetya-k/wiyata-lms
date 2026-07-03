# School Registration Requests API

Public visitors can submit a school registration request for later super admin review. This endpoint only creates a pending request; school creation, approval/rejection, invitation, and email delivery are handled by later onboarding steps.

## Submit Request

`POST /api/school-registration-requests`

Authentication: not required.

### Request

```json
{
  "schoolName": "SMA Wiyata Mandala",
  "npsn": "12345678",
  "picName": "Budi Santoso",
  "picEmail": "budi@example.com",
  "picPhone": "081234567890",
  "picRole": "Kepala Sekolah",
  "message": "Kami ingin mencoba Wiyata untuk semester baru."
}
```

Required fields:

- `schoolName`
- `picName`
- `picEmail`

Optional string fields are trimmed. Empty optional strings are stored as `null`.

### Response

`201 Created`

```json
{
  "message": "School registration request submitted",
  "request": {
    "requestId": "b2d3c64f-5c8c-47c1-8a35-b71fd67ef15e",
    "schoolName": "SMA Wiyata Mandala",
    "picName": "Budi Santoso",
    "picEmail": "budi@example.com",
    "status": "pending",
    "createdAt": "2026-07-02T04:00:00Z"
  }
}
```

`createdAt` uses the standard Wiyata API timestamp policy: RFC3339 timezone-aware timestamp.

### Validation

- `schoolName` is required and limited to 150 characters.
- `picName` is required and limited to 150 characters.
- `picEmail` is required and must be email-like.
- `npsn` is optional and limited to 50 characters.
- `picPhone` is optional and limited to 50 characters.
- `picRole` is optional and limited to 100 characters.
- `message` is optional and limited to 1000 characters.

### Duplicate Handling

If a pending request already exists with the same normalized `picEmail` or `schoolName`, the API returns:

`409 Conflict`

```json
{
  "error": "A pending registration request already exists for this school or contact email"
}
```

Approved and rejected historical requests are not blocked in this foundation step.

## Super Admin Management

The following endpoints require JWT authentication and system `super_admin` role.

### List Requests

`GET /api/super-admin/school-registration-requests?status=pending&page=1&limit=10`

Query parameters:

- `status`: optional, one of `pending`, `approved`, or `rejected`. Defaults to `pending`.
- `page`: optional, defaults to `1`.
- `limit`: optional, defaults to `10`, maximum `100`.

Response:

```json
{
  "data": [
    {
      "requestId": "b2d3c64f-5c8c-47c1-8a35-b71fd67ef15e",
      "schoolName": "SMA Wiyata Mandala",
      "picName": "Budi Santoso",
      "picEmail": "budi@example.com",
      "status": "pending",
      "createdAt": "2026-07-02T04:00:00Z",
      "updatedAt": "2026-07-02T04:00:00Z"
    }
  ],
  "totalItems": 1,
  "page": 1,
  "limit": 10,
  "totalPages": 1
}
```

### Get Request Detail

`GET /api/super-admin/school-registration-requests/:id`

Response includes submitted optional fields and review metadata when available:

```json
{
  "requestId": "b2d3c64f-5c8c-47c1-8a35-b71fd67ef15e",
  "schoolName": "SMA Wiyata Mandala",
  "npsn": "12345678",
  "picName": "Budi Santoso",
  "picEmail": "budi@example.com",
  "picPhone": "081234567890",
  "picRole": "Kepala Sekolah",
  "message": "Kami ingin mencoba Wiyata untuk semester baru.",
  "status": "pending",
  "createdAt": "2026-07-02T04:00:00Z",
  "updatedAt": "2026-07-02T04:00:00Z"
}
```

### Reject Request

`PATCH /api/super-admin/school-registration-requests/:id/reject`

Only `pending` requests can be rejected.

Request body:

```json
{
  "reason": "Data sekolah belum lengkap."
}
```

`reason` is optional and limited to 1000 characters.

Response:

```json
{
  "message": "School registration request rejected",
  "request": {
    "requestId": "b2d3c64f-5c8c-47c1-8a35-b71fd67ef15e",
    "schoolName": "SMA Wiyata Mandala",
    "picName": "Budi Santoso",
    "picEmail": "budi@example.com",
    "status": "rejected",
    "reviewedBy": "8c80d272-51a5-47e5-9078-74118dc77b5d",
    "reviewedAt": "2026-07-02T05:00:00Z",
    "reviewNote": "Data sekolah belum lengkap.",
    "createdAt": "2026-07-02T04:00:00Z",
    "updatedAt": "2026-07-02T05:00:00Z"
  }
}
```

### Approve Request

`PATCH /api/super-admin/school-registration-requests/:id/approve`

Approving a request creates a school, creates an admin invitation, and marks the request as approved in one transaction. It then sends the admin invitation email best-effort when SMTP is configured. Email failure does not roll back approval, and the response still returns the invitation link/token for manual delivery.

Only `pending` requests can be approved.

Request body:

```json
{
  "schoolCode": "SMWM",
  "schoolName": "SMA Wiyata Mandala",
  "adminName": "Budi Santoso",
  "adminEmail": "budi@example.com",
  "note": "Approved"
}
```

Rules:

- `schoolCode` is required and must be unique.
- `schoolName` defaults to the submitted `schoolName` when empty or omitted.
- `adminName` defaults to the submitted `picName` when empty or omitted.
- `adminEmail` defaults to the submitted `picEmail` when empty or omitted.
- `note` is optional and limited to 1000 characters.

Response:

```json
{
  "message": "School registration request approved",
  "request": {
    "requestId": "b2d3c64f-5c8c-47c1-8a35-b71fd67ef15e",
    "schoolName": "SMA Wiyata Mandala",
    "picName": "Budi Santoso",
    "picEmail": "budi@example.com",
    "status": "approved",
    "reviewedBy": "8c80d272-51a5-47e5-9078-74118dc77b5d",
    "reviewedAt": "2026-07-02T05:00:00Z",
    "reviewNote": "Approved",
    "createdAt": "2026-07-02T04:00:00Z",
    "updatedAt": "2026-07-02T05:00:00Z"
  },
  "school": {
    "schoolId": "7d521362-fb37-4137-824f-948d8acb2f45",
    "schoolCode": "SMWM",
    "schoolName": "SMA Wiyata Mandala"
  },
  "invitation": {
    "invitationId": "f68b33f8-7fcb-4d06-9e6d-cbf0fa0e41b0",
    "email": "budi@example.com",
    "role": "admin",
    "expiresAt": "2026-07-09T05:00:00Z",
    "acceptUrl": "/invite/FmVZgNLLXioYVCw7gN3NqTB6O1C5rjyHfBH0BRwsgH0",
    "token": "FmVZgNLLXioYVCw7gN3NqTB6O1C5rjyHfBH0BRwsgH0"
  }
}
```

The raw invitation token is returned only once for development/testing and manual fallback. The database stores only `inv_token_hash`.

## Public Invitation Accept

Invitation links are public because users receiving invitations may not have accounts yet. The raw token is never stored in the database; incoming tokens are hashed and compared to `inv_token_hash`.

### Get Invitation Metadata

`GET /api/invitations/:token`

Returns safe metadata for a valid invitation token.

```json
{
  "invitationId": "f68b33f8-7fcb-4d06-9e6d-cbf0fa0e41b0",
  "email": "budi@example.com",
  "role": "admin",
  "school": {
    "schoolId": "7d521362-fb37-4137-824f-948d8acb2f45",
    "schoolCode": "SMWM",
    "schoolName": "SMA Wiyata Mandala"
  },
  "expiresAt": "2026-07-09T05:00:00Z",
  "status": "valid"
}
```

Invalid, expired, revoked, or already accepted tokens return a generic invalid/expired error.

### Accept Invitation

`POST /api/invitations/:token/accept`

```json
{
  "name": "Budi Santoso",
  "password": "Password123!",
  "confirmPassword": "Password123!"
}
```

Behavior:

- Validates the token hash.
- Rejects expired, revoked, accepted, invalid, or deleted-school invitations.
- Creates a new user when the invited email does not exist.
- Reuses an existing user when the invited email already exists.
- Does not overwrite an existing user's password.
- Sets the password only for a new user or an existing user with no password.
- Creates or restores the school membership.
- Assigns the invited role to the membership.
- For student school-member invitations with a stored class, creates or reactivates the student enrollment during accept.
- If the stored invitation class no longer exists in the invitation school, accept is rejected and the invitation remains unaccepted.
- Marks the invitation as accepted.

Response:

```json
{
  "message": "Invitation accepted",
  "user": {
    "userId": "b1a3952d-0852-48b9-af25-d5c38c242721",
    "fullName": "Budi Santoso",
    "email": "budi@example.com"
  },
  "school": {
    "schoolId": "7d521362-fb37-4137-824f-948d8acb2f45",
    "schoolCode": "SMWM",
    "schoolName": "SMA Wiyata Mandala"
  },
  "role": "admin"
}
```

Accepting an invitation is atomic. If user creation/reuse, membership creation/restore, role assignment, or accepted marking fails, the transaction rolls back.
