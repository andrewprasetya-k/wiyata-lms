# Sprint 2 — P0 Completion: Media IDOR Closure + Critical Regression Tests + Production Cleanup

**Tanggal:** 2026-07-10
**Scope:** Menutup false negative `GET /medias/:id`, menghapus debug log production, dan menambahkan regression test untuk area security paling kritis, tanpa mengubah API contract atau business logic.

---

## 1. Audit (Ringkasan)

Diverifikasi ulang berbasis kode aktual (bukan asumsi), sesuai temuan audit P0 sebelumnya:

- **`GET /medias/:id`** — route (`main.go:350`) hanya memakai `AuthRequired` global, handler (`media_handler.go`) langsung `h.service.GetByID(id)` tanpa cek apa pun, repository query `WHERE med_id = ?` saja. Domain `Media` punya kolom `SchoolID` yang tidak pernah dibandingkan. Frontend tidak pernah memanggil endpoint ini secara langsung (dikonfirmasi lewat `grep` di `frontend/src/services`).
- **`stores/auth.ts:278`** — `console.log("SUPERADMIN_SCHOOL_ID", superAdminSchoolId)` dikonfirmasi masih ada persis seperti temuan audit.
- **Test coverage** — dikonfirmasi ulang: nol test untuk `RequireSchoolMember`/`RequireRole`/`RequireSystemSuperAdmin` sebagai unit, nol regression test IDOR, nol test untuk `RemoveMember()`.

---

## 2. Vulnerability & Root Cause

**`GET /medias/:id`** — cross-tenant IDOR. Siapa pun user yang login (role dan sekolah apa pun) bisa membaca metadata media (termasuk `fileUrl`, `storagePath`, `thumbnailUrl`, `ownerType`/`ownerId`) milik sekolah lain hanya dengan menebak/mengiterasi UUID media, karena tidak ada satu pun dari tiga lapis (route/handler/repository) yang memvalidasi kepemilikan sekolah. Pola root cause-nya identik dengan yang ditemukan di Sprint 1 (subject/class/academic-year/term): middleware generik saja tidak cukup karena `RequireSchoolMember` hanya memvalidasi bahwa user adalah anggota sekolah yang disebut di header — bukan bahwa resource `:id` yang diminta benar-benar milik sekolah itu.

---

## 3. Implementation

### Target 1 — Close `GET /medias/:id` IDOR

Menerapkan pola defense-in-depth yang identik dengan Subject/Class/AcademicYear handler di Sprint 1:

**`backend/cmd/api/main.go`**
```go
// sebelum
mediaAPI.GET("/:id", mediaHandler.GetByID)
// sesudah
mediaAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), mediaHandler.GetByID)
```

**`backend/internal/handler/media_handler.go`**
```go
func (h *MediaHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	activeSchoolID, ok := getMediaActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	media, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if media.SchoolID != activeSchoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: media does not belong to active school"})
		return
	}
	c.JSON(http.StatusOK, media)
}
```
`getMediaActiveSchoolID` adalah helper yang **sudah ada sebelumnya** (dipakai `Upload`/`Delete`) — tidak ada helper baru yang dibuat. `service.GetByID(id)` **tidak diubah signature-nya** karena hanya dipanggil dari file ini (2 pemanggil: `GetByID` dan `Delete`), sehingga ownership check cukup dilakukan di handler, persis pola Sprint 1.

### Target 2 — Remove production debug log

**`frontend/src/stores/auth.ts`** — baris `console.log("SUPERADMIN_SCHOOL_ID", superAdminSchoolId);` dihapus. Tidak ada perubahan lain di fungsi `persistCurrentSession()`.

### Target 3 — Regression tests

**A. Authorization middleware** — `backend/internal/middleware/rbac_middleware_test.go` (baru)
Test langsung terhadap middleware (bukan lewat satu fitur spesifik), pakai stub `RBACRepository` + stub `SchoolService`:
- `TestRequireSchoolMember` — allow (member sekolah yang benar), reject (sekolah salah), reject (school context hilang).
- `TestRequireSchoolMemberSuperAdminBypass` — super admin tanpa active role bypass cek membership; tapi begitu ada `Active-Role` eksplisit, tetap butuh membership asli.
- `TestRequireRole` — allow (role diizinkan), reject (role tidak ada di sekolah tsb).
- `TestRequireRoleWrongRole` — reject (role ada tapi tidak diizinkan untuk route ini).
- `TestRequireSystemSuperAdmin` — allow (super_admin di sekolah sistem), reject (bukan super_admin).

**B. Media IDOR regression** — `backend/internal/handler/media_handler_route_test.go` (baru)
`TestMediaGetByIDCrossTenantIsForbidden` — meregistrasikan route persis seperti production (`RequireSchoolMember` + `mediaHandler.GetByID`), memverifikasi: media milik sekolah aktif → 200, media milik sekolah lain → 403.

**C. Sprint 1 IDOR regression (representative)** — `backend/internal/handler/subject_handler_route_test.go` (baru)
`TestSubjectGetByIDCrossTenantIsForbidden` — pola sama seperti B, untuk `GET /subjects/:id`. Dipilih sebagai representative karena paling langsung mencerminkan pola yang direplikasi ke 8 endpoint lain di Sprint 1.

**D. Cascade unenrollment** — `backend/internal/service/admin_school_member_import_remove_test.go` (baru)
`RemoveMember()` memakai `*gorm.DB` langsung (bukan lewat repository interface) dan `DB_DSN` project ini menunjuk ke **database Supabase remote yang nyata** — jadi test **tidak boleh** menyentuh koneksi itu. Solusi realistis: dipakai [`go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock) (ditambahkan sebagai dependency baru, `go.mod`/`go.sum` diupdate) untuk mem-mock driver SQL di bawah GORM (`postgres.Config{Conn: sqlDB}`), sehingga transaksi & SQL yang benar-benar dijalankan GORM bisa diverifikasi tanpa DB sungguhan:
- `TestRemoveMemberCascadesEnrollmentLeftAt` — happy path: `school_users.deleted_at` dan `enrollments.left_at` di-update dalam SATU transaksi, dan nilai timestamp-nya **dipaksa sama** lewat custom argument matcher (`sameValueArg`) yang menolak match jika value kedua berbeda dari yang pertama.
- `TestRemoveMemberRollsBackWhenSchoolUserNotFound` — update pertama 0 rows affected (schoolUserID/schoolID tidak cocok) → transaksi ROLLBACK, update kedua (enrollments) **tidak pernah dieksekusi**.
- `TestRemoveMemberRollsBackWhenEnrollmentUpdateFails` — update kedua gagal → transaksi ROLLBACK, memverifikasi `school_users` tidak akan pernah ke-persist ter-soft-delete kalau enrollment cascade-nya gagal.

Ini bukan test "fake" — benar-benar mengeksekusi kode `RemoveMember` asli lewat GORM asli, hanya driver SQL-nya yang di-mock.

---

## 4. Files Changed

| File | Perubahan |
|---|---|
| `backend/cmd/api/main.go` | Tambah `RequireSchoolMember` ke `GET /medias/:id` |
| `backend/internal/handler/media_handler.go` | `GetByID` tambah school-context check + ownership check |
| `backend/go.mod`, `backend/go.sum` | Tambah dependency `github.com/DATA-DOG/go-sqlmock v1.5.2` (test-only) |
| `frontend/src/stores/auth.ts` | Hapus 1 baris `console.log` debug |
| `backend/internal/middleware/rbac_middleware_test.go` | **Baru** — unit test `RequireSchoolMember`/`RequireRole`/`RequireSystemSuperAdmin` |
| `backend/internal/handler/media_handler_route_test.go` | **Baru** — regression test IDOR untuk media |
| `backend/internal/handler/subject_handler_route_test.go` | **Baru** — regression test IDOR representative Sprint 1 |
| `backend/internal/service/admin_school_member_import_remove_test.go` | **Baru** — test transaksi cascade unenrollment (sqlmock) |

Tidak ada file DTO, route path (selain penambahan middleware), atau response contract yang berubah.

---

## 5. Flow Baru — `GET /medias/:id`

```
Request
  → RequireSchoolMember (validasi user adalah member dari SchoolId header)
  → MediaHandler.GetByID
      → ambil schoolID dari context (wajib ada, 400 jika tidak)
      → service.GetByID(id) — fetch media apa adanya
      → bandingkan media.SchoolID vs activeSchoolID
          → beda → 403 Forbidden
          → sama → 200 OK, response identik seperti sebelumnya
```

---

## 6. Regression Impact

- **Admin/Teacher/Student**: tidak terdampak. `GET /medias/:id` tidak pernah dipanggil langsung oleh frontend (dikonfirmasi ulang lewat `grep` di `frontend/src/services`) — metadata media selalu ikut ter-embed lewat `Preload` di response material/assignment/feed/chat, yang tidak melalui endpoint ini.
- **Super admin**: tidak terdampak — tidak ada flow super admin yang memakai endpoint ini.
- Diverifikasi: `go build ./...`, `go vet ./...`, `go test ./...` — **semua lulus**, termasuk 6 test suite (2 baru untuk middleware+handler regression, 1 baru untuk cascade transaction, plus 3 test file lama yang sudah ada sebelumnya).
- Frontend: `npm run build` — sukses, tanpa warning baru (warning yang ada sudah dikonfirmasi pre-existing sejak Sprint sebelumnya).

---

## 7. Security Improvement

`GET /medias/:id` sekarang mengikuti pola defense-in-depth yang sama seperti seluruh endpoint Sprint 1: route memastikan requester adalah anggota sah dari sekolah yang diklaim, handler memverifikasi resource yang di-fetch benar-benar milik sekolah itu (bukan hanya mengandalkan middleware). Ini menutup jalur enumerasi UUID media lintas sekolah yang sebelumnya terbuka untuk siapa pun yang punya JWT valid, terlepas dari role atau sekolah aktif mereka.

Regression test yang ditambahkan memastikan proteksi ini (dan pola Sprint 1 secara umum) tidak bisa regresi diam-diam di masa depan — refactor yang tanpa sengaja menghapus ownership check sekarang akan langsung gagal di CI, bukan baru ketahuan lewat insiden produksi.

---

## 8. Test Coverage yang Ditambahkan

| Area | File | Test |
|---|---|---|
| Authorization middleware | `middleware/rbac_middleware_test.go` | 5 test function, 11 sub-test (allow/reject/wrong-school/missing-role/super-admin-bypass) |
| Media IDOR | `handler/media_handler_route_test.go` | 1 test function, 2 sub-test |
| Sprint 1 IDOR (representative) | `handler/subject_handler_route_test.go` | 1 test function, 2 sub-test |
| Cascade unenrollment | `service/admin_school_member_import_remove_test.go` | 3 test function (happy path + 2 skenario rollback) |

Total **10 test function baru**, seluruhnya lulus, nol koneksi ke database sungguhan (baik untuk middleware/handler test yang pakai stub, maupun untuk `RemoveMember` yang pakai sqlmock).

---

## 9. Remaining P0 (belum diselesaikan sprint ini, di luar scope yang diminta)

- **Migration tooling** — masih belum ada (golang-migrate/atlas/goose), `schema.md` masih dokumentasi DBML manual, bukan migration yang bisa dijalankan. Ini butuh sprint tersendiri karena scope-nya besar (pemilihan tool, konversi schema.md ke migration file, testing terhadap DB nyata) dan eksplisit di luar target sprint ini.
- Test coverage backend masih jauh dari lengkap secara keseluruhan (repository layer, mayoritas service/handler lain) — sprint ini hanya menutup 4 area paling kritis yang diminta, bukan coverage menyeluruh.

---

## 10. Breaking Change

**Tidak ada.** Tidak ada perubahan pada URL, request/response DTO, atau status code semantics selain endpoint yang tadinya bisa diakses tanpa header `SchoolId` sekarang mengembalikan `400`/`403` — sama seperti pola breaking-change-non-breaking yang sudah dijelaskan di laporan Sprint 1, karena endpoint ini tidak pernah dipakai lintas sekolah oleh siapa pun secara sah.

Satu penambahan yang terlihat dari sisi development: dependency baru `github.com/DATA-DOG/go-sqlmock` di `go.mod` — ini **test-only dependency**, tidak dipakai di kode produksi (`cmd/api`, `internal/handler`, `internal/service` non-test), tidak berdampak ke binary yang di-deploy.
