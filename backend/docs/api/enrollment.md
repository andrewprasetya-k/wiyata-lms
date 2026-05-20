# 👥 Enrollment Module API Documentation

Base URL: `/api/enrollments`

## 1. Enroll Members to Class
- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**
```json
{
  "schoolId": "uuid",
  "schoolUserIds": ["uuid1", "uuid2"],
  "classId": "uuid",
  "role": "teacher|student"
}
```
- **Note:** Bulk enrollment supported

## 2. Get Enrollments by Class
- **URL:** `/class/:classId`
- **Method:** `GET`
- **Query Params:** `?page=1&limit=20&search=john`
  - `page` (optional): Page number, default 1
  - `limit` (optional): Items per page, default 20
  - `search` (optional): Search by user name or email
- **Response:** `ClassWithMembersDTO`

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
        "joinedAt": "23-02-2026 10:00:00"
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
- **Response:** List of classes the member is enrolled in

## 4. Get Enrollment by ID
- **URL:** `/:id`
- **Method:** `GET`
- **Response:** Single enrollment with user and class details

## 5. Update Enrollment Role
- **URL:** `/:id`
- **Method:** `PATCH`
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
- **Note:** Removes member from class

---

## Features

- **Bulk Enrollment:** Multiple users can be enrolled at once
- **Role Management:** Support for teacher and student roles
- **Bidirectional Queries:** Get by class or by member
- **Class Context:** Enrollment list includes class header
