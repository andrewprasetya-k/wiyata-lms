# 📚 Subject (Mata Pelajaran) Module API Documentation

Base URL: `/api/subjects`

## 1. List All Subjects

Retrieve a paginated list of all subjects (Super Admin view).

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`)
  - `limit` (default: `10`)
  - `search` (optional): Search by name or code.

---

## 2. List Subjects by School

Retrieve all subjects for a specific school, including school details.

- **URL:** `/school/:schoolCode`
- **Method:** `GET`

**Response Example:**

```json
{
  "school": {
    "schoolId": "uuid",
    "schoolName": "Wiyata Academy",
    "schoolCode": "EDU01",
    ...
  },
  "subjects": [
    {
      "subjectId": "uuid",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "color": "#4F46E5",
      "createdAt": "2026-02-13T15:00:00Z"
    }
  ]
}
```

---

## 3. Get Subject Detail (by ID)

- **URL:** `/:id`
- **Method:** `GET`

---

## 4. Get Subject Detail (by Code)

Retrieve subject details by school code and subject code.

- **URL:** `/school/:schoolCode/:subjectCode`
- **Method:** `GET`

---

## 5. Create Subject

Register a new subject for a school.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**
  | Field | Type | Required | Note |
  | :--- | :--- | :--- | :--- |
  | `schoolId` | uuid | Yes | |
  | `subjectName`| string | Yes | e.g., "Matematika" |
  | `subjectCode`| string | Yes | Unique per school, e.g., "MTK" |
  | `color` | string | No | Hex color: `#RGB`, `#RRGGBB`, or `#RRGGBBAA`. Used by calendar/timeline visual identity. |

---

## 6. Update Subject

- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `subject.SchoolID == activeSchoolID`. Returns `403 Forbidden` if the subject belongs to a different school.
- **Body:** `subjectName`, `subjectCode`, `color`.

`color` is optional. Send an empty string to clear it and let the frontend use its deterministic fallback color.

---

## 7. Delete Subject

- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `subject.SchoolID == activeSchoolID` before deleting.
