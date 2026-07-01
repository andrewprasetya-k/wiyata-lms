# 👥 Enrollment Module API Documentation

Base URL: `/api/enrollments`

## 1. Enroll Members to Class
- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **Body:**
```json
{
  "schoolId": "uuid",
  "schoolUserIds": ["uuid1", "uuid2"],
  "classId": "uuid",
  "role": "teacher|student"
}
```
- **Ownership rules:**
  - `schoolId` must match active `SchoolId` context.
  - `classId` must belong to the active school.
  - Every `schoolUserId` must belong to the active school.
- **Note:** Bulk enrollment supported. Existing duplicate class enrollments are skipped.
- **Re-enroll behavior:** If the member previously left the class (`leftAt` is set), enrolling the same member again reactivates the existing enrollment by clearing `leftAt` and updating the class role if needed. The original `joinedAt` is intentionally preserved as the first time the member joined the class.

## 2. Get Enrollments by Class
- **URL:** `/class/:classId`
- **Method:** `GET`
- **Auth:** Required school member in active `SchoolId` context.
- **Ownership rules:** `classId` must belong to the active school.
- **Query Params:** `?page=1&limit=20&search=john`
  - `page` (optional): Page number, default 1
  - `limit` (optional): Items per page, default 20
  - `search` (optional): Search by user name or email
- **Response:** `ClassWithMembersDTO`
- **Active filter:** Returns active enrollments only (`leftAt` is empty and omitted from normal active response items).

**Response Example:**
```json
{
  "class": {
    "classId": "uuid",
    "classTitle": "12 IPA 1",
    "classCode": "12IPA1"
  },
  "members": {
    "data": [
      {
        "enrollmentId": "uuid",
        "schoolUserId": "uuid",
        "userFullName": "John Doe",
        "userEmail": "john@example.com",
        "role": "student",
        "joinedAt": "2026-02-23T10:00:00Z"
      }
    ],
    "totalItems": 40,
    "page": 1,
    "limit": 20,
    "totalPages": 2
  }
}
```

## 3. Get Enrollments by Member
- **URL:** `/member/:schoolUserId`
- **Method:** `GET`
- **Auth:** Required school member in active `SchoolId` context.
- **Ownership rules:** `schoolUserId` must belong to the active school.
- **Response:** List of classes the member is enrolled in
- **Active filter:** Returns active enrollments only (`leftAt` is empty and omitted from normal active response items).

## 4. Get Enrollment by ID
- **URL:** `/:id`
- **Method:** `GET`
- **Auth:** Required school member in active `SchoolId` context.
- **Ownership rules:** Enrollment must belong to the active school.
- **Response:** Single enrollment with user and class details
- **Historical field:** `leftAt` appears only when the enrollment has been soft-unenrolled.

## 5. Update Enrollment Role
- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **Ownership rules:** Enrollment must belong to the active school.
- **Body:**
```json
{
  "role": "teacher|student"
}
```
- **Use Case:** Change member role (e.g., promote student to teacher assistant)

## 6. Unenroll Member
- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required (`admin` in active `SchoolId` context)
- **Ownership rules:** Enrollment must belong to the active school.
- **Safety rules:**
  - Student enrollment can be removed without deleting submissions, assessments, grades, materials, assignments, or subject classes.
  - Teacher enrollment cannot be removed while the teacher is still assigned to any `subject_class` in the same class. Unassign the subject teaching assignment first.
- **Delete behavior:** Soft-unenrolls by setting `left_at = now()`. The enrollment row remains for history.
- **Access behavior:** Unenrolled members no longer count as active class members and lose class-derived subject/material/assignment access.
- **Response `409`:** Teacher is still assigned to teach a subject in this class.
- **Note:** Re-enrolling the same school_user to the same class clears `left_at` instead of inserting a duplicate row. The original `joined_at` is preserved.

---

## Features

- **Bulk Enrollment:** Multiple users can be enrolled at once
- **Role Management:** Support for teacher and student roles
- **Bidirectional Queries:** Get by class or by member
- **Class Context:** Enrollment list includes class header
- **History Preservation:** Unenroll preserves the enrollment row and original `joined_at`, and does not delete academic history such as submissions, assessments, or grades.
