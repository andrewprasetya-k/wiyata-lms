# 📊 Grade Book API Documentation

Base URL: `/api/grades`

## Overview

Grade Book system untuk manage assessment weights dan calculate provisional weighted grades berdasarkan kategori penilaian.

---

## 1. Configure Assessment Weights
Set bobot penilaian per kategori untuk mata pelajaran.

- **URL:** `/weights`
- **Method:** `POST`
- **Auth:** Required (admin, teacher)
- **Body:**
```json
{
  "subjectId": "uuid-subject",
  "weights": [
    {
      "categoryId": "uuid-quiz",
      "weight": 20.00
    },
    {
      "categoryId": "uuid-uts", 
      "weight": 30.00
    },
    {
      "categoryId": "uuid-uas",
      "weight": 50.00
    }
  ]
}
```

**Validation:**
- Total weight harus = 100.00
- Weight per kategori: 0-100
- Minimum 1 kategori

**Response (200 OK):**
```json
{
  "message": "Weights configured successfully"
}
```

---

## 2. Get Assessment Weights by Subject
Retrieve konfigurasi bobot untuk mata pelajaran.

- **URL:** `/weights/subject/:subjectId`
- **Method:** `GET`
- **Auth:** Required (school member)

**Response (200 OK):**
```json
{
  "subjectId": "uuid",
  "subjectName": "Matematika",
  "subjectCode": "MTK",
  "weights": [
    {
      "weightId": "uuid",
      "categoryId": "uuid",
      "categoryName": "Quiz",
      "weight": 20.00
    },
    {
      "weightId": "uuid", 
      "categoryId": "uuid",
      "categoryName": "UTS",
      "weight": 30.00
    }
  ],
  "totalWeight": 100.00
}
```

---

## 3. Get My Gradebook by Class
Retrieve gradebook current student untuk active class context.

- **URL:** `/my-grades/:classId`
- **Method:** `GET`
- **Auth:** Required (student)
- **Headers:**
```http
Authorization: Bearer <token>
SchoolId: uuid-school-id
```

**Notes:**
- Identity student diambil dari JWT, bukan dari path/body/query.
- `SchoolId` header menentukan active school context.
- Student hanya bisa melihat gradebook untuk class tempat dia ter-enroll sebagai student.
- Response dikelompokkan berdasarkan `subjectClassId`.
- `finalGrade` dan `letterGrade` bernilai `null` jika bobot nilai belum dikonfigurasi atau belum ada nilai yang bisa dihitung.
- Untuk MVP, field `finalGrade` adalah nilai berbobot sementara/provisional. Nilai ini dihitung dari assignment yang sudah dinilai dan kategori yang memiliki bobot tersedia.
- Assignment yang belum dikumpulkan atau sudah dikumpulkan tetapi belum dinilai tidak masuk ke kalkulasi `finalGrade` saat ini.
- `finalGrade` belum berarti nilai rapor/final resmi karena belum ada policy finalisasi term, `max_score`, late penalty, atau rilis nilai resmi.

**Response (200 OK):**
```json
{
  "class": {
    "classId": "uuid",
    "className": "Kelas 10 A",
    "classCode": "10A"
  },
  "subjects": [
    {
      "subjectClassId": "uuid",
      "subjectId": "uuid",
      "subjectName": "Matematika",
      "subjectCode": "MTK",
      "finalGrade": 90,
      "letterGrade": "A",
      "gradedCount": 2,
      "submittedCount": 3,
      "pendingCount": 1,
      "assignments": [
        {
          "assignmentId": "uuid",
          "assignmentTitle": "Quiz Aljabar",
          "categoryName": "Quiz",
          "deadline": "2026-03-01T23:59:59Z",
          "status": "graded",
          "submittedAt": "02-03-2026 10:30:00",
          "score": 90,
          "feedback": "Bagus",
          "assessedAt": "03-03-2026 09:00:00",
          "assessorName": "Nama Guru"
        }
      ]
    }
  ],
  "summary": {
    "subjectCount": 1,
    "gradedAssignmentCount": 2,
    "submittedAssignmentCount": 3,
    "pendingAssessmentCount": 1
  }
}
```

**When student is not enrolled in class (403 Forbidden):**
```json
{
  "error": "Forbidden: student is not enrolled in this class"
}
```

---

## 4. Get Class Grade Report
Retrieve provisional weighted grades untuk seluruh student di kelas untuk mata pelajaran tertentu.

- **URL:** `/class/:classId/subject/:subjectId`
- **Method:** `GET`
- **Auth:** Required (teacher, admin)

**Response (200 OK):**
```json
{
  "class": {
    "classId": "uuid",
    "className": "12 IPA 1",
    "classCode": "12IPA1"
  },
  "subject": {
    "subjectId": "uuid",
    "subjectName": "Matematika",
    "subjectCode": "MTK"
  },
  "students": [
    {
      "studentId": "uuid",
      "studentName": "John Doe",
      "studentEmail": "john@example.com",
      "finalGrade": 82.50,
      "letterGrade": "A"
    },
    {
      "studentId": "uuid",
      "studentName": "Jane Smith", 
      "studentEmail": "jane@example.com",
      "finalGrade": 78.25,
      "letterGrade": "B"
    }
  ]
}
```

---

## Letter Grade Conversion

| Score Range | Letter Grade |
|-------------|--------------|
| 90-100      | A            |
| 80-89       | B            |
| 70-79       | C            |
| 60-69       | D            |
| 0-59        | E            |

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Total weight must be 100, got 95.00"
}
```

### 404 Not Found
```json
{
  "error": "No weights configured for this subject"
}
```

---

## Usage Flow

1. **Admin/Teacher configure weights** per subject
2. **Teacher grade assignments** (via assignment endpoints)
3. **System auto-calculate provisional weighted grades** based on weights and graded assignments
4. **Students view their own gradebook** via `/api/grades/my-grades/:classId`
5. **Teachers view grade reports** with breakdown

---

**Last Updated:** 2026-03-12
