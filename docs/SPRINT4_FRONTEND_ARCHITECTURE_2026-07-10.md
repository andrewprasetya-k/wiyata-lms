# Sprint 4 — Frontend Architecture Refactor

**Tanggal:** 2026-07-10
**Scope:** Structural refactor murni — CommentThread unifikasi, split assignment_handler.go, split admin_school_member_import_service.go, dan dekomposisi 3 god-component frontend. Tidak ada perubahan endpoint, DTO, response, business logic, authorization, database, styling visual, atau UX flow.

---

## 1. Audit Summary

### 1.1 FeedComments.vue vs DiscussionComments.vue

| | FeedComments.vue | DiscussionComments.vue |
|---|---|---|
| Baris | 350 | 376 |
| Props | `post: FeedPost` | `sourceType, sourceId, title?, placeholder?, emptyText?` |
| API call | `getFeedComments/createFeedComment/deleteFeedComment` (services/feed.ts) | `getComments/createComment/deleteComment` (services/comments.ts) |
| UX | Collapsible, default collapsed, badge count sebelum expand | Selalu terlihat, card besar |
| Copy | "Komentar hanya untuk feed kelas.", pesan error versi "feed" | "Diskusi terlihat oleh peserta...", pesan error versi "diskusi" |
| Race-guard reload | Tidak ada | Ada (`loadRequestId` + `watch([sourceType, sourceId])`) |

**Temuan kunci (dikonfirmasi lewat pembacaan kode, bukan asumsi):** `getFeedComments(feedId)` di `services/feed.ts` ternyata secara literal memanggil endpoint yang **sama persis** dengan `getComments({sourceType:'feed', sourceId: feedId})` — sama URL (`/comments`), sama query param (`type`/`id`). `FeedComment` dan `CommentItem` juga struktural identik. Duplikasi kode nyata: **~90% logic (state, load/submit/delete, optimistic UI, error handling) identik**; ~10% yang berbeda adalah UX shell (collapsible vs static) dan copy text.

**Kesimpulan:** cukup props (tidak perlu slot terpisah, tidak perlu composable terpisah karena hanya ada satu consumer akhir — component itu sendiri). Variant "feed" vs "diskusi" cukup diturunkan dari `sourceType === 'feed'` tanpa prop tambahan, karena keduanya selalu 1:1 di semua pemakaian saat ini.

### 1.2 assignment_handler.go

- 1055 baris, 32 fungsi dalam satu file.
- Sudah cukup ter-modularisasi jadi private method (`authorize*`, `getSchoolContext`, `mapAsgToResponse`, dst.) — masalahnya murni ukuran file, bukan hilangnya abstraksi.
- Kategori jelas: HTTP handler (route-bound), authorization helper (10 fungsi, ~192 baris), response mapper (2 fungsi, ~49 baris).

### 1.3 admin_school_member_import_service.go

- 755 baris, 27 fungsi.
- Kategori jelas: orchestration (PreviewCSV/Commit/ListMembers/AddMember/RemoveMember/RestoreMember + notifikasi email), CSV parsing (parseCSV/csvValue/isEmptyCSVRecord), validasi (validateRows/existingClassCodes), enrollment & member-creation tx-scoped (findOrCreateUser/findOrCreateSchoolUser/findRoleID/ensureRole/findClassIDByCode/ensureActiveStudentEnrollment/classCodesBySchoolUser*/mapSchoolMember).
- Tidak ada fungsi "rollback helper" terpisah — rollback sepenuhnya implisit lewat `db.Transaction()`, jadi tidak ada yang perlu diekstrak untuk kategori itu.

### 1.4 Frontend God Components (top 3 by line count)

| File | Baris | computed | watch | Kenapa god component |
|---|---|---|---|---|
| `ChatWorkspace.vue` | 2819 | 11 | 1 | Menggabungkan: room list, create-conversation (DM+grup), group info/rename/add/remove-member/leave, message pagination+polling, websocket lifecycle, drag-drop upload, lightbox — 97 fungsi dalam satu file |
| `AdminUsers.vue` | 1384 | 3 | 1 | Menggabungkan: member list+role sync, invite-by-link flow, manual member creation, CSV/Excel bulk import (termasuk konversi xlsx→csv) |
| `AdminAcademicYears.vue` | 1383 | 9 | 0 | Menggabungkan 5 entitas CRUD berbeda dalam satu file: academic year, term, subject (+color picker), assignment category, assessment weight |

`ChatWorkspace.vue` diperlakukan berbeda dari 2 file lain karena jauh lebih besar dan bersifat realtime/websocket (state saling terkait erat) — atas konfirmasi Anda, scope untuk file ini dibatasi ke ekstraksi moderate (2 modal + util), bukan dekomposisi penuh termasuk message-list/realtime core.

---

## 2. Root Cause

- **CommentThread**: dua component dibuat terpisah di waktu berbeda untuk konteks UI yang berbeda (inline-feed vs standalone-discussion) tanpa pernah disadari bahwa backend API-nya sudah identik sejak awal.
- **assignment_handler.go / admin_school_member_import_service.go**: pertumbuhan organik — setiap fitur baru ditambah sebagai method baru di file yang sama, tanpa pernah dipisah berdasarkan tanggung jawab.
- **God components frontend**: pola yang sama — form/modal/tab baru ditambahkan langsung ke file yang sudah ada alih-alih menjadi component baru, karena tidak ada aturan/konvensi ukuran maksimum file di project ini.

---

## 3. Files Changed

### Backend

| File | Perubahan |
|---|---|
| `internal/handler/assignment_handler.go` | 1055 → 814 baris |
| `internal/handler/assignment_handler_auth.go` | **Baru** — 206 baris, 10 fungsi authorization/context helper |
| `internal/handler/assignment_handler_mapper.go` | **Baru** — 58 baris, 2 fungsi response mapper |
| `internal/service/admin_school_member_import_service.go` | 755 → 408 baris |
| `internal/service/admin_school_member_import_csv.go` | **Baru** — 80 baris, CSV parsing |
| `internal/service/admin_school_member_import_validation.go` | **Baru** — 110 baris, validasi baris import |
| `internal/service/admin_school_member_import_enrollment.go` | **Baru** — 194 baris, helper tx-scoped enrollment/member |

### Frontend

| File | Perubahan |
|---|---|
| `components/comments/CommentThread.vue` | **Baru** — 530 baris, menggantikan FeedComments.vue + DiscussionComments.vue |
| `components/feed/FeedComments.vue` | **Dihapus** (350 baris) |
| `components/discussion/DiscussionComments.vue` | **Dihapus** (376 baris) |
| `services/feed.ts` | Hapus 3 fungsi (`getFeedComments`/`createFeedComment`/`deleteFeedComment`) yang jadi dead code setelah unifikasi |
| `types/feed.ts` | Hapus 2 type (`FeedComment`/`CreateFeedCommentPayload`) yang jadi dead code |
| `pages/student/StudentFeed.vue`, `pages/teacher/TeacherFeed.vue` | Ganti `<FeedComments>` → `<CommentThread source-type="feed">` |
| `pages/student/StudentMaterialDetail.vue`, `StudentAssignmentDetail.vue`, `pages/teacher/TeacherMaterialDetail.vue`, `TeacherAssignmentDetail.vue` | Ganti `<DiscussionComments>` → `<CommentThread>` |
| `components/chat/ChatWorkspace.vue` | 2819 → 1962 baris |
| `components/chat/ChatCreateConversationModal.vue` | **Baru** — 446 baris |
| `components/chat/ChatGroupInfoModal.vue` | **Baru** — 496 baris |
| `utils/chatDisplay.ts` | **Baru** — 44 baris (getInitials, isDirectMessageRoom, isCustomGroupRoom, roomDisplayName, resolveChatError) |
| `pages/admin/AdminUsers.vue` | 1384 → 1255 baris |
| `utils/schoolMemberImportFile.ts` | **Baru** — 135 baris (CSV/Excel template & konversi, murni pure function) |
| `pages/admin/AdminAcademicYears.vue` | 1383 → 1364 baris |
| `utils/color.ts` | Tambah 4 fungsi (normalizeSubjectColor, isValidSubjectColor, toColorPickerValue, subjectDisplayColor) |

---

## 4. Detail Perubahan per File

### CommentThread.vue

Satu component dengan varian internal berdasarkan `sourceType === 'feed'` (`isFeedVariant`), bukan prop terpisah — karena kedua UI-shape (collapsible-inline vs standalone-card) selalu berkorespondensi 1:1 dengan sourceType di semua pemakaian saat ini, menambah prop kedua hanya akan menduplikasi informasi yang sama.

Semua copy text yang berbeda per-konteks (pesan error 403/404, fallback pesan gagal load, empty text, visibility caption, placeholder, judul) dipertahankan **persis sama** lewat percabangan `isFeedVariant.value ? "..." : "..."` — tidak ada teks yang berubah dari versi asli.

`emit("comment-count-change", feedId, count)` (versi Feed) disederhanakan jadi `emit("count-change", count)` (menyamai versi Discussion yang sudah lebih sederhana) — 2 call site di `StudentFeed.vue`/`TeacherFeed.vue` diadaptasi memakai closure: `@count-change="(count) => updatePostCommentCount(post.feedId, count)"`, sehingga perilaku observable (update `commentCount` per post di list) tetap identik.

### assignment_handler.go split

`assignment_handler_auth.go`: `getSchoolContext`, `getActiveRoles`, `hasActiveRole`, `validateRequestSchool`, `authorizeUserForSubjectClassAccess`, `authorizeStudentForSubjectClass`, `authorizeAssignmentMutation`, `authorizeStudentForSubmission`, `authorizeTeacherForSubmission`, `authorizeTeacherForSubjectClass` — dipindah verbatim (copy-paste murni, tanpa mengubah satu baris logic pun), karena Go mengizinkan satu package tersebar di banyak file tanpa memengaruhi behavior sama sekali.

`assignment_handler_mapper.go`: `mapAsgToResponse`, `mapMySubmissionToResponse` — dipindah verbatim.

### admin_school_member_import_service.go split

`admin_school_member_import_csv.go`: `parseCSV`, `csvValue`, `isEmptyCSVRecord`.
`admin_school_member_import_validation.go`: `validateRows`, `existingClassCodes`, var `allowedSchoolMemberImportRoles`.
`admin_school_member_import_enrollment.go`: `findOrCreateUser`, `findOrCreateSchoolUser`, `findRoleID`, `ensureRole`, `findClassIDByCode`, `ensureActiveStudentEnrollment`, `classCodesBySchoolUser`, `classCodesBySchoolUserTx`, `mapSchoolMember`.

Semua verbatim copy-paste antar file dalam package yang sama — nol perubahan logic.

### ChatWorkspace.vue → ChatCreateConversationModal.vue + ChatGroupInfoModal.vue

**ChatCreateConversationModal.vue**: modal tab DM+Grup, membawa semua state lokal (`dmSearch`, `dmResults`, `groupRoomName`, `memberResults`, dst.) dan logic (`loadDMTargets`, `submitDirectMessage`, `loadChatMembers`, `submitCreateGroup`). Kontrak: `v-model:open`, prop `initial-tab`/`current-user-id`, emit `dm-opened(room)`/`group-created(room)`. Parent (`ChatWorkspace.vue`) tinggal punya handler tipis (`handleDmOpened`, `handleGroupCreated`) yang melakukan efek samping tingkat-parent (`refreshRooms`, set `selectedRoom`, `loadLatestMessages`, toast) — perilaku observable identik dengan sebelumnya, cuma pindah tempat.

**ChatGroupInfoModal.vue**: panel info grup (rename, tambah/keluarkan anggota, keluar grup), sama pola — state+logic pindah ke child, parent cuma bereaksi lewat emit (`renamed`, `members-changed`, `left`).

**utils/chatDisplay.ts**: `getInitials`, `isDirectMessageRoom`, `isCustomGroupRoom`, `roomDisplayName`, `resolveChatError` — fungsi pure yang dipakai baik oleh `ChatWorkspace.vue` sendiri (room list, header, composer error) maupun kedua modal baru, jadi diekstrak ke util bersama alih-alih di-duplikasi 3x.

### AdminUsers.vue → utils/schoolMemberImportFile.ts

`downloadTemplate`, `downloadExcelTemplate`, `csvEscape`, `toCsv`, `normalizeImportHeader`, `isExcelFile`, `convertXlsxToCsvFile`, `importTemplateRows` — seluruhnya fungsi **pure** (tidak menyentuh reactive state component sama sekali, hanya menerima/mengembalikan value biasa atau memicu file download), sehingga ekstraksi ini adalah risiko paling rendah di seluruh sprint ini.

### AdminAcademicYears.vue → utils/color.ts

`normalizeSubjectColor`, `isValidSubjectColor`, `toColorPickerValue`, `subjectDisplayColor` — pure function terkait warna, digabung ke `utils/color.ts` yang sudah ada (sudah berisi `getSubjectColor`/`resolveSubjectColor`) alih-alih membuat file baru, karena secara konsep memang satu domain.

`normalizeWeightInput`/`parseWeightValue`/`isWeightInputInvalid`/`formatWeight` (assessment-weight helper) **sengaja tidak diekstrak** — hanya dipakai satu tempat, tidak ada rumah konseptual yang jelas untuk digabung, ekstraksi di sini hanya jadi file-shuffling tanpa manfaat nyata (dilaporkan sebagai bagian dari Remaining Technical Debt, bukan diperbaiki tanpa alasan kuat).

---

## 5. Maintainability Improvement

| Metrik | Sebelum | Sesudah |
|---|---|---|
| FeedComments.vue + DiscussionComments.vue | 726 baris, ~90% logic duplikat | 530 baris, 0% duplikat (1 component) |
| assignment_handler.go | 1055 baris, 1 file | 814 + 206 + 58 = 1078 baris, 3 file (masing-masing bertanggung jawab tunggal) |
| admin_school_member_import_service.go | 755 baris, 1 file | 408 + 80 + 110 + 194 = 792 baris, 4 file |
| ChatWorkspace.vue | 2819 baris, 1 file | 1962 + 446 + 496 + 44 = 2948 baris, 4 file |
| AdminUsers.vue | 1384 baris, 1 file | 1255 + 135 = 1390 baris, 2 file |
| AdminAcademicYears.vue | 1383 baris, 1 file | 1364 + (+31 di color.ts) baris, tetap 1 file utama + util bersama |

Total baris kode naik sedikit (overhead deklarasi import/export per file baru) — ini normal dan diharapkan untuk refactor modularisasi murni; yang berkurang bukan jumlah baris total tapi **ukuran unit yang harus dibaca sekaligus** untuk memahami satu tanggung jawab, dan (untuk CommentThread) **duplikasi logic yang harus di-maintain dua kali**.

---

## 6. Regression Impact

- **Backend**: perubahan murni pemindahan fungsi antar file dalam package yang sama (`package handler`, `package service`) — nol perubahan logic, nol perubahan signature publik, nol perubahan routing. `go build`, `go vet`, `go test ./...` seluruhnya lulus tanpa perubahan hasil test.
- **CommentThread**: perilaku dan tampilan untuk kedua varian (feed collapsible, discussion standalone) dipertahankan 1:1 berdasarkan pembacaan baris-demi-baris kedua component asli — termasuk copy text, class Tailwind, urutan optimistic-UI, race-guard reload. Emit event disederhanakan tapi efeknya di parent (update `commentCount` per post) tetap sama.
- **ChatWorkspace.vue**: risiko regresi **tertinggi di sprint ini** karena state realtime yang saling terkait erat — mitigasi yang dilakukan: pembacaan baris-demi-baris tiap fungsi yang dipindah sebelum menulis ulang jadi child component, verifikasi `npm run build` (termasuk `vue-tsc` type-check penuh) lulus tanpa error setelah tiap tahap edit, dan verifikasi tidak ada reference yang tertinggal (`grep` untuk semua state/fungsi yang dipindah, hasil nol match tersisa di parent). **Belum ada verifikasi visual manual di browser** (butuh sesi login+role tertentu untuk mengakses fitur chat) — direkomendasikan QA manual sebelum deploy untuk alur: buka percakapan baru (DM & Grup), lihat info grup, ubah nama grup, tambah/keluarkan anggota, keluar dari grup.
- **AdminUsers.vue / AdminAcademicYears.vue**: risiko sangat rendah — hanya fungsi pure (tidak menyentuh reactive state) yang dipindah, dibuktikan lewat build TypeScript yang lulus (type-checker akan menangkap jika ada reference yang salah).

---

## 7. Breaking Change

**Tidak ada.** Tidak ada perubahan endpoint, API contract, DTO, response, business logic, authorization, atau struktur database. Perubahan yang terlihat dari luar aplikasi (user-facing): **nol** — seluruh tampilan dan alur interaksi dipertahankan identik by design. Satu-satunya perubahan yang terlihat dari sisi development: penghapusan 3 fungsi (`getFeedComments`/`createFeedComment`/`deleteFeedComment`) dan 2 type (`FeedComment`/`CreateFeedCommentPayload`) dari codebase karena sudah tidak dipakai lagi — bukan breaking change karena tidak ada konsumen lain yang tersisa (diverifikasi lewat grep sebelum dihapus).

---

## 8. Verification Result

```
Backend:
go build ./...   → OK
go vet ./...     → OK
go test ./...    → ok backend/internal/handler
                   ok backend/internal/middleware
                   ok backend/internal/service

Frontend:
npm run build    → vue-tsc -b (type-check penuh) → OK, vite build → OK
                   (1 warning CSS pre-existing yang sudah dikonfirmasi ada sejak sprint sebelumnya, tidak terkait perubahan ini)
```

---

## 9. Remaining Technical Debt

- **ChatWorkspace.vue masih 1962 baris** setelah ekstraksi moderate — message list rendering, websocket lifecycle, polling, drag-drop upload, dan lightbox masih menyatu di satu file. Dekomposisi lebih dalam (misal `ChatMessageList.vue`, `useChatSocket` composable) memungkinkan tapi berisiko lebih tinggi karena menyentuh state reaktif inti; disepakati di luar scope sprint ini.
- **AdminUsers.vue masih menggabungkan 3 alur berbeda** (member list+role, invite-by-link, manual create, CSV/Excel import) dalam satu file — pemisahan jadi child component per-tab akan lebih tuntas tapi butuh scope tersendiri.
- **AdminAcademicYears.vue masih menggabungkan 5 entitas CRUD** (academic year, term, subject, category, weight) — kandidat kuat untuk dipecah jadi component per-tab (`PeriodeTab.vue`/`MapelTab.vue`) di sprint mendatang.
- **`normalizeWeightInput`/`parseWeightValue`/`isWeightInputInvalid`/`formatWeight`** di `AdminAcademicYears.vue` sengaja tidak diekstrak (lihat bagian 4) — kandidat util jika ada consumer kedua di masa depan.
- **Pola non-atomic serupa di `material_service.go`** (ditemukan Sprint 3, belum diperbaiki — di luar scope Sprint 3 dan Sprint 4).
- **Migration tooling** masih belum ada (dikonfirmasi Sprint 3, sengaja di-skip atas keputusan Anda).

---

## Catatan

Sama seperti sprint sebelumnya, file laporan `docs/SPRINT3_BACKEND_RELIABILITY_2026-07-10.md` yang saya buat di sesi ini terdeteksi hilang dari working tree (muncul sebagai "deleted" di `git status`) sebelum laporan ini ditulis — bukan saya yang menghapusnya lewat perintah eksplisit, kemungkinan besar terkait proses auto-commit/checkpoint IDE yang sudah berulang kali terdeteksi di sesi-sesi sebelumnya. Memberi tahu untuk transparansi.
