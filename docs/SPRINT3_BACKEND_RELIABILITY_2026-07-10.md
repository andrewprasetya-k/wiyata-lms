# Sprint 3 — Backend Reliability & Performance Hardening

**Tanggal:** 2026-07-10
**Scope:** Assignment transaction atomicity, notification async, RBAC query deduplication, composite index — tanpa mengubah API contract, DTO, endpoint, atau business logic.

---

## 1. Audit Summary

### A. Assignment Transaction

Dikonfirmasi berdasarkan pembacaan kode aktual (`internal/service/assignment_service.go`) — pola `mutate entity → attachment link/cleanup` tersebar di 6 fungsi, **tidak satu pun** berada dalam satu `db.Transaction()`:

| Fungsi | File:Line (sebelum) | Pola |
|---|---|---|
| `CreateAssignment` | `assignment_service.go:81-87` | `repo.CreateAssignment` lalu `replaceSourceAttachments` — 2 operasi terpisah |
| `UpdateAssignment` | `assignment_service.go:350-359` | `repo.UpdateAssignment` lalu `replaceSourceAttachments` |
| `Submit` | `assignment_service.go:398-404` | `repo.UpsertSubmission` lalu `replaceSourceAttachments` |
| `UpdateSubmission` | `assignment_service.go:485-495` | pola identik dengan `Submit` |
| `DeleteAssignment` | `assignment_service.go:372-373` | `attService.UnlinkBySource` (error **dibuang diam-diam**) lalu `repo.DeleteAssignment` |
| `DeleteSubmission` | `assignment_service.go:499-500` | pola identik dengan `DeleteAssignment`, error juga dibuang |

**Assessment flow** (`Assess`, `UpdateAssessment`, `DeleteAssessment`) — dikonfirmasi **tidak melibatkan attachment sama sekali** (tidak ada pemanggilan `attService` di ketiganya). Tidak ada yang perlu diperbaiki di area ini.

`replaceSourceAttachments` sendiri (lewat `AttachmentRepository.ReplaceBySource`) sudah membungkus delete+insert attachment dalam transaksinya sendiri — tapi transaksi itu **terpisah** dari transaksi (atau ketiadaan transaksi) milik operasi assignment/submission-nya.

### B. Notification Fan-out

4 lokasi ditemukan dengan pola identik (insert notification satu per satu secara synchronous di request path, semuanya sudah diberi komentar "Best-effort" oleh developer sebelumnya — menunjukkan kesadaran akan masalah ini tapi belum diimplementasikan sebagai non-blocking):

| File:Line | Jumlah insert | Trigger |
|---|---|---|
| `comment_service.go:60` (`notifyCommentRecipients`) | 1 s.d. N (semua murid di kelas atau 1 pemilik konten) | Setiap komentar baru |
| `assignment_service.go:92-104` (`CreateAssignment`) | N (semua murid di kelas) | Setiap tugas baru dibuat |
| `feed_service.go:77-101` (`Create` feed) | N (semua anggota kelas) | Setiap pengumuman baru |
| `material_service.go:129-143` (`Create` material) | N (semua murid di kelas) | Setiap materi baru |

Semua synchronous, semua terjadi sebelum `return nil` ke handler → memperpanjang response time proporsional terhadap jumlah murid di kelas.

### C. RBAC Middleware

Dikonfirmasi di `internal/middleware/rbac_middleware.go`: `RequireRole` (baris ~183, sebelum perbaikan) **selalu** memanggil `rbacRepo.GetUserRoleNamesInSchool(userID, schoolID)` tanpa syarat — termasuk saat `RequireSchoolMember` **sudah** memanggil query yang **identik** beberapa baris sebelumnya di request yang sama.

Redundansi ini **hanya** terjadi pada skenario spesifik: request menyertakan header `Active-Role` (dipakai user multi-role untuk memilih peran aktifnya). Pada skenario tanpa `Active-Role` header (mayoritas request), `RequireSchoolMember` tidak pernah mengambil roles sama sekali (hanya cek membership), sehingga `RequireRole` di sana BUKAN redundansi — itu query pertama dan satu-satunya yang dibutuhkan.

Membership check (`IsUserInSchool`, `GetSchoolUserID`) tidak diulang oleh `RequireRole` — redundansi murni terbatas pada `GetUserRoleNamesInSchool`.

### D. Composite Index

Dikonfirmasi: `domain.Attachment` dan `domain.Comment` **tidak punya index apa pun** pada kolom `source_type`/`source_id` (hanya `DeletedAt` yang punya tag `gorm:"index"` di seluruh domain layer, sebagai perbandingan).

Lokasi query yang memakai pola `WHERE source_type = ? AND source_id = ?`:

| Table | File:Function | 
|---|---|
| attachments | `attachment_repo.go:GetBySource` (baris 29-33) |
| attachments | `attachment_repo.go:GetBySources` (baris 36-48, pakai `IN`) |
| attachments | `attachment_repo.go:DeleteBySource` (baris 63-66) |
| attachments | `attachment_repo.go:ReplaceBySource` (baris 69-90, delete step) |
| comments | `comment_repo.go:GetBySourceInSchool` (baris 28-33) |
| comments | `comment_repo.go:CountBySourceInSchool` (baris 59-64) |

Kedua tabel ini dibaca di HAMPIR SETIAP request GET material/assignment/feed detail (untuk menampilkan attachment & comment count), jadi full-table-scan di sini berdampak langsung ke latency halaman yang paling sering diakses.

**Catatan penting:** project ini tidak memakai migration tool atau `AutoMigrate` (dikonfirmasi Sprint sebelumnya) — index GORM tag di domain struct dan index block di `schema.md` **tidak otomatis membuat index di database sungguhan**. SQL manual untuk dijalankan disertakan di bagian 4 di bawah.

---

## 2. Root Cause

- **A**: Tidak ada kebiasaan/pola transaksi di layer service untuk operasi lintas-repository (assignment+attachment). Masing-masing repository method aman secara individual, tapi komposisinya di service layer tidak atomic.
- **B**: Notifikasi diimplementasikan sebagai pemanggilan langsung `notifService.Create()` di dalam alur request yang sama, bukan didesain untuk berjalan di luar request lifecycle sejak awal — komentar "Best-effort" di kode menunjukkan niat tersebut ada tapi belum direalisasikan.
- **C**: `RequireRole` didesain untuk bisa berdiri sendiri (dipakai tanpa `RequireSchoolMember` di beberapa route seperti `schools`/`users`), sehingga selalu melakukan query sendiri secara defensif — tapi tidak ada mekanisme untuk mendeteksi & reuse hasil dari middleware sebelumnya di chain yang sama.
- **D**: Tidak ada proses review index saat kolom `source_type`/`source_id` pertama kali dibuat; project tidak punya migration tool yang biasanya akan memaksa developer mendeklarasikan index secara eksplisit di file migrasi.

---

## 3. Files Changed

| File | Perubahan |
|---|---|
| `backend/internal/repository/assignment_repo.go` | Tambah `WithTx(tx) AssignmentRepository` |
| `backend/internal/repository/attachment_repo.go` | Tambah `WithTx(tx) AttachmentRepository` |
| `backend/internal/service/attachment_service.go` | Tambah `WithTx(tx) AttachmentService` |
| `backend/internal/service/assignment_service.go` | `CreateAssignment`, `UpdateAssignment`, `Submit`, `UpdateSubmission`, `DeleteAssignment`, `DeleteSubmission` dibungkus `db.Transaction()`; notification fan-out di `CreateAssignment` dipindah ke `runAsync` |
| `backend/internal/service/async.go` | **Baru** — helper `runAsync(fn func())`, goroutine + panic recovery |
| `backend/internal/service/comment_service.go` | `notifyCommentRecipients` dipanggil lewat `runAsync` |
| `backend/internal/service/feed_service.go` | Notification fan-out dipindah ke `runAsync` |
| `backend/internal/service/material_service.go` | Notification fan-out dipindah ke `runAsync` |
| `backend/internal/middleware/rbac_middleware.go` | `RequireSchoolMember` cache roles ke context (`school_role_names`); `RequireRole` reuse cache tsb sebelum query ulang |
| `backend/internal/domain/attachment.go` | Tambah composite index tag `idx_attachments_source` |
| `backend/internal/domain/comment.go` | Tambah composite index tag `idx_comments_source` |
| `backend/schema.md` | Tambah `indexes {}` block untuk `attachments` dan `comments` (dokumentasi DBML) |
| `backend/cmd/api/main.go` | `NewAssignmentService(...)` diberi parameter tambahan `db` |
| `backend/internal/service/assignment_service_test.go` | Update stub (`WithTx`), helper test pakai sqlmock untuk transaksi |
| `backend/internal/service/material_summary_service_test.go` | Update stub (`WithTx`) |
| `backend/internal/middleware/rbac_middleware_test.go` | Tambah counter query + 1 test baru pembuktian optimasi |

Tidak ada file DTO, route path, atau response contract yang berubah.

---

## 4. Detail Perubahan Tiap File

### Target 1 — Assignment Transaction

Pola `WithTx` ditambahkan ke `AssignmentRepository`, `AttachmentRepository`, dan `AttachmentService` — masing-masing mengembalikan instance baru yang terikat ke `*gorm.DB` transaksi yang sedang berjalan, tanpa mengubah method lain di interface tersebut:
```go
func (r *assignmentRepository) WithTx(tx *gorm.DB) AssignmentRepository {
    return &assignmentRepository{db: tx}
}
```

`assignmentService` menerima dependency baru `db *gorm.DB` (constructor `NewAssignmentService` bertambah 1 parameter — dipanggil dari `main.go` dengan `db` yang sudah ada di composition root, tidak ada koneksi baru). Keenam fungsi dibungkus:
```go
return s.db.Transaction(func(tx *gorm.DB) error {
    if err := s.repo.WithTx(tx).CreateAssignment(asg); err != nil {
        return err
    }
    return replaceSourceAttachments(s.attService.WithTx(tx), asg.SchoolID, domain.SourceAssignment, asg.ID, attachmentMediaIDs)
})
```
`replaceSourceAttachments` (dan `ReplaceBySource` di baliknya) tetap memanggil `.Transaction()` sendiri secara internal — ini aman karena GORM otomatis memakai `SAVEPOINT` untuk transaksi bersarang di Postgres, tidak perlu diubah.

Untuk `DeleteAssignment`/`DeleteSubmission`, error dari `UnlinkBySource` yang sebelumnya **dibuang diam-diam** (`s.attService.UnlinkBySource(...)` tanpa cek return value) sekarang **dihormati** — jika unlink gagal, seluruh delete di-ROLLBACK. Ini adalah perubahan failure-mode yang disetujui eksplisit oleh Anda sebelum implementasi (bukan business logic baru — hanya menghentikan silent-fail yang sudah menjadi bug laten).

Notifikasi (best-effort) di `CreateAssignment` sengaja diletakkan **di luar** blok transaksi (tetap best-effort, tidak memengaruhi commit/rollback assignment).

### Target 2 — Notification Async

`internal/service/async.go` (baru):
```go
func runAsync(fn func()) {
    go func() {
        defer func() { _ = recover() }()
        fn()
    }()
}
```
4 lokasi fan-out (comment, assignment, feed, material) masing-masing dibungkus `runAsync(func() { ... })` — isi logic-nya **sama persis**, hanya eksekusinya dipindah ke goroutine terpisah setelah data utama (comment/assignment/feed/material) berhasil disimpan. Response HTTP tidak lagi menunggu N insert notification selesai. Panic di goroutine manapun tidak akan menjatuhkan proses (di-recover, silent — konsisten dengan sifat "best-effort" yang sudah didokumentasikan di kode sebelumnya).

### Target 3 — RBAC Query Optimization

`RequireSchoolMember`, setelah mengambil `roles` (hanya terjadi saat ada header `Active-Role`), menyimpannya ke context:
```go
c.Set("school_role_names", roles)
```
`RequireRole` mengecek context ini lebih dulu:
```go
var roles []string
haveCachedRoles := false
if cached, exists := c.Get("school_role_names"); exists {
    if cachedRoles, ok := cached.([]string); ok {
        roles = cachedRoles
        haveCachedRoles = true
    }
}
if !haveCachedRoles {
    roles, err = rbacRepo.GetUserRoleNamesInSchool(userID, schoolID)
    ...
}
```
Dipakai flag `haveCachedRoles` (bukan `roles == nil`) secara sengaja — karena user dengan **nol role** di sekolah tsb akan menghasilkan `roles == nil, err == nil` dari query (dikonfirmasi dari `rbac_repo.go`), yang kalau dipakai sebagai penanda "belum ada cache" akan salah query ulang. Behavior untuk semua kombinasi (dengan/tanpa `Active-Role`, dengan/tanpa `RequireSchoolMember` di depan, role kosong) **identik** dengan sebelumnya — hanya jumlah query yang berubah.

### Target 4 — Composite Index

```go
// attachment.go
SourceID   string     `gorm:"column:att_source_id;type:uuid;index:idx_attachments_source,priority:2" json:"sourceId"`
SourceType SourceType `gorm:"column:att_source_type;type:source_type;index:idx_attachments_source,priority:1" json:"sourceType"`
```
```go
// comment.go
SourceType SourceType `gorm:"column:cmn_source_type;type:source_type;index:idx_comments_source,priority:1" json:"sourceType"`
SourceID   string     `gorm:"column:cmn_source_id;type:uuid;index:idx_comments_source,priority:2" json:"sourceId"`
```
`schema.md` diberi `indexes { (att_source_type, att_source_id) }` dan `indexes { (cmn_source_type, cmn_source_id) }`, mengikuti konvensi yang sudah dipakai untuk index lain di file yang sama (`subjects`, `assignment_categories`).

**Action item manual (project ini tidak punya migration runner):** karena tidak ada `AutoMigrate`/migration tool, index di atas baru benar-benar aktif di database setelah SQL berikut dijalankan manual (mis. lewat Supabase SQL editor):
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_attachments_source
  ON edv.attachments (att_source_type, att_source_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comments_source
  ON edv.comments (cmn_source_type, cmn_source_id);
```
(`CONCURRENTLY` dipakai supaya tidak mengunci tabel saat index dibuat — aman dijalankan di database yang sedang dipakai staging/production sekaligus, sesuai konteks Anda.)

---

## 5. Reliability Improvement

- Assignment/submission tidak akan lagi tersimpan tanpa attachment-nya, atau sebaliknya attachment ter-link ke assignment yang gagal tersimpan — kegagalan di salah satu langkah sekarang me-rollback semuanya.
- `DeleteAssignment`/`DeleteSubmission` tidak lagi bisa menghapus entity utama sambil diam-diam gagal membersihkan attachment-nya (celah orphan attachment tertutup).
- Panic di background notification goroutine tidak bisa lagi menjatuhkan proses API (di-recover secara eksplisit).

## 6. Performance Improvement

- Response time `POST /assignments`, `PATCH /assignments/:id`, `POST /assignments/submit/:id`, `POST /comments`, `POST /feeds`, `POST /materials` tidak lagi terikat linear dengan jumlah murid di kelas — insert notification kini terjadi setelah response dikirim.
- Request yang memakai `Active-Role` header dan melalui `RequireSchoolMember` + `RequireRole` sekarang hanya melakukan **1 query** `GetUserRoleNamesInSchool` alih-alih 2 (dibuktikan lewat test `TestRequireRoleReusesCachedRolesFromRequireSchoolMember`, meng-assert `roleQueryCount == 1`).
- Setelah index diterapkan manual ke database (lihat action item di atas), query attachment/comment by source akan pakai index scan alih-alih sequential scan — berdampak ke hampir semua halaman detail material/assignment/feed.

## 7. Regression Impact

Diverifikasi: `go build ./...`, `go vet ./...`, `go test ./...` — **semua lulus**, termasuk 6 test lama yang harus disesuaikan (stub `WithTx`, helper sqlmock untuk transaksi) dan 1 test baru yang membuktikan optimasi RBAC bekerja.

- **Admin/Teacher/Student**: tidak ada perubahan response, status code, atau urutan operasi yang terlihat dari luar untuk assignment/submission/comment/feed/material — hanya kegagalan partial yang sekarang konsisten (rollback total, bukan partial save).
- **Notifikasi**: user tetap menerima notifikasi yang sama persis, hanya dengan latency tambahan (goroutine terpisah, biasanya milidetik) — tidak ada notifikasi yang hilang pada kondisi normal.
- **RBAC**: behavior otorisasi identik untuk semua kombinasi role/membership/active-role — dibuktikan lewat regression test baru, bukan hanya diasumsikan.
- **Index**: penambahan index tidak mengubah hasil query, hanya execution plan-nya (butuh dijalankan manual, lihat action item Target 4).

## 8. Breaking Change

**Tidak ada.** Tidak ada perubahan endpoint, DTO, atau status code. Constructor `NewAssignmentService` bertambah 1 parameter (`db`) — ini murni internal wiring di `main.go`, bukan bagian dari public API aplikasi.

## 9. Verification Result

```
go build ./...   → OK
go vet ./...     → OK
go test ./...    → ok backend/internal/handler
                   ok backend/internal/middleware
                   ok backend/internal/service
```

Seluruh test lulus, termasuk test transaksi (`TestRemoveMember*` dari Sprint 2), test Sprint 3 baru (`TestRequireRoleReusesCachedRolesFromRequireSchoolMember`), dan seluruh test Assignment/Grade/Material/StudentNote yang sudah ada sebelumnya.

---

## Catatan

Ditemukan pola non-atomic yang identik (mutate + attachment link) juga di `material_service.go` (`CreateMaterial`) — **di luar scope eksplisit Sprint 3** (audit Target 1 secara spesifik hanya mencakup Assignment), jadi tidak disentuh. Dilaporkan di sini untuk pertimbangan sprint berikutnya jika diperlukan.

Terpisah dari pekerjaan sprint ini: file `docs/SPRINT2_P0_COMPLETION_2026-07-10.md` yang saya buat di sprint sebelumnya terdeteksi sudah tidak ada lagi di working tree saat sprint ini dimulai (muncul sebagai "deleted" di git status) — saya tidak menghapusnya lewat perintah eksplisit di sesi ini. Kemungkinan besar terkait proses auto-commit/checkpoint IDE yang sebelumnya juga pernah terdeteksi. Memberi tahu untuk transparansi, bukan sesuatu yang saya perbaiki/pulihkan tanpa konfirmasi Anda.
