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
- **URL:** `/`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
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
- **Note:** This endpoint gets assignment details with all submissions
- **Response:** `AssignmentWithSubmissionsDTO` (includes all submissions and assessments)

### 6. Get Assignment Status
- **URL:** `/status/:id`
- **Method:** `GET`
- **Response:** Assignment with submission statistics (total, submitted, graded, pending)

### 7. Update Assignment
- **URL:** `/:id`
- **Method:** `PATCH`
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

### 8. Delete Assignment
- **URL:** `/:id`
- **Method:** `DELETE`
- **Note:** Soft delete

---

## Submissions

### 9. Submit Assignment
- **URL:** `/submit/:assignmentId`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "schoolId": "uuid",
  "mediaIds": ["uuid"]
}
```
- **Note:** Upsert logic - updates existing submission if already submitted

### 10. Get Submission by ID
- **URL:** `/submit/:submissionId`
- **Method:** `GET`
- **Response:** Includes `isLate` indicator and assessment if graded

### 11. Update Submission
- **URL:** `/submit/:submissionId`
- **Method:** `PATCH`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "schoolId": "uuid",
  "mediaIds": ["uuid"]
}
```

### 12. Delete Submission
- **URL:** `/submit/:submissionId`
- **Method:** `DELETE`
- **Note:** Soft delete, can be restored by resubmitting

---

## Assessments

### 13. Grade Submission
- **URL:** `/assess/:submissionId`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "score": 90.5,
  "feedback": "Good job"
}
```
- **Note:** Upsert logic - updates existing assessment if already graded

### 14. Update Assessment
- **URL:** `/assess/:submissionId`
- **Method:** `PATCH`
- **Body:** (all fields optional)
```json
{
  "score": 95.0,
  "feedback": "Excellent work"
}
```

### 15. Delete Assessment
- **URL:** `/assess/:submissionId`
- **Method:** `DELETE`
- **Note:** Removes grading, submission remains

---

## Key Features

- **Late Submission Control:** `allowLateSubmission` flag per assignment
- **Upsert Logic:** Submissions and assessments auto-update if already exist
- **Soft Delete:** Assignments and submissions can be restored
- **IsLate Indicator:** Automatically calculated in submission responses
- **Attachments:** Support for multiple media files per assignment/submission
