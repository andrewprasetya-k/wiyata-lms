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

## Selesai (Phase 9)

### Evaluasi Multi-Role Invitation & Enrollment (9.1, read-only audit)
- [x] Evaluasi invitation & enroll untuk mendukung multi-role — **ditutup, tidak diimplementasikan (Option A)**. Root cause: tabel `enrollments` punya unique constraint `(enr_scu_id, enr_cls_id)` tanpa kolom `enr_role`, jadi satu orang secara struktural database hanya bisa punya satu role per kelas — multi-role invitation tidak akan mengubah batasan ini. Teacher invitation juga tidak pernah auto-create enrollment, jadi multi-role di invitation tidak menghilangkan langkah admin manual yang memang selalu diperlukan. Kesimpulan: single-role-per-invitation adalah desain yang benar, bukan keterbatasan yang perlu ditambal.

### Audit Frontend Single-Role Assumption (9.2, read-only audit)
- [x] Audit seluruh frontend terhadap asumsi single-role — **selesai, tidak hanya `AdminUsers.vue`**. Satu bug nyata ditemukan: `AdminEnrollments.vue`'s `inferPlacementRole()` mengembalikan `null` (memblokir UI enrollment) untuk member yang punya role `student` DAN `teacher` sekaligus — state yang baru bisa terjadi sejak multi-role editor Batch 2. `ReadProfile.vue` dan `AdminSubjectClasses.vue` dikonfirmasi sudah benar menangani multi-role (dipakai sebagai referensi pola). Bug ini diperbaiki di 9.3 dengan mencegah kombinasi role-nya sejak sumber, bukan menambal `AdminEnrollments.vue`.

### Validasi Kombinasi Role Ilegal (9.3, implementasi)
- [x] Business rule: `admin`+`teacher` satu-satunya kombinasi role sekolah yang diizinkan pada satu membership; `student` tidak boleh digabung dengan `teacher` maupun `admin`. `super_admin` di luar scope (platform role, tidak pernah dikelola dari `AdminUsers.vue`).
  - Backend jadi source of truth tunggal: `domain.ValidateSchoolRoleCombination` (`backend/internal/domain/role_validation.go`) — sengaja di layer `domain`, bukan `service`, supaya bisa dipanggil dari `service` maupun `repository` tanpa import cycle.
  - Dipanggil dari **setiap** jalur yang bisa mengubah role set sebuah `school_users`: `rbacService.SyncUserRoles` (role editor `AdminUsers.vue`) dan `AssignRoleToUser` (endpoint `POST /rbac/user-roles`, tidak dipakai frontend tapi reachable via API langsung), CSV import & direct member creation (`adminSchoolMemberImportService`), serta invitation accept — **kedua** endpoint (`POST /invitations/:token/accept` dan `/accept-authenticated`) via helper bersama `finalizeInvitationAcceptance`.
  - Frontend (`AdminUsers.vue`): inline validation live saat toggle checkbox role, pakai komponen `InlineFormError` yang sudah ada (tidak ada modal baru), tombol Simpan disabled saat kombinasi tidak valid, request tidak pernah dikirim untuk kombinasi ilegal.
  - Error response konsisten dengan pola project (`HandleError`, `errors.Is`), pesan Indonesia yang jelas menyebutkan kombinasi mana yang ditolak.
  - Detail lengkap: `backend/docs/api/rbac.md` §2, `backend/docs/api/invitation.md`, `backend/docs/api/school_member_import.md`, `docs/PROJECT_CONTEXT_HANDOFF.md` §24/§26/§27.

## Selesai (Phase 10 — Audit Log)

### Audit Log: Infrastruktur, Write Path, REST, dan Real-Time (10.1–10.10)
- [x] Item "Belum Dikerjakan" sebelumnya ("Logging untuk admin sekolah... dan superadmin... web socket") — **selesai**, dikerjakan bertahap 10 sub-phase, audit-only dulu (10.1–10.3, tanpa kode) baru implementasi:
  - **10.4** — infrastruktur: migration `0003_extend_logs_for_audit.sql` (8 kolom nullable baru di `edv.logs`), `domain.Log` diperluas, `domain.ActorContext` baru, `LogRepository`/`LogService` (`Log`/`LogBatch`) — additive, belum ada business service yang dipanggil.
  - **10.5–10.8** — 33 action ditulis di 9 domain: RBAC, Member Management (termasuk CSV import lewat `LogBatch` + `correlation_id`), Enrollment, Subject Class, School, Platform Bootstrap, Assignment, Grade, Authentication.
  - **10.9** — REST read surface: service baru `LogQueryService` (sengaja terpisah dari `LogService` yang dipakai write path), endpoint `GET /logs`, `/logs/:id` (super admin, platform-wide), `/logs/school/:schoolId/search`, `/logs/school/:schoolId/entries/:id` (school admin/super admin, dipin), plus Audit Viewer frontend pertama (`AuditLogsPage.vue`, dipakai di `/admin/audit-logs` dan `/superadmin/audit-logs`).
  - **10.10** — real-time: **bukan** websocket baru dari nol — reuse `realtime.Hub`/`Client` yang sudah dipakai chat (instance terpisah), 1 method baru (`BroadcastToRoom`) untuk room-wide fanout, `events.AuditBroadcaster` (mirror `events.SidebarBroadcaster`), endpoint `GET /api/ws/audit`. SSE `/api/events/sidebar` yang disebut di catatan lama **tidak** dipakai ulang untuk ini (payload/permission-nya beda — butuh room-wide fanout, bukan targeted per-user).
  - Detail lengkap: `backend/docs/api/log.md` (arsitektur, taxonomy+severity, REST+WebSocket contract, permission matrix, known limitation), `docs/PROJECT_CONTEXT_HANDOFF.md` §26.

## Belum Dikerjakan

(kosong per Phase 10.11)
