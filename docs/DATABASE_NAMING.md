# Database Naming Convention

Dokumen ini menjelaskan konvensi penamaan kolom di schema database Wiyata, diambil langsung dari `backend/schema.md` (DBML, source of truth skema) — bukan asumsi. Tujuannya supaya developer baru bisa langsung memahami arti setiap kolom tanpa harus menghafal atau menebak.

## Konvensi Umum

Setiap tabel punya **prefix 3-4 huruf** (disingkat dari nama tabelnya) yang dipakai di depan HAMPIR SEMUA nama kolomnya, contoh: tabel `subjects` → prefix `sub_` → kolom `sub_id`, `sub_name`, `sub_code`.

**Kenapa prefix, bukan nama kolom polos (`id`, `name`)?**
1. **Tidak ambigu saat JOIN** — di query SQL mentah dengan banyak JOIN, `sub_id` vs `sub_sch_id` vs `cls_sch_id` langsung jelas tabel asalnya tanpa perlu alias berlapis.
2. **Foreign key self-descriptive** — `mat_scl_id` (di tabel `materials`) langsung menunjukkan ia menunjuk ke `subject_classes.scl_id` hanya dari namanya, tanpa perlu buka schema.
3. **Konsisten dengan gaya project sejak awal** — semua 32 tabel di schema mengikuti pola ini.

### Pengecualian: kolom umum tanpa prefix

Kolom berikut **sengaja tidak diberi prefix** karena maknanya sudah generik dan sama di semua tabel — menambah prefix di sini justru jadi boilerplate tanpa manfaat disambiguasi:

| Kolom | Arti | Dipakai di |
|---|---|---|
| `created_at` | Timestamp dibuat | Hampir semua tabel |
| `updated_at` | Timestamp terakhir diubah | Tabel yang datanya bisa diedit (schools, classes, materials, dsb.) |
| `deleted_at` | Timestamp soft-delete (nullable — `NULL` berarti belum dihapus) | Tabel yang mendukung soft delete (schools, users, school_users, classes, materials, dsb.) |
| `is_active` | Boolean status aktif | academic_years, terms, classes |
| `is_public` | Boolean visibilitas | medias |
| `is_read` | Boolean status baca | notifications |
| `created_by` / `assessed_by` / `reviewed_by` | FK ke `users.usr_id` yang melakukan aksi | classes, feeds, assignments, assessments, dsb. |
| `joined_at` / `left_at` | Timestamp masuk/keluar (enrollments, chat_room_members) | enrollments, chat_room_members |

## Daftar Prefix

| Prefix | Kepanjangan | Tabel | Contoh Kolom |
|---|---|---|---|
| `sch_` | **sch**ool | `schools` | `sch_id`, `sch_name`, `sch_code`, `sch_logo` |
| `srr_` | **s**chool **r**egistration **r**equest | `school_registration_requests` | `srr_id`, `srr_school_name`, `srr_status`, `srr_reviewed_by` |
| `inv_` | **inv**itation | `invitations` | `inv_id`, `inv_school_id`, `inv_token_hash`, `inv_expires_at` |
| `acy_` | **a**cademic **y**ear | `academic_years` | `acy_id`, `acy_sch_id`, `acy_name` |
| `trm_` | **t**e**rm** | `terms` | `trm_id`, `trm_acy_id`, `trm_name` |
| `usr_` | **us**e**r** | `users` | `usr_id`, `usr_nama_lengkap`, `usr_email`, `usr_password` |
| `scu_` | **s**chool + **u**ser (school membership) | `school_users` | `scu_id`, `scu_usr_id`, `scu_sch_id` |
| `rol_` | **rol**e | `roles` | `rol_id`, `rol_name` |
| `urol_` | **u**ser **rol**e (assignment) | `user_roles` | `urol_id`, `urol_scu_id`, `urol_rol_id` |
| `med_` | **med**ia | `medias` | `med_id`, `med_sch_id`, `med_file_url`, `med_owner_type` |
| `att_` | **att**achment | `attachments` | `att_id`, `att_source_type`, `att_source_id`, `att_med_id` |
| `sub_` | **sub**ject | `subjects` | `sub_id`, `sub_sch_id`, `sub_code`, `sub_color` |
| `cls_` | **cl**a**ss** | `classes` | `cls_id`, `cls_sch_id`, `cls_trm_id`, `cls_code` |
| `scl_` | **s**ubject **cl**ass (mata pelajaran per kelas) | `subject_classes` | `scl_id`, `scl_cls_id`, `scl_sub_id`, `scl_scu_id` |
| `enr_` | **enr**ollment | `enrollments` | `enr_id`, `enr_scu_id`, `enr_cls_id`, `enr_role` |
| `mat_` | **mat**erial | `materials` | `mat_id`, `mat_sch_id`, `mat_scl_id`, `mat_types` |
| `map_` | **ma**terial **p**rogress | `material_progress` | `map_id`, `map_usr_id`, `map_mat_id`, `map_status` |
| `fds_` | **f**ee**ds** | `feeds` | `fds_id`, `fds_sch_id`, `fds_cls_id`, `fds_content` |
| `cmn_` | **c**o**m**me**n**t | `comments` | `cmn_id`, `cmn_source_type`, `cmn_source_id`, `cmn_usr_id` |
| `asc_` | **as**signment **c**ategory | `assignment_categories` | `asc_id`, `asc_sch_id`, `asc_name` |
| `asg_` | **as**si**g**nment | `assignments` | `asg_id`, `asg_sch_id`, `asg_scl_id`, `asg_deadline` |
| `sbm_` | **s**u**bm**ission | `submissions` | `sbm_id`, `sbm_asg_id`, `sbm_usr_id` |
| `asm_` | **as**sess**m**ent | `assessments` | `asm_id`, `asm_sbm_id`, `asm_score`, `asm_feedback` |
| `asw_` | **as**sessment **w**eight | `assessments_weights` | `asw_id`, `asw_sub_id`, `asw_asc_id`, `asw_weight` |
| `log_` | **log** | `logs` | `log_id`, `log_sch_id`, `log_usr_id`, `log_action` |
| `ntf_` | **n**o**t**i**f**ication | `notifications` | `ntf_id`, `ntf_usr_id`, `ntf_type`, `ntf_related_id` |
| `room_` | chat **room** (tanpa singkatan — nama tabel sudah pendek) | `chat_rooms` | `room_id`, `room_sch_id`, `room_type`, `room_ref_id` |
| `crm_` | **c**hat **r**oom **m**ember | `chat_room_members` | `crm_id`, `crm_room_id`, `crm_usr_id`, `crm_role` |
| `msg_` | **m**e**s**sa**g**e | `chat_messages` | `msg_id`, `msg_room_id`, `msg_content`, `msg_reply_to` |
| `cat_` | **c**hat **at**tachment | `chat_attachments` | `cat_id`, `cat_msg_id`, `cat_med_id` |
| `rct_` | **r**ead re**c**eip**t** | `chat_read_receipts` | `rct_id`, `rct_room_id`, `rct_usr_id`, `last_read_msg_id` |
| `snt_` | **s**tude**nt** note | `student_notes` | `snt_id`, `snt_sch_id`, `snt_usr_id`, `snt_mat_id` |

## Catatan Khusus

- **`chat_rooms` pakai `room_`, bukan `crm_`** — `crm_` justru dipakai tabel anaknya (`chat_room_members`). Ini satu-satunya prefix yang bukan singkatan 3-4 huruf standar; kemungkinan sengaja dipakai karena "room" sudah cukup pendek dan jelas tanpa disingkat lebih jauh, sementara `crm_` dicadangkan untuk child table-nya supaya tidak bentrok penamaan.
- **`cat_` di `chat_attachments`** — jangan tertukar dengan konsep "category" (yang prefix-nya `asc_` untuk assignment category). `cat_` di sini murni singkatan "chat attachment".
- **Multi-tenant scoping** — hampir semua tabel yang datanya milik satu sekolah punya kolom `<prefix>_sch_id` (mis. `sub_sch_id`, `cls_sch_id`, `mat_sch_id`) yang mereferensikan `schools.sch_id`. Ini kolom PALING PENTING untuk tenant-isolation (lihat Sprint 1 — seluruh IDOR yang ditemukan berakar dari query yang lupa memfilter kolom ini).
- **Composite index untuk polymorphic reference** — `source_type`/`source_id` (di `attachments` dan `comments`) dan `owner_type`/`owner_id` (di `medias`) adalah pola "polymorphic foreign key" (satu kolom ID bisa menunjuk ke tabel berbeda-beda tergantung nilai `*_type`-nya) — karena itu tidak punya `ref:` FK constraint eksplisit di DBML, dan butuh composite index manual (sudah ditambahkan di Sprint 3 untuk `attachments`/`comments`).
