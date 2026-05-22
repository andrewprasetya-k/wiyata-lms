# 📖 Subject Class (Penugasan Guru) Module API Documentation

Base URL: `/api/subject-classes`

## 1. Assign Subject and Teacher to Class
Link a specific subject and teacher to a class.

- **URL:** `/assign`
- **Method:** `POST`
- **Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `classId` | uuid | Yes | |
| `subjectId` | uuid | Yes | |
| `teacherId` | uuid | Yes | Reference to school_users (ID Guru) |

---

## 2. List Subjects in Class
Retrieve all subjects and their teachers for a specific class.

- **URL:** `/class/:classId`
- **Method:** `GET`
- **Response:** `SubjectPerClassDTO` (Includes class header and list of subject assignments)

---

## 3. List Current Teacher Subject Classes
Retrieve subject classes taught by the current logged-in teacher in the active school context.

- **URL:** `/my-teaching`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token and current school membership. Do not send `userId` or `schoolUserId` in body/query.

**Response:**
```json
{
  "data": [
    {
      "subjectClassId": "uuid",
      "classId": "uuid",
      "className": "XII IPA 1",
      "classCode": "XII-IPA-1",
      "subjectId": "uuid",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "studentCount": 30,
      "materialCount": 4,
      "assignmentCount": 3,
      "pendingSubmissions": 8
    }
  ]
}
```

**Notes:**
- `studentCount` is counted from class enrollments with role `student`.
- `materialCount` and `assignmentCount` are scoped to the subject class.
- `pendingSubmissions` counts submitted assignment submissions in this subject class that do not have an assessment yet.
- This endpoint is intended for teacher subject workspace pages such as `/teacher/subjects`.

---

## 4. Get Assignment Detail
- **URL:** `/:id`
- **Method:** `GET`

---

## 5. Update Assignment
Update teacher or subject for an existing assignment.

- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `subjectId` | uuid | No | |
| `teacherId` | uuid | No | |

---

## 6. Remove Assignment (Unassign)
- **URL:** `/:id`
- **Method:** `DELETE`
