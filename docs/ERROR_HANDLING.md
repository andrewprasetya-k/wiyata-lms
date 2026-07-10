# Error Handling Convention

Dokumen ini menjelaskan konvensi resmi penyajian error di frontend Wiyata — kapan pakai toast, kapan pakai inline, dan kenapa. Ditulis berdasarkan audit pola yang **sudah ada** di codebase (bukan dirancang dari nol), lalu diselaraskan agar konsisten.

## Kategori Error

| Kategori | Contoh | UX |
|---|---|---|
| 1. Validation | Field wajib kosong, format salah (sebelum request dikirim) | **Inline**, dekat field/form |
| 2. Permission / Authorization | 401/403 dari API | Tergantung konteks — lihat bagian "Load vs Action" di bawah |
| 3. Not Found | 404 dari API | **Inline**, sebagai content state |
| 4. Business rule violation | 409/400 dari API saat submit (mis. kode duplikat, deadline lewat) | **Toast** |
| 5. Network / API unavailable | Request gagal tanpa response (timeout, offline, 5xx) | Tergantung konteks — lihat bagian "Load vs Action" di bawah |
| 6. Unexpected internal error | Error tak terduga lain | **Toast** + `console.error` |

## Prinsip Utama: Load vs Action

Audit terhadap codebase menunjukkan bahwa pembagian yang **sebenarnya sudah konsisten** dan **berhasil** di sebagian besar halaman bukan murni berdasarkan kategori error (network/permission/dst), melainkan berdasarkan **apa yang sedang terjadi saat error itu muncul**:

### A. Memuat konten halaman/section (GET saat mount atau ganti konteks)

**Selalu inline**, ditampilkan sebagai content-state menggantikan area yang gagal dimuat, disertai tombol "Coba lagi". Ini berlaku **apa pun penyebabnya** — network timeout, 403, 404, atau 500 — karena usernya sedang melihat sebuah section yang butuh data, dan state yang persis di tempat itu jauh lebih actionable daripada toast yang hilang dalam 4 detik sementara section-nya kosong tanpa penjelasan.

Contoh (sudah konsisten di puluhan file): `CommentThread.vue` (`errorMessage` + tombol "Coba lagi"), `AdminAcademicYears.vue` (`academicYearsError`/`termsError`/`subjectsError`/`categoriesError`), `TeacherAssignmentReview.vue` (`errorMessage`), dsb.

**Ini sengaja menyimpang dari rule of thumb "Network → toast"** yang umum dipakai di aplikasi lain — karena audit menunjukkan pola inline+retry yang sudah dipakai luas di app ini justru UX yang lebih baik untuk kegagalan load, dan mengubahnya ke toast akan jadi regresi UX, bukan perbaikan.

### B. Melakukan aksi (submit form, delete, toggle, dsb.)

**Toast**, karena aksi bersifat transient — user baru saja menekan tombol, dan feedback singkat yang muncul lalu hilang sudah cukup. Ini mencakup kategori 4 (business rule violation dari API) dan 6 (unexpected error).

Contoh: `toast.error("Tahun ajaran belum bisa dibuat.")`, `toast.error(getApiError(error))` pada action `submitSubject`/`createAdminClass`/dsb.

### C. Form yang selalu terlihat di halaman (composer, form tambah data)

Untuk **validasi client-side** (field kosong, format salah) sebelum request dikirim ke API — **inline**, dekat form, **bukan toast**. Ini adalah salah satu perbaikan utama Sprint 5 (lihat bagian "Perbaikan Sprint 5").

Alasan: toast untuk validasi form itu error klasik — pesannya muncul sekilas lalu hilang padahal user belum tentu sempat membaca sebelum toast hilang, dan tidak ada asosiasi visual ke field mana yang salah.

## Anti-pattern (jangan lakukan)

- ❌ **Toast untuk validasi form** — `toast.error("Nama wajib diisi.")` sebelum request API. Pakai inline dekat form.
- ❌ **Inline untuk hasil aksi** (delete/toggle/submit) — user sudah pindah fokus dari form, inline error di tempat lama gampang terlewat. Pakai toast.
- ❌ **Toast tanpa konteks** untuk kegagalan load section — user tidak tahu bagian mana yang gagal atau bagaimana retry. Pakai inline + tombol retry.
- ❌ **console.error tanpa user-facing feedback** — kalau error mempengaruhi user, selalu sertai toast atau inline, `console.error` hanya untuk diagnostic developer, bukan pengganti UX.
- ❌ **Pola berbeda tanpa alasan** antar halaman yang punya bentuk interaksi sama (mis. dua form "create" yang satu toast satu inline untuk kasus validasi yang sama) — ikuti konvensi ini kecuali ada alasan UX spesifik yang didokumentasikan di kode.

## Helper

### `frontend/src/utils/errorPresentation.ts`

```ts
export type ErrorCategory =
  | "validation" | "permission" | "not_found"
  | "business_rule" | "network" | "unexpected";

export function classifyApiError(error: unknown): ErrorCategory
```

Mengklasifikasikan error dari axios (berdasarkan `error.response.status`) ke salah satu dari 6 kategori di atas. Dipakai sebagai referensi kategorisasi bila kode baru butuh percabangan berdasarkan jenis error — **bukan** pengganti fungsi resolusi-pesan yang sudah ada per-komponen (`getApiError`, `getCommentErrorMessage`, dsb.), yang sengaja tetap dipertahankan karena masing-masing punya copy pesan kontekstual yang sudah tepat untuk domainnya (mengganti semuanya jadi satu fungsi generik akan menghilangkan nuansa pesan yang sudah baik).

### `frontend/src/components/common/InlineFormError.vue`

Komponen kecil untuk menyeragamkan tampilan inline validation error:

```vue
<InlineFormError :message="myFormError" />
```

Tidak menampilkan apa pun jika `message` kosong. Dipakai di semua form yang diperbaiki Sprint 5 (lihat di bawah) agar styling-nya konsisten (`text-xs font-medium text-[#b42318]`), alih-alih setiap halaman menulis ulang class Tailwind sendiri.

## Contoh Implementasi

```ts
// Validation → inline
const myFormError = ref("");

async function submitMyForm() {
  myFormError.value = "";
  if (!form.value.name.trim()) {
    myFormError.value = "Nama wajib diisi.";
    return;
  }
  // ...lanjut ke API call, kegagalan API tetap via toast
  try {
    await createThing(form.value);
    toast.success("Berhasil dibuat.");
  } catch (error) {
    toast.error(getApiError(error)); // business_rule / unexpected → toast
  }
}
```

```vue
<form @submit.prevent="submitMyForm">
  <input v-model="form.name" />
  <InlineFormError :message="myFormError" />
  <button type="submit">Simpan</button>
</form>
```

## Perbaikan Sprint 5

Audit menemukan **15 kemunculan** `toast.error(...)` yang dipakai untuk validasi client-side murni (pola "jangan ada toast untuk validation form" secara eksplisit dilanggar), tersebar di 6 file:

| File | Jumlah | Ref baru |
|---|---|---|
| `pages/admin/AdminAcademicYears.vue` | 5 (termasuk validasi warna hex) | `academicYearFormError`, `termFormError`, `subjectFormError`, `categoryFormError` |
| `pages/admin/AdminUsers.vue` | 5 (termasuk 1 validasi CSV-import yang di-redirect ke ref `importError` yang **sudah ada**) | `memberFormError` (dipakai bersama oleh tab invite & manual) |
| `pages/admin/AdminClasses.vue` | 2 | `classFormError`, `editFormError` |
| `pages/admin/AdminEnrollments.vue` | 2 (termasuk 1 yang tidak eksplisit di audit awal, ditemukan saat implementasi: "Pilih warga sekolah dengan peran Siswa atau Guru.") | `enrollmentFormError` |
| `pages/teacher/TeacherFeed.vue` | 1 | `composeFormError` |
| `pages/teacher/TeacherAssignmentReview.vue` | 1 | `gradeFormError` |

Semua diganti dari `toast.error("...")` menjadi `xxxFormError.value = "..."` + `<InlineFormError :message="xxxFormError" />` di template, memakai copy pesan **yang sama persis** (tidak ada perubahan wording), dan error direset ke string kosong di awal setiap pemanggilan submit/edit (dan saat berpindah context seperti ganti tab atau ganti baris yang di-edit) supaya tidak ada error basi tersisa.

Kegagalan API (`catch` block) di form-form yang sama **sengaja tidak diubah** — tetap `toast.error(...)` karena itu kategori berbeda (business rule / unexpected, bukan validasi client-side).
