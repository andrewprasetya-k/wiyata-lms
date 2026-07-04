# School Member API

MVP warga sekolah digunakan oleh Admin Sekolah untuk mengelola membership siswa,
guru, atau admin sekolah pada sekolah aktif. Admin Sekolah tidak membuka daftar
akun global lintas platform.

`users.deleted_at` menandakan akun global tidak aktif/dihapus dari platform.
`school_users.deleted_at` menandakan akun tersebut dikeluarkan dari sekolah
tertentu. Membership aktif selalu berarti `school_users.deleted_at IS NULL`.

Jika email sudah ada sebagai akun global aktif, akun tersebut dipakai ulang dan
ditautkan ke sekolah aktif tanpa membuka membership sekolah lain.

## List Warga Sekolah

- **URL:** `/admin/school-members`
- **Method:** `GET`
- **Auth:** Admin sekolah pada sekolah aktif
- **Query optional:**
  - `search`: cari nama atau email
  - `role`: `student`, `teacher`, atau `admin`
  - `includeDeleted`: default `false`

Default response hanya berisi membership aktif di sekolah aktif.

```json
{
  "data": [
    {
      "schoolUserId": "...",
      "userId": "...",
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "roles": ["student"],
      "classCodes": ["X-IPA-1"],
      "createdAt": "2026-06-26T10:00:00Z"
    }
  ],
  "totalItems": 1,
  "page": 1,
  "limit": 50,
  "totalPages": 1
}
```

## Tambah Warga Sekolah Manual

- **URL:** `/admin/school-members`
- **Method:** `POST`
- **Auth:** Admin sekolah pada sekolah aktif

Request:

```json
{
  "fullName": "Budi Santoso",
  "email": "budi@siswa.sch.id",
  "password": "InitialPassword123!",
  "role": "student",
  "classCode": "X-IPA-1"
}
```

Behavior:

- `role` hanya boleh `student`, `teacher`, atau `admin`.
- `super_admin` selalu ditolak.
- Jika email belum ada, akun global dibuat memakai password awal.
- Jika email sudah ada sebagai akun global aktif, akun dipakai ulang.
- Jika membership sekolah pernah dihapus, membership dipulihkan dengan
  `school_users.deleted_at = NULL`.
- `classCode` opsional dan hanya berlaku untuk role `student`.
- Setelah operasi sukses, sistem mengirim email best-effort:
  - akun baru menerima email bahwa akun Wiyata sudah dibuat;
  - akun yang sudah ada menerima email bahwa akun tersebut ditambahkan ke
    sekolah aktif.
- Password awal tidak pernah dikirim melalui email. Admin/sekolah harus
  menyampaikan password awal melalui kanal aman di luar Wiyata.
- Kegagalan email tidak menggagalkan pembuatan/penautan warga sekolah.

Response tambahan:

```json
{
  "schoolUserId": "...",
  "userId": "...",
  "fullName": "Budi Santoso",
  "email": "budi@siswa.sch.id",
  "roles": ["student"],
  "classCodes": ["X-IPA-1"],
  "createdAt": "2026-06-26T10:00:00Z",
  "userCreated": true,
  "membershipAction": "created",
  "emailNotification": "account_created"
}
```

## Hapus dari Sekolah Aktif

- **URL:** `/admin/school-members/:schoolUserId`
- **Method:** `DELETE`
- **Auth:** Admin sekolah pada sekolah aktif

Endpoint ini hanya melakukan soft delete membership sekolah:
`school_users.deleted_at = now()`. Akun global di `users` tidak dihapus,
`user_roles` tidak dihapus, dan membership sekolah lain tidak disentuh.

## Pulihkan Membership

- **URL:** `/admin/school-members/:schoolUserId/restore`
- **Method:** `PATCH`
- **Auth:** Admin sekolah pada sekolah aktif

Endpoint ini mengaktifkan kembali membership sekolah aktif dengan
`school_users.deleted_at = NULL`. Import/manual add dengan email yang sama juga
dapat memulihkan membership yang pernah dihapus.

## Preview Import

- **URL:** `/admin/school-members/import/preview`
- **Method:** `POST`
- **Auth:** Admin sekolah pada sekolah aktif
- **Content-Type:** `multipart/form-data`
- **Fields:**
  - `file`: CSV dengan kolom `fullName,email,role,classCode`

`classCode` bersifat opsional dan hanya berlaku untuk role `student`.

Response:

```json
{
  "rows": [
    {
      "rowNumber": 2,
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "role": "student",
      "classCode": "X-IPA-1",
      "status": "valid",
      "errors": []
    }
  ],
  "validCount": 1,
  "invalidCount": 0
}
```

## Commit Import

- **URL:** `/admin/school-members/import/commit`
- **Method:** `POST`
- **Auth:** Admin sekolah pada sekolah aktif

Request:

```json
{
  "defaultPassword": "InitialPassword123!",
  "rows": [
    {
      "rowNumber": 2,
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "role": "student",
      "classCode": "X-IPA-1"
    }
  ]
}
```

Behavior:

- All-or-nothing commit.
- `fullName`, `email`, dan `role` wajib ada.
- `role` hanya boleh `student`, `teacher`, atau `admin`.
- `super_admin` selalu ditolak.
- Email duplikat dalam file ditolak.
- `defaultPassword` wajib diisi dan hanya dipakai untuk akun baru.
- Jika email sudah ada sebagai akun global aktif, user dipakai ulang.
- Membership `school_users` dibuat untuk sekolah aktif jika belum ada.
- Membership `school_users` yang soft-deleted dipulihkan jika email cocok.
- Role dimasukkan ke `user_roles` tanpa menghapus role lain.
- Jika `classCode` diisi dan role adalah `student`, student dienroll ke kelas
  aktif tersebut.
- Teacher class assignment dan subject assignment tidak dilakukan oleh import ini.
- Setelah transaksi import sukses, email dikirim best-effort hanya untuk baris
  yang berhasil diimpor. Baris gagal atau skipped tidak menerima email.
- Password awal tidak pernah dikirim melalui email.

Response:

```json
{
  "importedCount": 1,
  "skippedCount": 0,
  "failedCount": 0,
  "results": [
    {
      "rowNumber": 2,
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "role": "student",
      "classCode": "X-IPA-1",
      "status": "imported",
      "userCreated": true,
      "membershipAction": "created",
      "emailNotification": "account_created"
    }
  ]
}
```

Nilai `emailNotification`:

- `account_created`: akun global baru dibuat dan email akun dibuat dikirim.
- `added_to_school`: akun global sudah ada dan email penambahan ke sekolah
  dikirim.
- kosong/tidak ada: tidak ada email yang perlu dikirim, misalnya baris skipped.
