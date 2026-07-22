# 📅 Academic Year Module API Documentation

Base URL: `/api/academic-years`

## 1. List All Academic Years

Retrieve a paginated list of all academic years (Super Admin view).

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`)
  - `limit` (default: `10`)
  - `search` (optional): Search by academic year name.

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

## 2. List Academic Years by School

Retrieve all academic years for a specific school using its code.

- **URL:** `/school/:schoolCode`
- **Method:** `GET`

**Response Example:**

```json
[
  {
    "academicYearId": "uuid-string",
    "schoolId": "uuid-school-id",
    "schoolName": "Wiyata Academy",
    "schoolCode": "EDU01",
    "academicYearName": "2023/2024",
    "isActive": true,
    "createdAt": "2026-02-13T10:00:00Z"
  }
]
```

---

## 3. Get Academic Year Detail

Retrieve detail of a specific academic year by its ID.

- **URL:** `/:id`
- **Method:** `GET`

---

## 4. Create Academic Year

Create a new academic year for a school. Status is `false` by default.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth Note:** School is always taken from the caller's active school context (`SchoolId` header), never from the request body — any `schoolId` sent in the body is ignored.
- **Body:**
  | Field | Type | Required | Note |
  | :--- | :--- | :--- | :--- |
  | `academicYearName` | string | Yes | e.g., "2023/2024" |

---

## 5. Update Academic Year

Update basic information of an academic year.

- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `academicYear.SchoolID == activeSchoolID`. Returns `403 Forbidden` if the academic year belongs to a different school.
- **Body:**
  - `academicYearName` (string)

---

## 6. Activate Academic Year

Set an academic year as the active one for its school. This will automatically deactivate all other academic years in the same school.

- **URL:** `/activate/:id`
- **Method:** `PATCH`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `academicYear.SchoolID == activeSchoolID` before activating.
- **Response:** `{"message": "Academic year activated successfully"}`

---

## 7. Deactivate Academic Year

Manually deactivate an academic year.

- **URL:** `/deactivate/:id`
- **Method:** `PATCH`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `academicYear.SchoolID == activeSchoolID` before deactivating.
- **Response:** `{"message": "Academic year deactivated successfully"}`

---

## 8. Delete Academic Year

Permanently remove an academic year.

- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** `RequireSchoolMember + RequireRole("admin", "super_admin")`
- **Ownership:** Handler verifies `academicYear.SchoolID == activeSchoolID` before deleting.
