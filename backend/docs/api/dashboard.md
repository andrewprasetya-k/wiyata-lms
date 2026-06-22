# 📊 Dashboard API Documentation

Base URL: `/api/dashboard`

## 1. Student Dashboard
Get dashboard statistics for a student.

- **URL:** `/student/:userId`
- **Method:** `GET`
- **Response:**
```json
{
  "pendingAssignments": 5,
  "upcomingDeadlines": [
    {
      "assignmentId": "uuid",
      "assignmentTitle": "Quiz Chapter 1",
      "subjectName": "Matematika",
      "deadline": "01-03-2026 23:59",
      "isSubmitted": false
    }
  ],
  "averageScore": 85.5,
  "completedMaterials": 12,
  "totalMaterials": 20
}
```

**Metrics:**
- `pendingAssignments`: Number of assignments not yet submitted
- `upcomingDeadlines`: Next 5 assignments ordered by deadline
- `averageScore`: Average score across all graded submissions
- `completedMaterials`: Number of materials marked as completed
- `totalMaterials`: Total materials available to the student

---

## 2. Teacher Dashboard
Get dashboard statistics for a teacher.

- **URL:** `/teacher/:schoolUserId`
- **Method:** `GET`
- **Note:** `schoolUserId` is the `school_user_id` (scu_id), NOT user_id
- **Response:**
```json
{
  "pendingReviews": 15,
  "totalStudents": 120,
  "submissionRate": 78.5,
  "classPerformance": [
    {
      "classId": "uuid",
      "className": "12 IPA 1",
      "subjectName": "Matematika",
      "averageScore": 82.3,
      "submissionRate": 85.0,
      "totalStudents": 30
    }
  ]
}
```

**Metrics:**
- `pendingReviews`: Number of submissions waiting for grading
- `totalStudents`: Total unique students across all teacher's classes
- `submissionRate`: Overall submission rate across all assignments (%), calculated as submitted active-student assignment slots divided by total eligible active-student assignment slots. Active students are class enrollments with `left_at IS NULL`.
- `classPerformance`: Performance breakdown per class/subject
- `classPerformance[].submissionRate`: Per class/subject submission rate using the same eligible active-student assignment slot denominator.

---

## 3. Admin Dashboard
Get dashboard statistics for school admin.

- **URL:** `/admin/:schoolId`
- **Method:** `GET`
- **Response:**
```json
{
  "totalStudents": 450,
  "totalTeachers": 35,
  "totalClasses": 15,
  "activeClasses": 12,
  "enrollmentTrends": [
    {
      "className": "12 IPA 1",
      "totalEnrolled": 32,
      "teachers": 2,
      "students": 30
    }
  ],
  "recentActivities": [
    {
      "userName": "John Doe",
      "action": "Created new assignment",
      "timestamp": "24-02-2026 10:30:00"
    }
  ]
}
```

**Metrics:**
- `totalStudents`: Total students enrolled in the school
- `totalTeachers`: Total teachers assigned to classes
- `totalClasses`: Total classes (including inactive)
- `activeClasses`: Only active classes
- `enrollmentTrends`: Enrollment breakdown per class
- `recentActivities`: Last 10 activities from logs

---

## Key Features

### Real-time Calculations
All metrics are calculated in real-time from the database:
- No caching required
- Always up-to-date data
- Efficient SQL queries with proper joins

### Role-based Views
Each dashboard is tailored to the user's role:
- **Student**: Focus on personal progress and deadlines
- **Teacher**: Focus on grading workload and class performance
- **Admin**: Focus on school-wide statistics and trends

### Performance Optimized
- Uses aggregated queries to minimize database load
- Limits on list results (e.g., top 5 deadlines, last 10 activities)
- Indexed columns for fast lookups

---

## Usage Examples

### Student checking their dashboard
```bash
GET /api/dashboard/student/123e4567-e89b-12d3-a456-426614174000
# Parameter: userId (usr_id from users table)
```

### Teacher viewing pending reviews
```bash
GET /api/dashboard/teacher/223e4567-e89b-12d3-a456-426614174000
# Parameter: schoolUserId (scu_id from school_users table)
```

### Admin monitoring school statistics
```bash
GET /api/dashboard/admin/323e4567-e89b-12d3-a456-426614174000
# Parameter: schoolId (sch_id from schools table)
```

---

## Notes

- **Authentication Required**: These endpoints should be protected with JWT middleware
- **Authorization**: Ensure users can only access their own dashboard (or admin can access school dashboard)
- **Caching**: Consider adding Redis caching for admin dashboard if school is large
- **Pagination**: Currently returns fixed limits (5 deadlines, 10 activities). Can be made configurable.
