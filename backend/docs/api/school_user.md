# 🏫 School User (Membership) Module API Documentation

Base URL: `/api/school-users`

## 1. Enroll User to School

Connect an existing user to a school.

- **URL:** `/enroll`
- **Method:** `POST`
- **Body:**
  | Field | Type | Required | Note |
  | :--- | :--- | :--- | :--- |
  | `userId` | uuid | Yes | |
  | `schoolId` | uuid | Yes | |

---

## 2. List Members by School

Retrieve all users who belong to a specific school, including school details.

- **URL:** `/school/:schoolCode`
- **Method:** `GET`
- **Query Params:** `?page=1&limit=20&search=john`
  - `page` (optional): Page number, default 1
  - `limit` (optional): Items per page, default 20
  - `search` (optional): Search by user name or email

**Response Example:**

```json
{
  "school": {
    "schoolId": "uuid",
    "schoolName": "Wiyata Academy",
    "schoolCode": "SD-ANGKASA"
  },
  "members": {
    "data": [
      {
        "schoolUserId": "uuid",
        "userId": "uuid",
        "fullName": "John Doe",
        "email": "john@doe.com",
        "roles": ["Teacher", "Admin"],
        "createdAt": "13-02-2026 14:00:00"
      }
    ],
    "totalItems": 150,
    "page": 1,
    "limit": 20,
    "totalPages": 8
  }
}
```

---

## 3. List Schools by User

Retrieve all schools that a specific user belongs to.

- **URL:** `/user/:userId`
- **Method:** `GET`

---

## 4. Unenroll User

Remove a user's membership from a school.

- **URL:** `/:id`
- **Method:** `DELETE`
