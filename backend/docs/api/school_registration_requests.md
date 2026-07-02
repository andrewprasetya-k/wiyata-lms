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
