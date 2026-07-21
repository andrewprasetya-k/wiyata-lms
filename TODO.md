# TODO

## Selesai (Phase 8)

### RBAC Improvements
- [x] Audit dan implementasi UI multi-role assignment (checkbox/multi-select). — `AdminUsers.vue`, Batch 2.
- [x] Ubah AdminUsers dari single-role menjadi multi-role editor. — Batch 2.
- [x] Pastikan update role tidak lagi menghapus role lain secara tidak sengaja. — ini bug data-loss nyata: `SyncUserRoles` backend sudah full-replace-by-array sejak awal, tapi UI lama cuma dropdown single-select yang selalu mengirim array 1 elemen, jadi menyimpan member yang punya 2+ role akan diam-diam menghapus role lainnya. Diperbaiki Batch 2 dengan multi-checkbox editor yang pre-select semua role yang sedang aktif.

### Bug: Super Admin dapat 400/403 di hampir semua route platform
- [x] Root cause ada dua: (1) `RequireRole(...)` sendirian (tanpa `RequireSchoolMember`) butuh school context yang memang tidak pernah dikirim frontend untuk route yang murni platform-level (`/schools`, `/schools/summary`, `/dashboard/super-admin`, `/rbac/roles`, `/rbac/super-admin`) → 400 "School context required". (2) `RequireSystemSuperAdmin` sendiri selalu 403 untuk siapa pun, termasuk super admin asli, karena resolve "system school" berkode `000000` yang tidak pernah dibuat oleh kode manapun (`CreateSuperAdmin` justru enroll ke sekolah bernama "admin", bukan berkode "000000"). Diperbaiki Batch 1: `RequireSystemSuperAdmin` sekarang cek `rbacRepo.IsSuperAdmin(userID)` langsung (primitive yang sama dipakai `RequireSchoolMember`'s bypass), dan 9 route super-admin-only dipindah dari `RequireRole(schoolService, "super_admin")` ke `RequireSystemSuperAdmin`.
  Detail lengkap + trace: `docs/PROJECT_CONTEXT_HANDOFF.md` §27.

### Existing-user invitation flow
- [x] "admin invite anggota sekolah by email saja. ketika user mau accept, yang sudah punya akun lgsg klik 'terima'. yang belum, buat akun dulu" — selesai (Batch 3 + follow-up JWT-enforcement):
  - `GET /invitations/:token` sekarang balikin `existingUser: bool`.
  - User baru: form nama/password lama tidak berubah sama sekali, tetap `POST /invitations/:token/accept` (public).
  - User existing: login dulu (redirect otomatis balik ke halaman invite lewat mekanisme redirect yang sudah ada di router/login), lalu accept lewat `POST /invitations/:token/accept-authenticated` — endpoint baru, wajib JWT, backend verifikasi `email akun login == email invitation` sebelum accept (bukan cuma dicek di frontend).
  - Form Create Invitation di `AdminUsers.vue` sekarang cuma minta email + role (+ class kalau siswa) — field Nama dihapus. `Invitation.FullName` di database jadi optional/nullable; tidak pernah dipakai oleh accept flow manapun (nama akun baru selalu dari form Accept, bukan dari form admin). Satu-satunya pemakai sisa: daftar "Undangan Tertunda" di `AdminDashboard.vue`, sudah fallback ke email kalau nama kosong.
  Detail API: `backend/docs/api/invitation.md`, `backend/docs/api/school_member_invitations.md`.

## Belum Dikerjakan

- Evaluasi invitation & enroll untuk mendukung multi-role bila memang dibutuhkan — `CreateSchoolMemberInvitationDTO.Role` masih satu string, satu invitation masih satu role. Multi-role editor Batch 2 hanya untuk mengedit role member yang *sudah* jadi anggota sekolah, bukan untuk mengundang/enroll dengan banyak role sekaligus.
- Audit seluruh frontend terhadap asumsi single-role — baru `AdminUsers.vue` yang diaudit dan diperbaiki. Halaman lain yang menampilkan/mengedit role member belum dicek.
- Logging untuk admin sekolah mengenai sekolah (backend banyak bertambah, web socket) dan juga superadmin (lebih umum, ga sedetail admin sekolah) — belum dikerjakan sama sekali. Perlu scoping (event apa saja, retensi, granularitas admin vs superadmin) sebelum implementasi; ada mekanisme SSE yang sudah ada (`/api/events/sidebar`) yang berpotensi dipakai ulang untuk ini alih-alih bikin websocket baru dari nol.
