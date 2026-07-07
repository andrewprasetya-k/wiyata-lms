# 📜 System Logs (Audit Trail) Module API Documentation

Base URL: `/api/logs`

## 1. List Logs by School
Retrieve a paginated history of system activities for a specific school.

- **URL:** `/school/:schoolId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `:schoolId` equals the active `school_id` from context. Requests where the param does not match the active school return `403 Forbidden`.
- **Query Parameters:** `page`, `limit`.

**Response Example:**
```json
{
  "data": [
    {
      "logId": "uuid",
      "userId": "uuid",
      "userName": "Admin Budi",
      "action": "CREATE_CLASS",
      "metadata": "{"classTitle": "X-IPA-1"}",
      "createdAt": "2026-02-13T17:00:00Z"
    }
  ],
  "totalItems": 500,
  ...
}
```
