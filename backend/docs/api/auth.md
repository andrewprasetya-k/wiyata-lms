# 🔐 Authentication API Documentation

Base URL: `/api`

## 1. Register
Create a new user account and receive JWT token.

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
        "name": "Eduverse Academy"
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

All endpoints except `/login` and `/register` require authentication.

**Public (No Auth):**
- `POST /api/login`
- `POST /api/register`

**Protected (Auth Required):**
- All `/api/schools/*` endpoints
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

**Last Updated:** 2026-02-24
