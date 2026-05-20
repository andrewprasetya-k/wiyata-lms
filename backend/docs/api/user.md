# 👤 User Profile Module API Documentation

Base URL: `/api/users`

## 1. List All Users
Retrieve a paginated list of all global users.

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`)
  - `limit` (default: `10`)
  - `search` (optional): Search by full name or email.

**Response Example:**
```json
{
  "data": [...],
  "totalItems": 100,
  "page": 1,
  "limit": 10,
  "totalPages": 10
}
```

---

## 2. Create User
Register a new global user profile. Password will be securely hashed.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `fullName` | string | Yes | |
| `email` | string | Yes | Unique |
| `password` | string | Yes | Min 6 characters |

---

## 3. Get User Detail
- **URL:** `/:id`
- **Method:** `GET`

---

## 4. Update User Detail
- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:** `fullName`, `email`.

---

## 5. Change Password
Update user password. Old password verification is required.

- **URL:** `/:id/change-password`
- **Method:** `PATCH`
- **Body:**
  - `oldPassword` (string, required)
  - `newPassword` (string, required, min 6)

---

## 6. Delete User
- **URL:** `/:id`
- **Method:** `DELETE`
