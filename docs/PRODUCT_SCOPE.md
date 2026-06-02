# Eduverse LMS — Product Scope Locked Version

## 1. Product Vision

Eduverse LMS adalah platform learning management system multi-sekolah yang membantu sekolah mengelola proses pembelajaran digital secara terstruktur, mulai dari setup akademik, kelas, mata pelajaran, materi, tugas, submission, penilaian, komunikasi kelas, notifikasi, media, hingga fitur kolaborasi lanjutan seperti chat dan student notes.

Arah utama produk adalah membuat pengalaman belajar harian yang jelas dan nyaman untuk siswa, sekaligus tetap kuat untuk kebutuhan guru, admin sekolah, dan super admin.

Eduverse bukan sekadar admin dashboard sekolah. Eduverse adalah academic workspace yang menggabungkan:

* struktur akademik sekolah,
* aktivitas belajar per mata pelajaran,
* komunikasi kelas,
* pengumpulan tugas,
* penilaian,
* notifikasi,
* media pembelajaran,
* dan personal learning notes.

---

## 2. Core Product Principle

Prinsip domain utama Eduverse:

```text
Class = academic context
Subject Class = daily learning workspace
Feed = class-level communication
Chat = future realtime collaboration
Notes = personal learning workspace per material
```

Keputusan ini menjadi dasar routing, UI, dan proses bisnis frontend/backend.

---

## 3. Target Users

### 3.1 Super Admin

Super Admin mengelola level platform dan setup sekolah. Super Admin tidak otomatis menjadi aktor akademik dalam sekolah kecuali memiliki role school-level tertentu.

Tanggung jawab utama:

* mengelola school tenant,
* melihat konfigurasi platform,
* memonitor kondisi sistem,
* membantu setup awal sekolah.

### 3.2 Admin Sekolah

Admin sekolah mengelola struktur akademik dan operasional sekolah.

Tanggung jawab utama:

* academic year,
* term,
* subject,
* class,
* subject class,
* user membership,
* role assignment,
* enrollment siswa/guru.

### 3.3 Teacher

Teacher mengelola pembelajaran dalam subject class yang dia ampu.

Tanggung jawab utama:

* membuat material,
* membuat assignment,
* melihat submission,
* memberi assessment/feedback,
* membuat feed/pengumuman level class,
* berkomunikasi melalui fitur chat nanti.

### 3.4 Student

Student adalah pengguna harian utama aplikasi.

Tanggung jawab utama:

* memilih active class,
* melihat subject yang diikuti,
* membuka material,
* mengerjakan assignment,
* melihat nilai dan feedback,
* membaca feed kelas,
* menerima notifikasi,
* membuat catatan pribadi per material nanti.

---

## 4. Domain Model Summary

### 4.1 School

School adalah tenant utama Eduverse. Semua data akademik seperti academic year, term, subject, class, material, feed, media, dan notification berada dalam konteks school.

### 4.2 User vs School User

`users` adalah identitas global.

`school_users` adalah membership user dalam sebuah school.

Role tidak langsung melekat ke user global. Role melekat ke school membership melalui:

```text
users → school_users → user_roles → roles
```

Implikasi:

* satu user bisa berada di banyak school,
* satu user bisa punya role berbeda di school berbeda,
* frontend harus memiliki active school context,
* request school-scoped harus mengirim `SchoolId` header.

### 4.3 Academic Year dan Term

Academic year dan term menentukan periode akademik. Class berada dalam term tertentu.

### 4.4 Class

Class adalah konteks akademik siswa dalam satu term.

Contoh:

```text
XII IPA 1 - Semester Ganjil
```

Class digunakan untuk:

* enrollment siswa/guru,
* feed kelas,
* daftar subject class,
* active class context di frontend.

Class bukan pusat materi/tugas harian.

### 4.5 Subject

Subject adalah master mata pelajaran di school.

Contoh:

```text
Matematika
Fisika
Bahasa Indonesia
```

### 4.6 Subject Class

Subject Class adalah kombinasi antara class, subject, dan guru pengampu.

Contoh:

```text
XII IPA 1 + Matematika + Pak Budi
```

Subject Class adalah daily learning workspace.

Material dan assignment selalu berada di subject class.

```text
subject_classes
├── materials
├── assignments
└── subject chat room nanti
```

### 4.7 Enrollment

Enrollment menentukan siapa saja anggota sebuah class.

```text
school_user → enrollment → class
```

Enrollment memiliki role class:

* teacher,
* student.

Catatan penting:

* enrollment menunjukkan anggota class,
* `subject_classes.scl_scu_id` menunjukkan guru pengampu subject tertentu.

---

## 5. Main Business Workflows

## 5.1 School Setup Flow

Aktor utama: Super Admin / Admin Sekolah

Alur:

```text
Create school
→ create academic year
→ create term
→ create subjects
→ create classes
→ create subject classes
→ create/enroll users
→ assign roles
→ enroll students/teachers to class
```

Tujuan:

* membentuk struktur akademik dasar,
* memastikan siswa masuk ke class,
* memastikan subject class punya guru pengampu.

---

## 5.2 Student Daily Learning Flow

Aktor utama: Student

Alur:

```text
Login
→ pilih active school
→ pilih active class
→ lihat subject dalam active class
→ buka subject
→ lihat materials dan assignments
→ buka material / submit assignment
→ lihat feedback dan nilai
→ baca feed kelas
→ cek notification
```

UI utama student harus subject-centric.

Routing utama:

```text
/student/dashboard
/student/subjects
/student/subjects/:sclId
/student/subjects/:sclId/materials/:matId
/student/subjects/:sclId/assignments/:asgId
/student/feed
/student/grades
/student/chat
/student/notes
```

---

## 5.3 Teacher Learning Management Flow

Aktor utama: Teacher

Alur:

```text
Login
→ lihat subject class yang diajar
→ buat material pada subject class
→ buat assignment pada subject class
→ lihat submission siswa
→ beri assessment dan feedback
→ buat feed/pengumuman class
→ cek notification/activity
```

Teacher bekerja berdasarkan subject class yang dia ampu.

---

## 5.4 Material Flow

Aktor utama: Teacher dan Student

Alur:

```text
Teacher membuat material
→ material tersimpan pada subject_class
→ teacher attach file melalui media/attachments
→ student membuka material
→ material_progress tercatat per user-material
→ student membuat notes pribadi per material nanti
```

Material selalu terkait ke:

```text
materials.mat_scl_id → subject_classes.scl_id
```

Material progress:

```text
material_progress.map_usr_id + material_progress.map_mat_id
```

Student notes:

```text
student_notes.snt_usr_id + student_notes.snt_mat_id
```

Keputusan penting:

* membuka material tidak otomatis berarti completed,
* jika ingin passive tracking, update `last_opened_at`,
* status `completed` harus berdasarkan aksi eksplisit atau rule yang jelas.

---

## 5.5 Assignment Flow

Aktor utama: Teacher dan Student

Alur:

```text
Teacher membuat assignment pada subject_class
→ student melihat assignment di subject detail
→ student submit assignment
→ teacher assess submission
→ student melihat score dan feedback
→ notification dikirim
```

Relasi inti:

```text
assignments.asg_scl_id → subject_classes.scl_id
submissions.sbm_asg_id → assignments.asg_id
submissions.sbm_usr_id → users.usr_id
assessments.asm_sbm_id → submissions.sbm_id
```

Constraint:

```text
1 student hanya punya 1 submission per assignment
```

UI status assignment student:

```text
Tidak ada submission → Belum dikerjakan
Ada submission, belum ada assessment → Menunggu penilaian
Ada assessment → tampilkan score dan feedback
Submit telat jika allowed_late true → tampilkan badge Terlambat
```

---

## 5.6 Grade Book Flow

Aktor utama: Student dan Teacher

Grade book dihitung dari:

```text
subjects
assignment_categories
assessments_weights
assignments
submissions
assessments
```

Bobot nilai ada per subject dan assignment category.

```text
assessments_weights.asw_sub_id
assessments_weights.asw_asc_id
```

Tampilan grade book sebaiknya subject-based.

---

## 5.7 Feed Flow

Aktor utama: Teacher dan Student

Feed adalah komunikasi level class.

Relasi:

```text
feeds.fds_cls_id → classes.cls_id
```

Feed digunakan untuk:

* pengumuman kelas,
* komunikasi umum lintas subject dalam class,
* diskusi melalui comments,
* attachment jika diperlukan.

Keputusan penting:

* Feed bukan level subject,
* Feed bukan chat realtime,
* Feed adalah REST-based class communication.

Feed type badge seperti Tugas/Materi/Pengumuman tidak diprioritaskan sekarang karena tidak ada field eksplisit di schema.

Future TODO:

* auto-create feed post ketika teacher membuat material,
* auto-create feed post ketika teacher membuat assignment.

---

## 5.8 Comment Flow

Comment bersifat polymorphic.

Relasi:

```text
comments.cmn_source_type
comments.cmn_source_id
```

Comment bisa menempel ke:

* material,
* assignment,
* feed,
* submission,
* comment.

Frontend harus selalu memahami konteks source sebelum render comment.

---

## 5.9 Media and Attachment Flow

Media adalah storage metadata terpusat.

File disimpan di:

```text
medias
```

Lalu ditempel ke konten melalui:

```text
attachments.att_source_type
attachments.att_source_id
attachments.att_med_id
```

Media bisa digunakan untuk:

* material attachment,
* assignment attachment,
* submission attachment,
* feed attachment,
* chat attachment nanti.

Prinsip penting:

* storage upload harus real, bukan fake URL,
* metadata disimpan setelah upload sukses,
* delete media harus mempertimbangkan cleanup storage.

---

## 5.10 Notification Flow

Notification adalah activity signal untuk user.

Relasi utama:

```text
notifications.ntf_usr_id → users.usr_id
notifications.ntf_related_id → konten terkait
```

Notification trigger yang didukung:

* assignment created,
* assignment graded,
* material added,
* feed posted,
* comment added.

Notification bersifat best-effort.

Aksi utama tidak boleh gagal hanya karena notifikasi gagal.

---

## 5.11 Chat Flow — Future Feature

Schema mendukung:

* subject chat,
* direct message,
* group chat,
* class chat.

Keputusan produk saat ini:

* class chat tidak digunakan di UI utama,
* class-level communication cukup melalui Feed,
* subject chat boleh digunakan nanti,
* DM dan group chat tetap dimungkinkan,
* chat belum masuk implementasi MVP frontend saat ini.

Chat menggunakan:

```text
chat_rooms
chat_room_members
chat_messages
chat_attachments
chat_read_receipts
```

Room type yang relevan untuk UI nanti:

* subject,
* dm,
* group.

Class room tidak diprioritaskan untuk UI.

---

## 5.12 Student Notes Flow — Future Feature

Student notes adalah catatan pribadi per material.

Relasi:

```text
student_notes.snt_usr_id → users.usr_id
student_notes.snt_mat_id → materials.mat_id
```

Constraint:

```text
1 student hanya punya 1 note per material
```

Notes akan muncul di Material Detail, bukan sebagai fitur utama terpisah terlebih dahulu.

Notes belum diimplementasikan penuh di frontend/backend saat ini.

---

## 6. Frontend Product Scope

## 6.1 Global Frontend Context

Frontend harus menjaga context berikut:

```text
activeSchoolId
activeSchoolUserId
activeRoles
activeClassId
```

`activeSchoolId` berasal dari login membership/defaultContext.

`activeClassId` berasal dari enrollment siswa dalam term aktif.

Jika user punya lebih dari satu class aktif, frontend perlu class switcher.

---

## 6.2 Student Navigation Scope

Student navigation final:

```text
Dashboard
Subjects
Feed
Assignments
Grades
Chat
Notes
Profile
```

### Dashboard

Dashboard adalah ringkasan aktivitas, bukan pusat CRUD.

Dashboard boleh menampilkan:

* active class context,
* subject shortcut,
* tugas pending,
* notification/recent activity,
* feed snippet,
* placeholder chat preview jika chat belum siap.

### Subjects

Subjects adalah halaman utama aktivitas belajar harian.

```text
/student/subjects
```

Menampilkan subject class dalam active class.

### Subject Detail

```text
/student/subjects/:sclId
```

Tabs:

* Materi,
* Tugas,
* Catatan.

Tidak ada Feed tab.

Tidak ada Chat tab.

Chat subject boleh berupa button placeholder di topbar.

### Feed

```text
/student/feed
```

Feed adalah komunikasi level class berdasarkan active class.

### Assignments

```text
/student/assignments
```

Halaman agregat semua tugas dari subject dalam active class.

### Grades

```text
/student/grades
```

Grade book per subject.

### Chat

```text
/student/chat
```

Placeholder untuk sekarang.

Future: subject chat, DM, group.

### Notes

```text
/student/notes
```

Placeholder untuk sekarang.

Future: daftar catatan pribadi lintas material.

---

## 6.3 Teacher Navigation Scope

Teacher navigation utama:

```text
Dashboard
Subjects / Teaching
Assignments
Submissions
Create Content
Feed
Chat
Profile
```

Teacher bekerja berdasarkan subject class yang dia ampu.

---

## 6.4 Admin Navigation Scope

Admin navigation utama:

```text
Dashboard
Academic Years
Terms
Subjects
Classes
Subject Classes
Users
Enrollments
Roles
Profile
```

Admin fokus ke struktur akademik dan membership.

---

## 6.5 Super Admin Navigation Scope

Super Admin navigation utama:

```text
Dashboard
Schools
Users
Platform Settings
Profile
```

Super Admin fokus ke platform-level management.

---

## 7. Backend Contract Priorities

Agar frontend tidak terlalu banyak melakukan request kecil, backend sebaiknya nanti menyediakan endpoint agregat berikut.

### 7.1 Active Class / Student Context

Endpoint ideal:

```text
GET /api/me/context
```

Response:

```text
active school
memberships
active/enrolled classes
active term
candidate active class
```

### 7.2 Subjects by Active Class

Endpoint ideal:

```text
GET /api/classes/:classId/subjects
```

atau existing:

```text
GET /api/subject-classes/class/:classId
```

### 7.3 Materials by Subject Class

Existing:

```text
GET /api/materials?subjectClassId=:sclId
```

### 7.4 Assignments by Subject Class

Ideal:

```text
GET /api/assignments?subjectClassId=:sclId
```

atau endpoint khusus jika sudah ada.

### 7.5 Student Dashboard Aggregate

Dashboard sebaiknya memakai endpoint agregat agar tidak perlu banyak request:

```text
GET /api/dashboard/student/:userId
```

Jika endpoint ini belum cukup, perlu diperluas sesuai kebutuhan UI.

### 7.6 Student Subject Detail Aggregate

Future improvement:

```text
GET /api/student/subjects/:sclId/overview
```

Berisi:

* subject info,
* teacher,
* material count,
* completed material count,
* pending assignments,
* latest activity.

---

## 8. Explicit Out of Scope for Current MVP

Fitur berikut tidak diimplementasikan sekarang dan harus tampil sebagai placeholder/TODO jika muncul di frontend:

* realtime chat WebSocket,
* notes autosave,
* rich text editor,
* nested comments,
* transcript/export,
* advanced analytics,
* class chat UI,
* feed type badge,
* notification preferences,
* email delivery,
* thumbnail generation,
* signed/private download URL.

---

## 9. Current Development Priority

Urutan sehat setelah scope ini:

```text
1. Active class context/store
2. Student subjects list
3. Student subject detail — materials tab
4. Student subject detail — assignments tab
5. Student feed page
6. Student assignment detail/submit flow
7. Student material detail
8. Teacher dashboard basics
9. Teacher create material/assignment
10. Teacher grading flow
11. Admin academic setup pages
12. Chat and notes later
```

---

## 10. Non-Negotiable Product Decisions

Keputusan berikut dianggap fixed sampai ada alasan besar untuk mengubahnya:

1. Class adalah konteks, bukan pusat konten harian.
2. Subject Class adalah pusat aktivitas belajar harian.
3. Material dan assignment selalu berada di subject class.
4. Feed berada di level class.
5. Chat tidak menggantikan feed.
6. Class chat tidak dipakai di UI utama.
7. DM dan group chat tetap dimungkinkan sebagai fitur masa depan.
8. Notes adalah catatan pribadi per material per student.
9. Frontend tidak boleh menampilkan fake data sebagai real data.
10. Fitur yang belum siap harus tampil sebagai placeholder/TODO.
11. Actor identity berasal dari JWT, bukan body request.
12. School context dikirim via `SchoolId` header.
13. Backend authorization tetap source of truth; frontend guard hanya untuk UX.

---

## 11. Summary

Eduverse LMS adalah LMS multi-school dengan struktur akademik yang jelas:

```text
School → Term → Class → Subject Class → Materials / Assignments
```

Model mental utama:

```text
Class = context
Subject = daily workspace
Feed = class communication
Chat = future realtime collaboration
Notes = personal learning workspace
```

Scope ini harus menjadi acuan utama agar frontend, backend, schema, dan product direction tidak berubah-ubah di tengah development.

in short:
School adalah tenant.

Class adalah konteks akademik siswa dalam term tertentu.

Subject Class adalah ruang belajar harian:

- materi
- tugas
- nilai
- chat subject nanti

Feed adalah komunikasi umum level class.

Material dan assignment selalu hidup di subject.

Submission dan assessment adalah lifecycle tugas siswa.

Notes adalah ruang pribadi siswa per material.

Media adalah file storage terpusat.

Notifications adalah activity signal lintas fitur.