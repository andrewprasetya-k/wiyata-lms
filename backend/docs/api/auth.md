# 🔐 Authentication API Documentation

Base URL: `/api`

## 1. Register

Create a new plain global user account and receive JWT token.

- **URL:** `/register`
- **Method:** `POST`
- **Authentication:** Not required
- **Body:**

```json
{
  "fullName": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Validation:**

- `fullName`: Required
- `email`: Required, valid email format
- `password`: Required, minimum 6 characters
- Registration does not accept `schoolId`, `schoolCode`, role, enrollment, or class fields.
- Registration does not create `school_users`, assign roles, or grant school access.
  School access comes later either by creating a school directly (self-service, once the
  email is verified — see section 4) or by accepting an invitation from an existing school admin.
- Registration auto-logs in (this response is identical in shape to Login) and issues a
  best-effort email verification token/email (see section 4). A failure to send the
  verification email never blocks registration itself.

**Response (201 Created):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "fullName": "John Doe",
    "email": "john@example.com"
  },
  "memberships": [],
  "globalRoles": []
}
```

Newly registered users may not have school memberships or roles yet, so `memberships` can be empty.

**Error Responses:**

- `400 Bad Request`: Validation error
- `409 Conflict`: Email already registered

---

## 2. Login

Authenticate user and receive JWT token.

- **URL:** `/login`
- **Method:** `POST`
- **Authentication:** Not required
- **Body:**

```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "fullName": "John Doe",
    "email": "john@example.com"
  },
  "memberships": [
    {
      "schoolUserId": "uuid",
      "school": {
        "id": "uuid",
        "code": "SCH001",
        "name": "Wiyata Academy"
      },
      "roles": ["teacher"],
      "isDefault": true
    }
  ],
  "globalRoles": [],
  "defaultContext": {
    "schoolId": "uuid",
    "schoolUserId": "uuid",
    "roles": ["teacher"]
  }
}
```

`memberships` is the source for frontend role-based routing after login. A user can belong to multiple schools and can have multiple roles per school. If multiple memberships or roles are returned, the frontend should ask the user to choose the active school/role context.

**Error Responses:**

- `400 Bad Request`: Validation error
- `401 Unauthorized`: Invalid credentials

---

## 3. Verify Email

Consume a single-use, hashed, expiring email verification token and stamp `usr_email_verified_at`. Required before the user can create a school (see `backend/docs/api/school.md`, section 4).

- **URL:** `/verify-email`
- **Method:** `POST`
- **Authentication:** Not required (the token itself is the credential; the browser opening the emailed link may not be logged in)
- **Body:**

```json
{ "token": "raw-token-from-emailed-link" }
```

**Response (200 OK):**

```json
{
  "message": "Email verified",
  "emailVerifiedAt": "2026-07-17T09:00:00Z"
}
```

- Every failure (unknown token, already consumed, expired) returns the same generic `400` message — the endpoint never reveals which case occurred.
- Consuming a token invalidates every other outstanding token for that user.
- If the browser has an active session, the frontend calls `refreshUserContext()` right after success so `emailVerified` reflects the change without a manual reload.

**Error Responses:**

- `400 Bad Request`: `{"error": "Verification link is invalid or expired"}` (covers all failure cases)

---

## 4. Resend Verification

Reissue a verification token for the current user if not yet verified.

- **URL:** `/me/resend-verification`
- **Method:** `POST`
- **Authentication:** Required

**Response (200 OK):**

```json
{ "message": "Verification email sent" }
```

**Error Responses:**

- `400 Bad Request`: `{"error": "Email is already verified"}`

---

## 5. Refresh Auth Context

Return the authoritative school membership and role context for the current
authenticated user.

- **URL:** `/me/context`
- **Method:** `GET`
- **Authentication:** Required

**Response (200 OK):**

```json
{
  "memberships": [
    {
      "schoolUserId": "uuid",
      "school": {
        "id": "uuid",
        "code": "SCH001",
        "name": "Wiyata Academy"
      },
      "roles": ["teacher", "student"],
      "isDefault": true
    }
  ],
  "globalRoles": [],
  "defaultContext": {
    "schoolId": "uuid",
    "schoolUserId": "uuid",
    "roles": ["teacher", "student"]
  },
  "emailVerified": true,
  "emailVerifiedAt": "2026-07-17T09:00:00Z"
}
```

Soft-deleted `school_users` memberships are excluded. `defaultContext`, when
present, always points to an active membership. `emailVerified`/`emailVerifiedAt`
are the **only** source of truth the frontend uses for verification status —
neither Login nor Register responses carry this field; the frontend must call
this endpoint (via `refreshUserContext()`) to learn about a verification that
happened elsewhere.

---

## Active School and Role Headers

School-scoped endpoints continue to use:

```http
SchoolId: <school-id>
```

During the staged active-role rollout, clients may also send:

```http
Active-Role: admin|teacher|student
```

Behavior:

- If `Active-Role` is present, the backend validates that the role is assigned
  to the current user in the active school.
- `RequireRole` authorizes only the selected active role. It does not fall back
  to another role the user also owns.
- If `Active-Role` is absent, legacy behavior is preserved temporarily:
  `RequireRole` authorizes against any role the user owns in the active school.
- Unsupported active role values return `400 Bad Request`.
- Role not assigned in the active school returns `403 Forbidden`.
- Role assigned but not allowed by the route returns `403 Forbidden`.
- `super_admin` platform routes remain separate and do not use `Active-Role`.

## JWT Token Structure

**Claims:**

```json
{
  "user_id": "uuid",
  "sub": "uuid",
  "email": "john@example.com",
  "exp": 1234567890
}
```

**Expiry:** 24 hours from issue time

Roles are not embedded as the main JWT authority. Backend authorization checks role membership from the database using school context.

---

## Using Authentication

### 1. Get Token

Login or register to receive JWT token.

### 2. Include Token in Requests

Add `Authorization` header to all protected endpoints:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3. Example Request

```bash
curl -X GET http://localhost:8080/api/schools \
  -H "Authorization: Bearer eyJhbGc..."
```

---

## Protected Endpoints

**Public (No Auth):**

- `POST /api/login`
- `POST /api/register`
- `POST /api/verify-email`
- `GET /api/invitations/:token`
- `POST /api/invitations/:token/accept`

**Protected (Auth Required):**

- `POST /api/me/resend-verification`
- `POST /api/schools` — additionally requires a verified email (`RequireVerifiedUser()`), not a role
- All other `/api/schools/*` endpoints
- All `/api/users/*` endpoints
- All `/api/materials/*` endpoints
- All `/api/dashboard/*` endpoints
- All other API endpoints

---

## Error Responses

### 401 Unauthorized

**Missing Token:**

```json
{
  "error": "Unauthorized"
}
```

**Invalid Token Format:**

```json
{
  "error": "Invalid token format"
}
```

**Invalid/Expired Token:**

```json
{
  "error": "Invalid token"
}
```

---

## Security Notes

1. **Token Storage:**
   - Store token securely (localStorage/sessionStorage for web)
   - Never expose token in URL or logs

2. **Token Expiry:**
   - Token expires after 24 hours
   - User must login again after expiry
   - Implement refresh token for better UX (future enhancement)

3. **HTTPS:**
   - Always use HTTPS in production
   - Never send tokens over HTTP

4. **Secret Key:**
   - `JWT_SECRET` must be strong (min 32 characters)
   - Keep secret key secure in environment variables
   - Never commit secret key to version control

---

## Helper Functions (Backend)

For backend developers, use these helpers in handlers:

```go
import "backend/internal/middleware"

func (h *Handler) SomeEndpoint(c *gin.Context) {
    // Get authenticated user info from token
    userID := middleware.GetUserID(c)
    email := middleware.GetEmail(c)

    // Use in business logic
    data := h.service.GetByUser(userID)

    c.JSON(200, data)
}
```

---

## Testing with Postman

1. **Login:**
   - POST `/api/login`
   - Copy token from response

2. **Set Authorization:**
   - Go to Authorization tab
   - Type: Bearer Token
   - Token: Paste token from step 1

3. **Make Requests:**
   - All requests will include token automatically

---

**Last Updated:** 2026-07-17
