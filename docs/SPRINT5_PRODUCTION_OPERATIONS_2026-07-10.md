# Sprint 5 — Production Operations & Developer Experience

**Tanggal:** 2026-07-10
**Scope:** Standardisasi error handling, dokumentasi konvensi database, rate limiting per-tenant, dan structured logging + request ID. Tidak ada perubahan endpoint, DTO, business logic, authorization, atau struktur database.

---

## 1. Audit Summary

### Target 1 — Error Surfacing

- `toast.success`: 46 pemakaian. `toast.error`: 77 pemakaian. `console.error`: 2 pemakaian (keduanya sudah legitimate — dipasangkan dengan feedback user-facing, dikonfirmasi ulang dari audit sebelumnya).
- `getApiError` (helper generik existing di `utils/error.ts`) dipakai di 30 tempat, terbagi antara toast (16) dan inline (12).
- **Pola yang sudah konsisten dan baik** (tidak disentuh): kegagalan **memuat** konten (GET saat mount) selalu ditampilkan **inline** dengan tombol "Coba lagi" — dipakai luas di puluhan komponen (`CommentThread.vue`, `AdminAcademicYears.vue`, `TeacherAssignmentReview.vue`, dll.), terlepas dari penyebabnya (network/403/404/500). Kegagalan **aksi** (submit/delete/toggle) sudah konsisten pakai **toast**.
- **Pelanggaran ditemukan**: **15 kemunculan** `toast.error(...)` dipakai untuk validasi client-side murni (field kosong/format salah, sebelum request API dikirim) — melanggar aturan eksplisit "jangan ada toast untuk validation form". Tersebar di 6 file: `AdminAcademicYears.vue` (5), `AdminUsers.vue` (4, salah satunya sudah punya ref inline `importError` yang belum dipakai untuk validasi), `AdminClasses.vue` (2), `AdminEnrollments.vue` (2, satu di antaranya ditemukan saat implementasi bukan saat audit awal), `TeacherFeed.vue` (1), `TeacherAssignmentReview.vue` (1).
- Tidak ditemukan pelanggaran sebaliknya (inline dipakai untuk pure network/action error) — pola yang ada sudah rapi untuk kasus itu.

### Target 2 — Database Naming

- `backend/schema.md` (DBML) dibaca penuh — 32 tabel, semua memakai prefix kolom 3-4 huruf, konsisten sejak awal. Tidak ada tabel yang menyimpang polanya kecuali `chat_rooms` (prefix `room_`, bukan singkatan standar).

### Target 3 — Rate Limiting

- Tidak ada rate limiting middleware sama sekali sebelumnya (dicek `internal/middleware/`: hanya `auth_middleware.go`, `rbac_middleware.go`).
- `school_id` di context di-set oleh `RequireSchoolMember`/`RequireRole`/`RequireSystemSuperAdmin` — tapi middleware ini berjalan **setelah** titik yang paling masuk akal untuk rate-limit global (`api.Use(...)`, yang berjalan sebelum middleware per-route manapun). Solusi: pakai header `SchoolId` mentah langsung (bukan context `school_id` yang divalidasi), karena rate limiting adalah concern abuse-prevention, bukan authorization — aman dipakai sebelum validasi keanggotaan selesai.
- Endpoint publik pra-auth: `/login`, `/register`, `/school-registration-requests` (tidak ada school context), `/invitations/:token` (GET+accept, token-gated), `/events/sidebar` (SSE), `/ws/chat` (WebSocket).
- `golang.org/x/time/rate` (official Go extended library, bukan pihak ketiga) dipilih sebagai library ringan untuk token-bucket in-memory.

### Target 4 — Structured Logging

- Tidak ada structured logger sama sekali — hanya `gin.Default()` (plain-text access log bawaan Gin) dan beberapa `fmt.Printf("[Email Warning] ...")` untuk notifikasi email gagal (di luar scope — bukan request log).
- Tidak ada request-ID tracking di mana pun.
- Go 1.25 (dari `go.mod`) sudah mendukung `log/slog` di standard library sejak Go 1.21 — dipilih sebagai logger karena **tidak perlu dependency baru** (paling ringan yang mungkin), dan merupakan solusi structured-logging resmi dari tim Go.

---

## 2. Root Cause

- **Error handling**: tidak ada dokumentasi/konvensi tertulis sejak awal project — developer menambah form baru cenderung meniru file terdekat yang pernah dilihat, bukan aturan eksplisit, sehingga toast-untuk-validasi menyebar organik ke beberapa file tanpa disadari sebagai anti-pattern.
- **Rate limiting**: project awalnya single-tenant-per-request-cycle sederhana, belum pernah ada insiden abuse yang memaksa penambahan proteksi ini.
- **Structured logging**: `gin.Default()` dipakai sebagai default bawaan saat setup awal project dan tidak pernah diganti karena belum ada kebutuhan operasional (debugging lintas request, korelasi log) yang mendesak.

---

## 3. Files Changed

### Backend

| File | Perubahan |
|---|---|
| `internal/middleware/rate_limit.go` | **Baru** — `RateLimiterStore` interface + `InMemoryRateLimiterStore` (token bucket per key via `golang.org/x/time/rate`, dengan sweep otomatis entry basi) + `RateLimitPerTenant()` middleware |
| `internal/middleware/rate_limit_test.go` | **Baru** — 3 test (burst/block per tenant, fallback IP, sweep basi) |
| `internal/middleware/request_id.go` | **Baru** — `RequestID()` middleware + `GetRequestID()` helper |
| `internal/middleware/logging.go` | **Baru** — `StructuredLogger()` middleware (slog, JSON) |
| `internal/middleware/logging_test.go` | **Baru** — 2 test (request-id generate/echo, field log lengkap) |
| `cmd/api/main.go` | Wiring: `gin.Default()` → `gin.New()+Recovery()+RequestID()+StructuredLogger()`; rate limiter dipasang di `api.Use(...)` (protected routes) + 3 endpoint publik individual (login/register/school-registration) |
| `go.mod`, `go.sum` | Tambah `golang.org/x/time` (rate limiter) |

### Frontend

| File | Perubahan |
|---|---|
| `components/common/InlineFormError.vue` | **Baru** — komponen kecil untuk styling konsisten pesan validasi inline |
| `utils/errorPresentation.ts` | **Baru** — `classifyApiError()` untuk kategorisasi error (referensi, bukan pengganti resolver pesan yang sudah ada) |
| `pages/admin/AdminAcademicYears.vue` | 5 validasi toast → inline (`academicYearFormError`, `termFormError`, `subjectFormError`, `categoryFormError`) |
| `pages/admin/AdminUsers.vue` | 4 validasi toast → inline (3 direct ke `importError` yang sudah ada, 1 ref baru `memberFormError` dipakai bersama tab invite+manual) |
| `pages/admin/AdminClasses.vue` | 2 validasi toast → inline (`classFormError`, `editFormError`) |
| `pages/admin/AdminEnrollments.vue` | 2 validasi toast → inline (`enrollmentFormError`) |
| `pages/teacher/TeacherFeed.vue` | 1 validasi toast → inline (`composeFormError`) |
| `pages/teacher/TeacherAssignmentReview.vue` | 1 validasi toast → inline (`gradeFormError`) |

### Dokumentasi

| File | Isi |
|---|---|
| `docs/ERROR_HANDLING.md` | **Baru** — konvensi lengkap, prinsip "load vs action", anti-pattern, contoh, daftar perbaikan Sprint 5 |
| `docs/DATABASE_NAMING.md` | **Baru** — 32 prefix tabel, kepanjangan, contoh kolom, kolom umum tanpa prefix, catatan khusus |

---

## 4. Detail Perubahan

### Error Handling

Setiap form yang diperbaiki mengikuti pola yang sama: tambah 1 ref string kosong per form (atau reuse ref inline yang sudah ada bila relevan — lihat `AdminUsers.vue`'s `importError`), reset ke `""` di awal fungsi submit, isi dengan pesan **yang sama persis seperti sebelumnya** (tidak ada perubahan wording) saat validasi gagal, lalu render `<InlineFormError :message="xRef" />` di template dekat field/tombol submit. Kegagalan dari API (`catch` block) **tidak diubah** — tetap toast, karena itu kategori berbeda (business rule/unexpected).

### Rate Limiting

```go
rateLimiterStore := middleware.NewInMemoryRateLimiterStore(20, 40, 10*time.Minute)
```
20 request/detik sustained, burst 40, per key. Key = `"school:" + SchoolId header` jika ada, else `"ip:" + ClientIP()`. Dipasang di:
- `api.Use(middleware.RateLimitPerTenant(rateLimiterStore))` — setelah `AuthRequired()`, otomatis melindungi **seluruh route terproteksi** (mayoritas endpoint aplikasi) tanpa menyentuh satu pun definisi route individual.
- 3 endpoint publik individual (`/login`, `/register`, `/school-registration-requests`) — karena belum ada school context, otomatis fallback ke IP, melindungi dari brute-force/spam.

**Pengecualian (dengan alasan):**
- `/invitations/:token` (GET+accept) — token panjang/di-hash, menebak tidak feasible; frekuensi pemakaian rendah secara alami.
- `/events/sidebar` (SSE) dan `/ws/chat` (WebSocket) — koneksi persisten berumur panjang; rate limiter per-request tidak relevan untuk satu koneksi yang bertahan lama, dan membatasi handshake berisiko menolak reconnect sah saat ada gangguan jaringan sesaat.

### Structured Logging

```go
requestLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
r := gin.New()
r.Use(gin.Recovery())
r.Use(middleware.RequestID())
r.Use(middleware.StructuredLogger(requestLogger))
```
Setiap request menghasilkan satu baris JSON log:
```json
{"time":"...","level":"INFO","msg":"http_request","request_id":"...","method":"GET","path":"/api/subjects","status":200,"latency_ms":12,"client_ip":"...","user_id":"...","school_id":"..."}
```
`user_id`/`school_id` hanya muncul kalau ada (request belum authenticated atau belum resolve school context tidak akan menampilkan key tersebut). Level otomatis naik ke `WARN` untuk status 4xx dan `ERROR` untuk 5xx. Response selalu membawa header `X-Request-ID` — dipakai ulang kalau caller sudah mengirimnya sendiri (berguna untuk trace lintas layanan di masa depan).

---

## 5. Error Handling Convention

Lihat `docs/ERROR_HANDLING.md` untuk detail lengkap. Ringkasan prinsip: **load-failure selalu inline+retry** (berlaku untuk semua penyebab termasuk network — ini sengaja menyimpang dari saran awal sprint "Network → toast" karena audit membuktikan pola inline yang sudah ada adalah pola yang lebih baik dan sudah luas dipakai), **action-failure selalu toast**, **validasi client-side selalu inline dekat form** (bagian yang diperbaiki sprint ini).

---

## 6. Rate Limiting Design

Lihat bagian 4 di atas + `internal/middleware/rate_limit.go`. Desain sengaja dibuat swappable: `RateLimiterStore` adalah interface satu method (`Allow(key string) bool`), implementasi in-memory saat ini bisa diganti backend terdistribusi (mis. Redis) kelak tanpa mengubah middleware atau titik pemasangannya di `main.go`.

---

## 7. Structured Logging Design

Lihat bagian 4 di atas + `internal/middleware/logging.go` dan `request_id.go`. `log/slog` dipilih karena bagian dari standard library Go (tidak menambah dependency), dan merupakan API structured-logging resmi sejak Go 1.21.

---

## 8. Database Naming Convention

Lihat `docs/DATABASE_NAMING.md` — 32 prefix, diambil langsung dari `backend/schema.md`, tidak ada yang ditebak.

---

## 9. Regression Impact

- **Error handling**: nol perubahan wording pesan, nol perubahan endpoint/DTO. Perilaku yang terlihat: pesan validasi yang tadinya muncul sebagai toast (hilang otomatis 4 detik) sekarang muncul sebagai teks statis di dekat form (hilang saat validasi lolos atau form direset) — perubahan UX yang secara eksplisit disahkan sprint ini ("konsistensi penyajian error"), bukan regresi.
- **Rate limiting**: default 20 req/s burst 40 per tenant sengaja dibuat generous — jauh di atas pola pemakaian normal (dashboard yang memuat beberapa request paralel saat mount jauh di bawah 40). Risiko false-positive (user sah kena limit) rendah, tapi **perlu dipantau di production** — kalau ada halaman yang ternyata memicu lebih dari 40 request beruntun dari satu sekolah dalam waktu singkat, limit perlu dinaikkan (achievable dengan mengubah 2 angka di `main.go`, tidak perlu ubah kode lain).
- **Structured logging**: `gin.Default()` → `gin.New()+Recovery()+...` mengubah FORMAT log dari teks ke JSON, tapi **tidak mengubah apa pun yang terlihat oleh client** (bukan bagian dari response, kecuali header baru `X-Request-ID` yang ditambahkan, tidak menghapus header lain).
- Semua perubahan diverifikasi lewat `go build`, `go vet`, `go test ./...` (termasuk 5 test baru: 3 rate-limit + 2 logging), dan `npm run build` (full type-check `vue-tsc`).

---

## 10. Breaking Change

**Tidak ada.** Tidak ada perubahan endpoint, DTO, response JSON, atau status code (kecuali munculnya **kemungkinan** status `429 Too Many Requests` baru dari rate limiter — ini bukan breaking change terhadap kontrak yang ada, melainkan status code baru yang hanya muncul saat limit terlampaui, sesuai standar HTTP). Header baru `X-Request-ID` ditambahkan ke response — aditif, tidak mengubah/menghapus apa pun yang sudah ada.

---

## 11. Verification Result

```
Backend:
go build ./...   → OK
go vet ./...     → OK
go test ./...    → ok backend/internal/handler
                   ok backend/internal/middleware (10 test baru: rate limit x3, logging x2, plus semua test lama)
                   ok backend/internal/service

Frontend:
npm run build    → vue-tsc -b (full type-check) → OK, vite build → OK
                   (1 warning CSS pre-existing, tidak terkait perubahan ini)
```

---

## 12. Remaining Technical Debt

- **Rate limit belum diverifikasi di traffic production sungguhan** — angka 20 req/s / burst 40 adalah estimasi masuk akal berdasarkan pola pemakaian yang teramati di kode (bukan diukur dari traffic real), perlu dipantau dan disesuaikan.
- **Rate limiter in-memory tidak scale ke multi-instance** — kalau backend suatu saat di-deploy lebih dari 1 instance di belakang load balancer, tiap instance akan punya bucket terpisah (efektifnya limit jadi `N × limit` untuk N instance). `RateLimiterStore` sengaja dibuat sebagai interface supaya bisa diganti backend terdistribusi (Redis, dsb.) tanpa mengubah pemanggilnya, tapi implementasi terdistribusi itu sendiri belum dibuat.
- **Email-warning `fmt.Printf` belum dimigrasi ke structured logging** — `admin_school_member_import_service.go`, `school_registration_request_service.go`, `school_member_invitation_service.go` masih pakai `fmt.Printf("[Email Warning] ...")`. Sengaja tidak disentuh sprint ini karena scope target 4 secara eksplisit tentang *request* logging, bukan seluruh logging di codebase — tapi ini kandidat migrasi natural ke `slog` di sprint berikutnya untuk konsistensi penuh.
- **Belum ada log level configuration** (mis. via env var) — saat ini selalu `slog.NewJSONHandler(os.Stdout, nil)` dengan level default (Info). Menambah kontrol level via env var adalah perbaikan kecil yang masuk akal untuk sprint operabilitas berikutnya.
- **Test rate-limit tidak menguji token-bucket refill secara presisi** (hanya burst awal dan sweep) — cukup untuk membuktikan mekanisme bekerja, tapi tidak menguji laju pengisian ulang token per detik secara eksak (dianggap cukup karena itu adalah perilaku internal `golang.org/x/time/rate` yang sudah teruji sendiri oleh tim Go).

---

## Catatan

Sama seperti sprint-sprint sebelumnya, file laporan `docs/SPRINT4_FRONTEND_ARCHITECTURE_2026-07-10.md` terdeteksi hilang dari working tree sebelum laporan ini ditulis (muncul sebagai "deleted" di `git status`) — bukan saya yang menghapusnya lewat perintah eksplisit di sesi ini, pola yang sama dengan laporan Sprint 2 dan Sprint 3 sebelumnya. Memberi tahu untuk transparansi.
