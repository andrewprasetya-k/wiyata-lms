# 🏫 Class (Grup Belajar) Module API Documentation

Base URL: `/api/classes`

## 1. List All Classes
Retrieve a paginated list of all classes.

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Parameters:**
  - `page` (default: `1`)
  - `limit` (default: `10`)
  - `search` (optional): Search by title or code.
  - `schoolCode` (optional): Filter by school.
  - `termId` (optional): Filter by semester.

**Response Example:**
```json
{
  "data": [
    {
      "classId": "uuid",
      "schoolName": "Eduverse Academy",
      "termName": "Semester Ganjil",
      "academicYearName": "2023/2024",
      "classCode": "X-IPA-1",
      "classTitle": "Kelas 10 IPA 1",
      "creatorName": "Admin Budi",
      "isActive": true,
      ...
    }
  ],
  ...
}
```

---

## 2. Get Class Detail
- **URL:** `/:id`
- **Method:** `GET`

---

## 3. Create Class
- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `schoolId` | uuid | Yes | |
| `termId` | uuid | Yes | Reference to Semester |
| `classCode` | string | Yes | e.g., "X-IPA-1" |
| `classTitle`| string | Yes | e.g., "IPA 1" |
| `classDescription` | string | No | |

---

## 4. Update Class
- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:** `classTitle`, `classDescription`, `isActive`.

---

## 5. Delete Class
- **URL:** `/:id`
- **Method:** `DELETE`
