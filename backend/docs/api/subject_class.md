# 📖 Subject Class (Teacher Assignment) Module API Documentation

Base URL: `/api/subject-classes`

`subject_class` adalah workspace belajar untuk kombinasi class + subject + teacher school_user.
Materials dan assignments selalu melekat ke `subjectClassId`.

## 1. Assign Subject and Teacher to Class
Link one active-school subject and one active-school teacher school_user to one active-school class.

- **URL:** `/assign`
- **Method:** `POST`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **School Context:** Requires `SchoolId` header
- **Body:**
```json
{
  "classId": "uuid",
  "subjectId": "uuid",
  "teacherId": "uuid"
}
```

**Field notes:**
- `teacherId` means `school_users.scu_id`, not `users.usr_id`.

**Ownership and eligibility rules:**
- `classId` must belong to the active school.
- `subjectId` must belong to the active school.
- `teacherId` must be a `school_users.scu_id` in the active school.
- The teacher school_user must have school role `teacher`.
- The teacher school_user must already be enrolled in the selected class with `class_role = teacher`.
- MVP allows only one subject_class for the same `classId + subjectId`. Co-teaching is deferred.

---

## 2. List Subjects in Class
Retrieve all subject classes and their teachers for a specific class.

- **URL:** `/class/:classId`
- **Method:** `GET`
- **Auth:** Required school member in active `SchoolId` context
- **Ownership rules:** `classId` must belong to the active school.
- **Response:** `SubjectPerClassDTO`

**Response Example:**
```json
{
  "class": {
    "classId": "uuid",
    "classTitle": "Kelas 10 A",
    "classCode": "10A"
  },
  "subjects": [
    {
      "subjectClassId": "uuid",
      "subjectId": "uuid",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "teacherId": "school_user_uuid",
      "teacherName": "Nama Guru"
    }
  ]
}
```

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

## 4. Get Subject Class Detail
- **URL:** `/:id`
- **Method:** `GET`
- **Auth:** Required school member in active `SchoolId` context
- **Ownership rules:** Subject class must belong to the active school.

---

## 5. Update Subject Class Assignment
Update subject or teacher for an existing subject_class.

- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **School Context:** Requires `SchoolId` header
- **Body:**
```json
{
  "subjectId": "uuid",
  "teacherId": "uuid"
}
```

**Rules:**
- Existing subject_class must belong to the active school.
- If `subjectId` is changed, the new subject must belong to the active school.
- If `teacherId` is changed, it must satisfy the same teacher eligibility rules as create.
- MVP still prevents duplicate `classId + subjectId` assignments.

---

## 6. Remove Subject Class Assignment
- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **Ownership rules:** Subject class must belong to the active school.
