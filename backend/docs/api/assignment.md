# 📝 Assignment & Grading Module API Documentation

Base URL: `/api/assignments`

## Categories

### 1. Create Category
- **URL:** `/categories`
- **Method:** `POST`
- **Body:** `{"schoolId": "uuid", "categoryName": "Kuis"}`

### 2. Get Categories by School
- **URL:** `/categories/school/:schoolCode`
- **Method:** `GET`
- **Response:** `SchoolWithAssignmentCategoriesDTO`

---

## Assignments

### 3. Create Assignment
- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Authorization:** The current teacher must teach the requested `subjectClassId`. Returns `403` if not.
- **Body:**
```json
{
  "schoolId": "uuid",
  "subjectClassId": "uuid",
  "categoryId": "uuid",
  "assignmentTitle": "string",
  "assignmentDescription": "string",
  "deadline": "2026-03-01T23:59:59Z",
  "allowLateSubmission": false,
  "mediaIds": ["uuid"]
}
```

### 4. List Assignments by Subject Class
- **URL:** `/subject-class/:subjectClassId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `admin`, `teacher`, or `student`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Admin can read active-school subject classes. Teacher can read only subject classes they teach. Student can read only subject classes in classes where they are enrolled.
- **Query Params:** `?page=1&limit=20&search=quiz`
  - `page` (optional): Page number, default 1
  - `limit` (optional): Items per page, default 20
  - `search` (optional): Search by assignment title or description
- **Response:** `AssignmentPerSubjectClassResponseDTO`

**Response Example:**
```json
{
  "subjectClass": {
    "subjectClassId": "uuid",
    "subjectCode": "MTK",
    "subjectName": "Matematika",
    "teacherId": "uuid",
    "teacherName": "John Doe"
  },
  "data": {
    "data": [
      {
        "assignmentId": "uuid",
        "assignmentTitle": "Quiz Chapter 1",
        "deadline": "2026-03-01T23:59:59Z",
        "allowLateSubmission": false
      }
    ],
    "totalItems": 25,
    "page": 1,
    "limit": 20,
    "totalPages": 2
  }
}
```

### 5. Get Assignment with Submissions
- **URL:** `/:assignmentId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send `teacherId`, `schoolUserId`, or `userId` in body/query.
- **Authorization:** The current teacher must teach the assignment's subject class. Returns `403` if not.
- **Note:** This endpoint gets assignment details with all submissions
- **Response:** `AssignmentWithSubmissionsDTO` (includes all submissions and assessments)

### 6. Get Subject Class Submissions
- **URL:** `/subject-class/submissions/:subjectClassId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send `teacherId`, `schoolUserId`, or `userId` in body/query.
- **Purpose:** Teacher-safe aggregate endpoint for the submissions of assignments in one teacher-owned subject class. This avoids frontend N+1 calls.
- **Authorization:** The current teacher must teach the requested subject class. Returns `403` if not.

**Response:**
```json
{
  "subjectClass": {
    "subjectClassId": "uuid",
    "subjectCode": "MTK",
    "subjectName": "Matematika",
    "teacherId": "uuid",
    "teacherName": "Nama Guru"
  },
  "assignments": [
    {
      "assignment": {
        "assignmentId": "uuid",
        "assignmentTitle": "Quiz Chapter 1",
        "subjectName": "Matematika",
        "categoryName": "Kuis",
        "deadline": "2026-03-01T23:59:59Z"
      },
      "submissionCount": 2,
      "gradedCount": 1,
      "pendingCount": 1,
      "submissions": [
        {
          "submissionId": "uuid",
          "studentName": "Nama Siswa",
          "submittedAt": "02-03-2026 10:30:00",
          "isLate": false,
          "attachments": [
            {
              "mediaId": "uuid",
              "mediaName": "jawaban.pdf",
              "fileUrl": "https://...",
              "mimeType": "application/pdf",
              "fileSize": 12345
            }
          ],
          "assessment": {
            "score": 90,
            "feedback": "Bagus",
            "assessorName": "Nama Guru",
            "assessedAt": "03-03-2026 09:00:00"
          }
        }
      ]
    }
  ],
  "summary": {
    "assignmentCount": 1,
    "submissionCount": 2,
    "gradedCount": 1,
    "pendingCount": 1,
    "lateCount": 0
  }
}
```

**Notes:**
- Returns `assignments: []` if the teacher owns the subject class but there are no assignments.
- Route must be registered before `/subject-class/:subjectClassId`.

### 7. Get Assignment Status
- **URL:** `/status/:id`
- **Method:** `GET`
- **Response:** Assignment with submission statistics (total, submitted, graded, pending)

### 8. Get My Submission Status
- **URL:** `/my-submission/:assignmentId`
- **Method:** `GET`
- **Auth:** Required
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Student identity is taken from the JWT token. Do not send `userId` in body or query.
- **Purpose:** Student-safe endpoint for the current user's submission status for one assignment. It does not expose other students' submissions.

**Not submitted response:**
```json
{
  "status": "not_submitted",
  "submission": null
}
```

**Submitted response:**
```json
{
  "status": "submitted",
  "submission": {
    "submissionId": "uuid",
    "assignmentId": "uuid",
    "submittedAt": "02-03-2026 10:30:00",
    "attachments": [
      {
        "mediaId": "uuid",
        "mediaName": "file.pdf",
        "fileUrl": "https://...",
        "mimeType": "application/pdf",
        "fileSize": 12345
      }
    ],
    "assessment": null
  }
}
```

**Graded response:**
```json
{
  "status": "graded",
  "submission": {
    "submissionId": "uuid",
    "assignmentId": "uuid",
    "submittedAt": "02-03-2026 10:30:00",
    "attachments": [
      {
        "mediaId": "uuid",
        "mediaName": "file.pdf",
        "fileUrl": "https://...",
        "mimeType": "application/pdf",
        "fileSize": 12345
      }
    ],
    "assessment": {
      "assessmentId": "uuid",
      "score": 90,
      "feedback": "Bagus",
      "assessedAt": "03-03-2026 09:00:00",
      "assessorName": "Nama Guru"
    }
  }
}
```

### 9. Update Assignment
- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** Required
- **Role:** `teacher` or `admin`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Teacher must teach the assignment's subject class. Admin can update only active-school assignments.
- **Body:** (all fields optional)
```json
{
  "categoryId": "uuid",
  "assignmentTitle": "string",
  "assignmentDescription": "string",
  "deadline": "2026-03-01T23:59:59Z",
  "allowLateSubmission": true,
  "mediaIds": ["uuid"]
}
```

### 10. Delete Assignment
- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required
- **Role:** `teacher` or `admin`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Teacher must teach the assignment's subject class. Admin can delete only active-school assignments.
- **Note:** Soft delete

---

## Submissions

### 11. Submit Assignment
- **URL:** `/submit/:assignmentId`
- **Method:** `POST`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Authorization:** Student must be enrolled in the class behind the assignment's subject class. Request `schoolId` must match active `SchoolId`.
- **Body:**
```json
{
  "schoolId": "uuid",
  "mediaIds": ["uuid"]
}
```
- **Note:** Upsert logic - updates existing submission if already submitted

### 12. Get Submission by ID
- **URL:** `/submit/:submissionId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send `teacherId`, `schoolUserId`, or `userId` in body/query.
- **Authorization:** The current teacher must teach the subject class of the submission's assignment. Returns `403` if not.
- **Response:** Includes `isLate` indicator and assessment if graded

### 13. Update Submission
- **URL:** `/submit/:submissionId`
- **Method:** `PATCH`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Authorization:** Submission must belong to the current JWT user and active school. Student must still be enrolled in the assignment class.
- **Body:**
```json
{
  "schoolId": "uuid",
  "mediaIds": ["uuid"]
}
```

### 14. Delete Submission
- **URL:** `/submit/:submissionId`
- **Method:** `DELETE`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Submission must belong to the current JWT user and active school. Student must still be enrolled in the assignment class.
- **Note:** Soft delete, can be restored by resubmitting

---

## Assessments

### 15. Grade Submission
- **URL:** `/assess/:submissionId`
- **Method:** `POST`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Authorization:** The current teacher must teach the subject class of the submission's assignment. Returns `403` if not.
- **Body:**
```json
{
  "score": 90.5,
  "feedback": "Good job"
}
```
- **Note:** Upsert logic - updates existing assessment if already graded

### 16. Update Assessment
- **URL:** `/assess/:submissionId`
- **Method:** `PATCH`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Authorization:** The current teacher must teach the subject class of the submission's assignment. Returns `403` if not.
- **Body:** (all fields optional)
```json
{
  "score": 95.0,
  "feedback": "Excellent work"
}
```

### 17. Delete Assessment
- **URL:** `/assess/:submissionId`
- **Method:** `DELETE`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send identity fields in body/query.
- **Authorization:** The current teacher must teach the subject class of the submission's assignment. Returns `403` if not.
- **Note:** Removes grading, submission remains

---

## Key Features

- **Late Submission Control:** `allowLateSubmission` flag per assignment
- **Upsert Logic:** Submissions and assessments auto-update if already exist
- **Soft Delete:** Assignments and submissions can be restored
- **IsLate Indicator:** Automatically calculated in submission responses
- **Attachments:** Support for multiple media files per assignment/submission
