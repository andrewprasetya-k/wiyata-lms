# 📝 Assignment & Grading Module API Documentation

Base URL: `/api/assignments`

## Categories

### 1. Create Category
- **URL:** `/categories`
- **Method:** `POST`
- **School Context:** Requires `SchoolId` header
- **Authorization:** `schoolId` in the request body must match the active `SchoolId` header.
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
- **Category Rule:** `categoryId` must exist and belong to the active school.
- **Attachment Rule:** Every `mediaId` must exist, belong to the active school, and be owned/uploaded by the current teacher.
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
          "submittedAt": "2026-03-02T10:30:00Z",
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
            "assessedAt": "2026-03-03T09:00:00Z"
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

### 7. Get Teacher Assignments Inbox
- **URL:** `/teacher-assignments`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send `teacherId`, `schoolUserId`, or `userId` in body/query.
- **Purpose:** Teacher-safe aggregate endpoint for the global assignments overview across all subject classes taught by the current teacher in the active school.
- **Authorization:** Returns only assignments from active-school subject classes owned by the current teacher where teacher class enrollment is still active (`left_at IS NULL`).

**Response:**
```json
{
  "summary": {
    "totalAssignments": 2,
    "activeAssignments": 1,
    "overdueAssignments": 1,
    "pendingReviewCount": 2,
    "totalSubmissions": 4
  },
  "items": [
    {
      "assignmentId": "uuid",
      "subjectClassId": "uuid",
      "assignmentTitle": "Quiz Chapter 1",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "className": "Kelas 10 A",
      "classCode": "10A",
      "categoryName": "Kuis",
      "deadline": "2026-03-01T23:59:59Z",
      "submissionCount": 2,
      "pendingCount": 1,
      "gradedCount": 1,
      "lateCount": 0
    }
  ]
}
```

**Counting Rules:**
- `submissionCount`: all active submissions for the assignment.
- `gradedCount`: submissions that already have an assessment.
- `pendingCount`: submissions that exist but do not have an assessment yet.
- `lateCount`: submissions where `submittedAt > deadline`, only when deadline exists.
- `activeAssignments`: assignments where deadline is not past or deadline is empty.
- `overdueAssignments`: assignments where deadline is past.
- `pendingReviewCount`: sum of `pendingCount` across all returned items.
- `totalSubmissions`: sum of `submissionCount` across all returned items.
- Items are assignment-level rows and may include assignments with zero submissions.

### 8. Get Teacher Submissions Inbox
- **URL:** `/teacher-submissions`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Teacher identity is taken from the JWT token. Do not send `teacherId`, `schoolUserId`, or `userId` in body/query.
- **Purpose:** Teacher-safe aggregate endpoint for the global submissions inbox across all subject classes taught by the current teacher in the active school.
- **Authorization:** Returns only assignments from active-school subject classes owned by the current teacher. Teacher class enrollment must still be active.

**Response:**
```json
{
  "summary": {
    "totalSubmissions": 4,
    "pendingCount": 2,
    "gradedCount": 2,
    "lateCount": 1
  },
  "items": [
    {
      "assignmentId": "uuid",
      "subjectClassId": "uuid",
      "assignmentTitle": "Quiz Chapter 1",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "className": "Kelas 10 A",
      "classCode": "10A",
      "deadline": "2026-03-01T23:59:59Z",
      "submissionCount": 2,
      "pendingCount": 1,
      "gradedCount": 1,
      "lateCount": 0
    }
  ]
}
```

**Counting Rules:**
- `submissionCount`: all active submissions for the assignment.
- `gradedCount`: submissions that already have an assessment.
- `pendingCount`: submissions that exist but do not have an assessment yet.
- `lateCount`: submissions where `submittedAt > deadline`, only when deadline exists.
- Summary totals are sums across all returned items.
- Items are assignment-level rows with at least one submission.

### 9. Get Student Assignments Inbox
- **URL:** `/student-assignments`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Student identity is taken from the JWT token. Do not send `userId`, `schoolUserId`, or enrollment fields in body/query.
- **Purpose:** Student-safe aggregate endpoint for the global assignments list across all active classes and subject classes where the current student is enrolled.
- **Authorization:** Returns only assignments from active-school classes where the current student has active enrollment (`left_at IS NULL`).

**Response:**
```json
{
  "summary": {
    "totalAssignments": 3,
    "notSubmittedCount": 1,
    "submittedCount": 2,
    "gradedCount": 1,
    "overdueCount": 1
  },
  "items": [
    {
      "assignmentId": "uuid",
      "subjectClassId": "uuid",
      "assignmentTitle": "Quiz Chapter 1",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "className": "Kelas 10 A",
      "classCode": "10A",
      "categoryName": "Kuis",
      "deadline": "2026-03-01T23:59:59Z",
      "submissionId": "uuid",
      "submittedAt": "2026-03-01T10:30:00Z",
      "score": 90,
      "isSubmitted": true,
      "isGraded": true,
      "isOverdue": false,
      "isSubmittedLate": false
    }
  ]
}
```

**Status Rules:**
- `totalAssignments`: all returned active assignments.
- `isSubmitted`: current student has an active submission for the assignment.
- `isGraded`: current student submission has an assessment.
- `notSubmittedCount`: assignments where `isSubmitted = false`.
- `submittedCount`: assignments where `isSubmitted = true`.
- `gradedCount`: assignments where `isGraded = true`.
- `isOverdue`: deadline has passed and the current student has not submitted.
- `overdueCount`: count of `isOverdue = true`.
- `isSubmittedLate`: `submittedAt > deadline`, only when both values exist.

### 10. Get Student Assignment Detail
- **URL:** `/student/:assignmentId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Authorization:** The current student must have an active enrollment (`left_at IS NULL`) in the assignment's subject class and the assignment must belong to the active school.
- **Purpose:** Returns one student-safe assignment detail without returning any other student's submissions or teacher-only aggregate data.

**Response:**
```json
{
  "assignmentId": "uuid",
  "subjectClassId": "uuid",
  "subjectName": "Matematika",
  "subjectCode": "MTK",
  "assignmentTitle": "Quiz Chapter 1",
  "assignmentDescription": "Kerjakan soal yang tersedia.",
  "categoryName": "Kuis",
  "deadline": "2026-03-01T23:59:59Z",
  "allowLateSubmission": false,
  "createdAt": "2026-03-01T09:00:00Z",
  "updatedAt": "2026-03-01T09:00:00Z",
  "attachments": [
    {
      "mediaId": "uuid",
      "mediaName": "soal.pdf",
      "fileSize": 12345,
      "mimeType": "application/pdf",
      "fileUrl": "https://...",
      "thumbnailUrl": "",
      "ownerType": "user",
      "createdAt": "2026-03-01T08:55:00Z"
    }
  ]
}
```

Attachment entries whose media has been soft-deleted or does not belong to the same school are omitted. Non-HTTP(S) file and thumbnail URLs are returned as empty strings.
The web client uses absolute HTTP(S) `fileUrl` values directly for inline image/PDF preview and does not prefix them with the API base URL.

### 11. Get Assignment Status
- **URL:** `/status/:id`
- **Method:** `GET`
- **Response:** Assignment with submission statistics (total, submitted, graded, pending)

### 12. Get My Submission Status
- **URL:** `/my-submission/:assignmentId`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Student identity is taken from the JWT token. Do not send `userId` in body or query.
- **Purpose:** Student-safe endpoint for the current user's submission status for one assignment. It does not expose other students' submissions.
- **Authorization:** The assignment must belong to the active school and the current student must still have active enrollment (`left_at IS NULL`) in its subject class.

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
    "submittedAt": "2026-03-02T10:30:00Z",
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
    "submittedAt": "2026-03-02T10:30:00Z",
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
      "assessedAt": "2026-03-03T09:00:00Z",
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
- **Category Rule:** If `categoryId` is provided, it must exist and belong to the active school.
- **Attachment Rule:** If `mediaIds` is provided, every media must exist and belong to the active school. Teachers can attach only their own uploaded media; admins can attach active-school media.
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
- **Attachment Rule:** Every `mediaId` must exist, belong to the active school, and be owned/uploaded by the current student.
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
- **Attachment Rule:** Every `mediaId` must exist, belong to the active school, and be owned/uploaded by the current student.
- **Business Rules (mutability check, shared with Delete Submission below):**
  - The submission must **not** already have an assessment. If it does, the request is rejected — a graded submission cannot be edited.
  - The assignment must still be open: this passes if `deadline` is not set, if `allowLateSubmission` is `true`, or if the current time is still before `deadline`. Otherwise the request is rejected.
  - These two checks run before any write; if either fails, no attachment or timestamp change is applied.
- **Side Effect:** On success, `submittedAt` is updated to the current time (i.e. editing a submission re-stamps it as if freshly submitted), and if `mediaIds` is provided, the submission's attachments are fully replaced with the given list.
- **Body:**
```json
{
  "schoolId": "uuid",
  "mediaIds": ["uuid"]
}
```
- **Error Responses (business rule violations, `400 Bad Request`):**
```json
{ "error": "Cannot modify a graded submission" }
```
```json
{ "error": "Cannot modify submission after the assignment is closed" }
```

### 14. Withdraw (Delete) Submission
- **URL:** `/submit/:submissionId`
- **Method:** `DELETE`
- **Auth:** Required
- **Role:** `student`
- **School Context:** Requires `SchoolId` header
- **Auth Note:** Actor identity is taken from the JWT token.
- **Request Parameters:** Path parameter `submissionId` only. No query parameters or request body.
- **Authorization/Permission Requirements:** Resolved by `authorizeStudentForSubmission` before the service is called:
  1. Caller must be an authenticated user with the `student` role (enforced by the route's `RequireRole` middleware).
  2. The submission must belong to the active school (matching the `SchoolId` header).
  3. The submission's `userId` must match the current JWT user — a student can only withdraw their own submission.
  4. The submission's parent assignment must also belong to the active school.
  5. The student must still have an active (non-left) enrollment in the assignment's subject class.
  - Any failure returns `401` (no/invalid identity), `400` (missing `SchoolId` header), or `403` (school/ownership/enrollment mismatch), before the business-rule checks below ever run.
- **Business Rules (a.k.a. "withdrawal eligibility" — same `validateSubmissionMutable` check used by Update Submission):**
  1. **Not graded:** the submission must have no associated assessment. If an assessment exists, withdrawal is rejected with `400` and `{"error": "Cannot modify a graded submission"}`. There is no way to withdraw a graded submission through this endpoint.
  2. **Assignment still open:** withdrawal is allowed only if the assignment has no `deadline`, or `allowLateSubmission` is `true`, or the current time is still before `deadline`. If the assignment is closed (a deadline has passed and late submission is not allowed), withdrawal is rejected with `400` and `{"error": "Cannot modify submission after the assignment is closed"}`.
  - Both checks are evaluated in the service layer immediately before the delete transaction runs; if either fails, the transaction never starts and nothing is modified.
- **State Transition:** `Submitted` → `Not Submitted`. Concretely:
  - The submission row is **soft-deleted** (`deleted_at` is set) — the row is not physically removed.
  - All attachment links for the submission (`edv.attachments` rows with `source_type = 'submission'`) are unlinked in the same transaction. The underlying uploaded media files/records are **not** deleted — only the link between the submission and its attachments is removed.
  - Because `(sbm_asg_id, sbm_usr_id)` is a unique constraint on the submissions table, a subsequent `POST /submit/:assignmentId` (Submit) by the same student for the same assignment will **resurrect this same soft-deleted row** (same `submissionId`, `deleted_at` reset to null, `submittedAt` and attachments refreshed) rather than creating a new row. This is existing `Submit`/`UpsertSubmission` behavior, unchanged by this feature — withdrawal relies on it to make "withdraw then resubmit" work.
  - After withdrawal, `GET /my-submission/:assignmentId` (see §12) reports `"status": "not_submitted", "submission": null` for this assignment/student, since the soft-deleted row is excluded by the default query scope.
- **Side Effects:** None beyond the soft-delete and attachment unlink described above — no notification is sent to the teacher or student on withdrawal, and no separate audit-log entry is written (state is recoverable/inspectable only via the `deleted_at` timestamp on the submission row itself).
- **Success Response:** `200 OK`
```json
{ "message": "Submission deleted" }
```
- **Error Responses:**
  - `401 Unauthorized` — no valid JWT / user identity.
  - `400 Bad Request` — missing `SchoolId` header, **or** one of the two business-rule violations above:
    ```json
    { "error": "Cannot modify a graded submission" }
    ```
    ```json
    { "error": "Cannot modify submission after the assignment is closed" }
    ```
  - `403 Forbidden` — submission does not belong to the active school, does not belong to the current user, the assignment does not belong to the active school, or the student's enrollment is no longer active.
  - `404 Not Found` — `submissionId` does not exist.
- **Implementation Notes:**
  - The success response message text (`"Submission deleted"`) has not been updated to reflect the "withdraw" terminology used by the frontend and in this document's business rules — the HTTP contract still reads as a deletion, not a withdrawal, even though the actual effect (soft-delete + resurrectable via resubmission) is a withdrawal in practice. This is a documentation/naming inconsistency in the current implementation, not a bug.
  - The two business-rule error strings are matched by exact string comparison in the handler (`handleSubmissionMutationError`), the same pattern already used for the `"submission past due"` error on `Submit`. Changing either error message on the service side without updating the handler's `switch` statement will silently fall through to the generic `HandleError` path instead of returning the intended `400`.

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
- **Data Integrity:** One submission can have at most one assessment. `POST` creates the assessment if none exists; if an assessment already exists, it updates the existing assessment for that submission.
- **Body:**
```json
{
  "score": 90.5,
  "feedback": "Good job"
}
```
- **Note:** Idempotent upsert by `submissionId` - updates existing assessment if already graded.

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
- **Assessment Uniqueness:** `assessments.asm_sbm_id` should be unique at database level. Backend also upserts by `submissionId` and removes duplicate assessment rows for the same submission during grading.
- **Soft Delete:** Assignments and submissions can be restored
- **IsLate Indicator:** Automatically calculated in submission responses
- **Attachments:** Support for multiple media files per assignment/submission
