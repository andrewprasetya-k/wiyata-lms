# 📊 Grade Book API Documentation

Base URL: `/api/grades`

## Overview

Grade Book system untuk manage assessment weights dan calculate provisional weighted grades berdasarkan kategori penilaian.

---

## 1. Configure Assessment Weights
Set bobot penilaian per kategori untuk mata pelajaran.

- **URL:** `/weights`
- **Method:** `POST`
- **Auth:** Required (admin only)
- **Scope:** Active school from `SchoolId` context/header
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
- Subject must belong to the active school.
- Every category must belong to the active school.
- Category IDs must be unique in one request.
- Weight per kategori: 0-100
- Total weight harus = 100.00 with small decimal tolerance (±0.01).
- Minimum 1 kategori
- The top-level `subjectId` is the source of truth. Item-level `subjectId` is not required and is ignored for MVP compatibility.
- Assessment weights are subject-level for MVP, not subject_class/class-specific.
- These weights feed the provisional weighted grade, not an official final report grade.

**Transaction Behavior:**
- Weight replacement is **atomic**: the existing weights for the subject are deleted and the new weights are inserted in a single database transaction (`ReplaceBySubject`). A partial failure rolls back entirely — no orphaned weight records are left.

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
- **Scope:** The subject must belong to the active school. Cross-school subject IDs are rejected.

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
          "submittedAt": "2026-03-02T10:30:00Z",
          "score": 90,
          "feedback": "Bagus",
          "assessedAt": "2026-03-03T09:00:00Z",
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

> Untuk drill-down ke satu siswa saja (dengan breakdown kategori dan daftar assignment), lihat **#5 Get Student Grade Detail**. Untuk melihat semua mata pelajaran satu siswa sekaligus, lihat **#6 Get Student Report**.

- **URL:** `/class/:classId/subject/:subjectId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("teacher", "admin")`
- **Ownership:** Service verifies both class and subject belong to the active school. Catatan: otorisasi teacher di sini hanya cek role level sekolah, **bukan** kepemilikan subject_class — beda dengan #5/#6 yang lebih ketat (lihat Authorization Matrix).

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

## 5. Get Student Grade Detail

Retrieve provisional weighted grade detail untuk **satu siswa** pada **satu mata pelajaran** di suatu kelas — drill-down dari Class Grade Report (endpoint #4) ke satu baris siswa, lengkap dengan breakdown per kategori dan daftar assignment.

- **URL:** `/class/:classId/subject/:subjectId/student/:studentId`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("teacher", "admin", "student")`, dengan pemeriksaan tambahan di level handler (lihat Authorization Matrix di bawah)
- **Path Parameters:**
  - `classId` — UUID kelas
  - `subjectId` — UUID mata pelajaran
  - `studentId` — UUID siswa (`users.usr_id`, sama seperti `studentId` di response Class Grade Report)

**Authorization (bukan sekadar cek role):**
- **Admin** — selalu boleh, selama school context valid.
- **Teacher** — boleh **hanya jika benar-benar mengajar mata pelajaran ini di kelas ini** (subject_class ownership check). Role "teacher" saja tidak cukup — ini **lebih ketat** dibanding endpoint #4 (Class Grade Report), yang hanya memeriksa role level sekolah tanpa cek kepemilikan subject_class.
- **Student** — boleh **hanya untuk dirinya sendiri** (`studentId` di URL harus sama dengan ID user yang login). Siswa tidak bisa melihat detail nilai siswa lain.
- Kalau tidak memenuhi salah satu di atas → `403 Forbidden`.

**Response (200 OK):**
```json
{
  "studentId": "b6b8f5b1-8b8b-4b7b-9b7b-1b8b8b8b8b8b",
  "studentName": "Andrew Prasetya",
  "studentEmail": "andrew.prasetya@example.sch.id",
  "class": {
    "classId": "d2f1a9e0-1234-4a5b-8c9d-abcdef123456",
    "classTitle": "12 IPA 1",
    "classCode": "12IPA1"
  },
  "subject": {
    "subjectId": "a1b2c3d4-5678-4abc-9def-0123456789ab",
    "subjectName": "Matematika",
    "subjectCode": "MTK"
  },
  "finalGrade": 82.5,
  "letterGrade": "B",
  "breakdown": [
    {
      "categoryId": "c1d2e3f4-1111-2222-3333-444455556666",
      "categoryName": "Quiz",
      "averageScore": 85,
      "weightedScore": 17,
      "weight": 20,
      "assignmentCount": 2
    },
    {
      "categoryId": "c2d3e4f5-2222-3333-4444-555566667777",
      "categoryName": "UTS",
      "averageScore": 80,
      "weightedScore": 24,
      "weight": 30,
      "assignmentCount": 1
    },
    {
      "categoryId": "c3d4e5f6-3333-4444-5555-666677778888",
      "categoryName": "UAS",
      "averageScore": 83,
      "weightedScore": 41.5,
      "weight": 50,
      "assignmentCount": 1
    }
  ],
  "assignments": [
    {
      "assignmentId": "e1f2a3b4-aaaa-bbbb-cccc-ddddeeeeffff",
      "assignmentTitle": "Quiz Aljabar",
      "categoryName": "Quiz",
      "deadline": "2026-03-01T23:59:59Z",
      "status": "graded",
      "submittedAt": "2026-03-02T10:30:00Z",
      "score": 90,
      "feedback": "Bagus, perhatikan tanda negatif",
      "assessedAt": "2026-03-03T09:00:00Z",
      "assessorName": "Bu Siti Aminah"
    },
    {
      "assignmentId": "e2f3a4b5-bbbb-cccc-dddd-eeeeffff0000",
      "assignmentTitle": "UTS Semester Ganjil",
      "categoryName": "UTS",
      "deadline": "2026-03-10T23:59:59Z",
      "status": "graded",
      "submittedAt": "2026-03-10T14:00:00Z",
      "score": 80,
      "feedback": null,
      "assessedAt": "2026-03-11T08:00:00Z",
      "assessorName": "Bu Siti Aminah"
    }
  ]
}
```

**Error Responses:**

| Status | Kapan terjadi | Body |
|---|---|---|
| `401 Unauthorized` | Tidak login / token invalid | `{"error": "Unauthorized"}` |
| `400 Bad Request` | `classId`/`subjectId`/`studentId` kosong, atau school context tidak ada | `{"error": "classId, subjectId, and studentId are required"}` |
| `403 Forbidden` | Teacher tidak mengajar subject ini di kelas ini, atau student mencoba lihat siswa lain | `{"error": "Forbidden: you cannot access this student's grade detail"}` |
| `403 Forbidden` | Siswa target tidak ter-enroll di kelas tersebut | `{"error": "Forbidden: student is not enrolled in this class"}` |
| `404 Not Found` | `subjectId` tidak ditemukan, atau belum ada bobot penilaian dikonfigurasi untuk subject ini | `{"error": "The requested data was not found"}` atau `{"error": "No weights configured for this subject"}` |

---

## 6. Get Student Report

Retrieve rapor lengkap **satu siswa** untuk **seluruh mata pelajaran** yang diambil di satu kelas — agregasi lintas subject, bukan pengganti endpoint #4/#5/#3 (Class Grade Report, Student Grade Detail, My Gradebook), melainkan endpoint tambahan untuk kebutuhan "lihat semua nilai satu siswa sekaligus".

- **URL:** `/class/:classId/student/:studentId/report`
- **Method:** `GET`
- **Auth:** `RequireSchoolMember + RequireRole("teacher", "admin", "student")`, dengan pemeriksaan tambahan di level handler (lihat Authorization Matrix di bawah)
- **Path Parameters:**
  - `classId` — UUID kelas
  - `studentId` — UUID siswa (`users.usr_id`)

**Authorization (bukan sekadar cek role):**
- **Admin** — selalu boleh, selama school context valid.
- **Teacher** — boleh **hanya jika mengajar minimal satu mata pelajaran di kelas ini** (tidak harus semua mata pelajaran yang muncul di report). Ini beda dengan endpoint #5: karena report ini mencakup semua subject, syaratnya "mengajar salah satu", bukan "mengajar subject yang sedang dilihat".
- **Student** — boleh **hanya untuk dirinya sendiri** (`studentId` di URL harus sama dengan ID user yang login).
- Kalau tidak memenuhi salah satu di atas → `403 Forbidden`.

**Notes:**
- `subjects[]` mencakup semua mata pelajaran yang diajarkan di kelas tersebut untuk siswa ini (berdasarkan data assignment/submission yang ada), bukan hanya yang sudah dikonfigurasi bobotnya — subject yang belum punya bobot penilaian akan **tidak muncul** di `subjects[]` (dilewati secara diam-diam, konsisten dengan perilaku endpoint #4).
- `summary.averageFinalGrade` dihitung dari rata-rata `finalGrade` subject yang berhasil dihitung saja (subject yang di-skip karena belum ada bobot tidak ikut memengaruhi rata-rata).
- **Tidak ada konsep "lulus"/"tidak lulus" (passing grade/KKM) di project ini** — summary sengaja tidak berisi field semacam itu karena akan mengarang data yang tidak ada dasarnya di sistem.

**Response (200 OK):**
```json
{
  "studentId": "b6b8f5b1-8b8b-4b7b-9b7b-1b8b8b8b8b8b",
  "studentName": "Andrew Prasetya",
  "studentEmail": "andrew.prasetya@example.sch.id",
  "class": {
    "classId": "d2f1a9e0-1234-4a5b-8c9d-abcdef123456",
    "classTitle": "12 IPA 1",
    "classCode": "12IPA1"
  },
  "subjects": [
    {
      "subject": {
        "subjectId": "a1b2c3d4-5678-4abc-9def-0123456789ab",
        "subjectName": "Matematika",
        "subjectCode": "MTK"
      },
      "finalGrade": 82.5,
      "letterGrade": "B",
      "breakdown": [
        {
          "categoryId": "c1d2e3f4-1111-2222-3333-444455556666",
          "categoryName": "Quiz",
          "averageScore": 85,
          "weightedScore": 17,
          "weight": 20,
          "assignmentCount": 2
        }
      ],
      "assignments": [
        {
          "assignmentId": "e1f2a3b4-aaaa-bbbb-cccc-ddddeeeeffff",
          "assignmentTitle": "Quiz Aljabar",
          "categoryName": "Quiz",
          "deadline": "2026-03-01T23:59:59Z",
          "status": "graded",
          "submittedAt": "2026-03-02T10:30:00Z",
          "score": 90,
          "feedback": "Bagus",
          "assessedAt": "2026-03-03T09:00:00Z",
          "assessorName": "Bu Siti Aminah"
        }
      ]
    },
    {
      "subject": {
        "subjectId": "f1e2d3c4-9876-4cba-fedc-ba9876543210",
        "subjectName": "Fisika",
        "subjectCode": "FIS"
      },
      "finalGrade": 88,
      "letterGrade": "B",
      "breakdown": [],
      "assignments": []
    }
  ],
  "summary": {
    "totalSubjects": 2,
    "averageFinalGrade": 85.25
  }
}
```

**Error Responses:**

| Status | Kapan terjadi | Body |
|---|---|---|
| `401 Unauthorized` | Tidak login / token invalid | `{"error": "Unauthorized"}` |
| `400 Bad Request` | `classId`/`studentId` kosong, atau school context tidak ada | `{"error": "classId and studentId are required"}` |
| `403 Forbidden` | Teacher tidak mengajar mata pelajaran apa pun di kelas ini, atau student mencoba lihat siswa lain | `{"error": "Forbidden: you cannot access this student's report"}` |
| `403 Forbidden` | Siswa target tidak ter-enroll di kelas tersebut | `{"error": "Forbidden: student is not enrolled in this class"}` |
| `404 Not Found` | Data terkait tidak ditemukan | `{"error": "The requested data was not found"}` |

---

## Authorization Matrix (Endpoint #5 & #6)

| Role | Get Student Grade Detail (#5) | Get Student Report (#6) |
|---|---|---|
| Admin | ✅ | ✅ |
| Teacher yang mengajar subject/kelas terkait | ✅ | ✅ (cukup mengajar 1 subject di kelas ini) |
| Teacher lain (tidak mengajar kelas/subject ini) | ❌ | ❌ |
| Student (dirinya sendiri) | ✅ | ✅ |
| Student lain | ❌ | ❌ |

---

## DTO Reference (Endpoint #5 & #6)

Kedua endpoint baru ini menyusun response dari kombinasi DTO yang sudah ada, bukan mendefinisikan ulang struktur breakdown/assignment dari nol:

- **`StudentGradeDetailDTO`** (response endpoint #5): `studentId`, `studentName`, `studentEmail` (string) + `class` (reuse `ClassHeaderDTO`) + `subject` (reuse `SubjectHeaderDTO`) + `finalGrade` (float64) + `letterGrade` (string) + `breakdown` (reuse `[]CategoryBreakdownDTO`, sama seperti yang dipakai internal `CalculateFinalGrade`) + `assignments` (reuse `[]MyGradebookAssignmentDTO`, struktur sama persis dengan `assignments[]` di endpoint My Gradebook #3).
- **`StudentReportDTO`** (response endpoint #6): `studentId`, `studentName`, `studentEmail` (string) + `class` (reuse `ClassHeaderDTO`) + `subjects` (list `StudentReportSubjectDTO`) + `summary` (`StudentReportSummaryDTO`).
- **`StudentReportSubjectDTO`** (satu entry di `subjects[]`): `subject` (reuse `SubjectHeaderDTO`) + `finalGrade`/`letterGrade` + `breakdown` (reuse `[]CategoryBreakdownDTO`) + `assignments` (reuse `[]MyGradebookAssignmentDTO`).
- **`StudentReportSummaryDTO`**: `totalSubjects` (int) — jumlah subject yang berhasil dihitung nilainya; `averageFinalGrade` (float64) — rata-rata `finalGrade` dari subject-subject tersebut. Tidak ada field jumlah "lulus"/"belum lulus" — project ini tidak punya konsep passing grade/KKM di manapun di domain model, jadi field tersebut sengaja tidak dibuat.

DTO yang di-reuse (didefinisikan di tempat lain, tidak diulang di sini): `ClassHeaderDTO` (`classId`, `classTitle`, `classCode`), `SubjectHeaderDTO` (`subjectId`, `subjectName`, `subjectCode`, `subjectColor` opsional), `CategoryBreakdownDTO` (`categoryId`, `categoryName`, `averageScore`, `weightedScore`, `weight`, `assignmentCount`), `MyGradebookAssignmentDTO` (`assignmentId`, `assignmentTitle`, `categoryName`, `deadline`, `status`, `submittedAt`, `score`, `feedback`, `assessedAt`, `assessorName` — lihat definisi lengkap di section #3 My Gradebook di atas).

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

1. **Admin configures weights** per subject
2. **Teacher grade assignments** (via assignment endpoints)
3. **System auto-calculate provisional weighted grades** based on weights and graded assignments
4. **Students view their own gradebook** via `/api/grades/my-grades/:classId`
5. **Teachers/admin view grade reports** with breakdown — either for a whole class (`/api/grades/class/:classId/subject/:subjectId`), one student's detail on one subject (`/api/grades/class/:classId/subject/:subjectId/student/:studentId`), or one student's full report across every subject in the class (`/api/grades/class/:classId/student/:studentId/report`)
6. **Students can also view their own grade detail/report** via the same endpoints in step 5 (self-access only — see Authorization Matrix)

---

**Last Updated:** 2026-07-14
