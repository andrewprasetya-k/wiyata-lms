# 🏫 School Module API Documentation

Base URL: `/api/schools`

## 1. List Schools
Retrieve a paginated list of schools with filtering, searching, and sorting capabilities.

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`): Page number.
  - `limit` (default: `10`): Items per page.
  - `search` (optional): Filter by school name or code (Case Insensitive).
  - `status` (optional): 
    - `active` (default): Only non-deleted schools.
    - `deleted`: Only soft-deleted schools.
    - `all`: All schools including deleted.
  - `sortBy` (optional): 
    - `name`: Sort by school name.
    - `code`: Sort by school code.
    - `createdAt`: Sort by creation date (default).
    - `updatedAt`: Sort by last update date.
  - `order` (optional): `asc` (A-Z) or `desc` (Z-A/Newest, default).

**Response Example:**
```json
{
  "data": [
    {
      "schoolId": "uuid-string",
      "schoolName": "Eduverse Academy",
      "schoolCode": "EDU01",
      "schoolLogo": "uuid-media-id",
      "schoolAddress": "Jl. Merdeka No. 1",
      "schoolEmail": "admin@edu.com",
      "schoolPhone": "081234567890",
      "schoolWebsite": "https://edu.com",
      "isDeleted": false,
      "createdAt": "13-02-2026 09:00:00",
      "updatedAt": "13-02-2026 09:00:00"
    }
  ],
  "totalItems": 50,
  "page": 1,
  "limit": 10,
  "totalPages": 5
}
```

---

## 2. Get School Summary
Get high-level statistics for school management cards.

- **URL:** `/summary`
- **Method:** `GET`

**Response Example:**
```json
{
  "totalActive": 10,
  "totalDeleted": 2,
  "totalSchools": 12
}
```

---

## 3. Check Code Availability
Quickly check if a school code is already taken.

- **URL:** `/check-code/:schoolCode`
- **Method:** `GET`

**Response Example:**
```json
{
  "schoolCode": "EDU01",
  "available": true
}
```

---

## 4. Create School
Register a new school. Input will be automatically trimmed of leading/trailing spaces.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**
| Field | Type | Required | Validation |
| :--- | :--- | :--- | :--- |
| `schoolName` | string | Yes | Min 1 char |
| `schoolCode` | string | No | Unique, auto-generated if empty |
| `schoolLogo` | uuid | No | Reference to Media ID |
| `schoolAddress`| string | Yes | Min 1 char |
| `schoolEmail` | string | Yes | Unique, valid email format |
| `schoolPhone` | string | Yes | Unique, numeric, min 10 chars |
| `schoolWebsite`| string | No | Valid URL format |

**Example Response (201 Created):**
```json
{
  "schoolId": "uuid",
  "schoolName": "Eduverse Academy",
  "schoolCode": "EDU01",
  ...
}
```

---

## 5. Get School Detail
Get full information of a specific school by its code.

- **URL:** `/:schoolCode`
- **Method:** `GET`

**Response Example:** Same as List School item.

---

## 6. Update School
Update existing school information. Partial updates are supported.

- **URL:** `/:schoolCode`
- **Method:** `PATCH`
- **Body:** Same as Create School (all fields are optional).

**Error Responses:**
- `400 Bad Request`: Validation failed (e.g., invalid email format).
- `500 Internal Server Error`: Conflict error (e.g., "email already exists").

---

## 7. Management Actions

### Soft Delete
Moves school to trash (sets `deleted_at`).
- **URL:** `/:schoolCode`
- **Method:** `DELETE`
- **Response:** `{"message": "School deleted successfully"}`

### Restore
Brings back a school from trash.
- **URL:** `/restore/:schoolCode`
- **Method:** `PATCH`
- **Response:** `{"message": "School restored successfully"}`

### Hard Delete (Permanent)
Permanently removes school from database. **Warning: This cannot be undone.**
- **URL:** `/permanent/:schoolCode`
- **Method:** `DELETE`
- **Response:** `{"message": "School permanently deleted successfully"}`
