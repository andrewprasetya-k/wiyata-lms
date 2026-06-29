# 🗓️ Term (Semester) Module API Documentation

Base URL: `/api/terms`

## 1. List All Terms

Retrieve a paginated list of all terms (Super Admin view).

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`)
  - `limit` (default: `10`)
  - `search` (optional): Search by term name.

---

## 2. List Terms by Academic Year

Retrieve all terms for a specific academic year.

- **URL:** `/academic-year/:academicYearId`
- **Method:** `GET`

**Response Example:**

```json
[
  {
    "termId": "uuid-string",
    "academicYearId": "uuid-acy-id",
    "academicYearName": "2023/2024",
    "schoolName": "Wiyata Academy",
    "termName": "Semester Ganjil",
    "isActive": true,
    "createdAt": "13-02-2026 11:00:00"
  }
]
```

---

## 3. Get Term Detail

Retrieve detail of a specific term by its ID.

- **URL:** `/:id`
- **Method:** `GET`

---

## 4. Create Term

Create a new term for an academic year. Status is `false` by default.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**
  | Field | Type | Required | Note |
  | :--- | :--- | :--- | :--- |
  | `academicYearId` | uuid | Yes | Reference to Academic Year ID |
  | `termName` | string | Yes | e.g., "Semester Ganjil" |

---

## 5. Update Term

Update name of a term.

- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:**
  - `termName` (string)

---

## 6. Activate Term

Set a term as the active one for its academic year. This will automatically deactivate all other terms in the same academic year.

- **URL:** `/activate/:id`
- **Method:** `PATCH`
- **Response:** `{"message": "Term activated successfully"}`

---

## 7. Deactivate Term

Manually deactivate a term.

- **URL:** `/deactivate/:id`
- **Method:** `PATCH`
- **Response:** `{"message": "Term deactivated successfully"}`

---

## 8. Delete Term

Permanently remove a term.

- **URL:** `/:id`
- **Method:** `DELETE`
