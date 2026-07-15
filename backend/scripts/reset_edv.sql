-- =============================================================================
-- Wiyata — Reset data untuk schema "edv"
-- =============================================================================
--
-- Tujuan: mengosongkan SELURUH data aplikasi pada schema "edv" agar
-- backend/scripts/seed_edv.sql bisa dijalankan berulang kali dari kondisi
-- bersih (setara kondisi tepat setelah migration selesai, sebelum data apa
-- pun di-insert) tanpa gagal karena duplicate key / duplikasi role, dsb.
--
-- Yang TIDAK dilakukan file ini:
--   - Tidak DROP TABLE / DROP SCHEMA / DROP TYPE (enum) / DROP FUNCTION / DROP
--     EXTENSION.
--   - Tidak mengubah migration atau struktur tabel apa pun.
--   - Tidak menyentuh schema "public" maupun "silk" — keduanya adalah schema
--     milik aplikasi LAIN yang kebetulan berbagi database Postgres yang sama
--     (server ini bukan database khusus Wiyata). Hanya schema "edv" yang
--     dipakai backend Wiyata (lihat "TableName()" pada setiap model Go dan
--     full_backup.sql), jadi hanya schema itu yang boleh dikosongkan di sini.
--
-- Kenapa TRUNCATE, bukan DELETE:
--   - Lebih cepat: TRUNCATE tidak men-scan baris satu per satu dan tidak
--     menghasilkan WAL sebesar DELETE untuk tabel berukuran besar.
--   - Tidak butuh ORDER BY dependency manual: ketika beberapa tabel di-
--     TRUNCATE dalam SATU statement (seperti di bawah), Postgres otomatis
--     memvalidasi seluruh relasi FK di antara tabel-tabel yang disebutkan
--     sekaligus, jadi urutan penulisan nama tabel di bawah tidak memengaruhi
--     hasil (murni disusun per kelompok untuk keterbacaan).
--   - CASCADE ditambahkan sebagai jaring pengaman: seandainya ada tabel yang
--     mereferensikan salah satu tabel di bawah tapi lupa disertakan dalam
--     daftar, CASCADE akan ikut mengosongkannya alih-alih membuat statement
--     ini gagal karena FK violation. Berdasarkan audit terhadap
--     full_backup.sql, seluruh foreign key pada schema "edv" hanya mengarah
--     ke tabel LAIN yang juga ada di schema "edv" (tidak ada FK dari
--     "public"/"silk" ke "edv"), jadi CASCADE di sini dijamin tidak akan
--     merambat keluar schema "edv".
--
-- Soal RESTART IDENTITY:
--   - Seluruh primary key di schema "edv" bertipe uuid dengan default
--     gen_random_uuid() (lihat schema.md) — tidak ada satu pun kolom
--     serial/identity/sequence di schema ini. RESTART IDENTITY tetap
--     disertakan sesuai pola yang diminta, tapi secara fungsional tidak
--     berefek apa pun untuk tabel-tabel ini (tidak ada sequence yang perlu
--     di-reset).
--
-- Urutan dependency (audit, untuk referensi — bukan urutan eksekusi karena
-- semua tabel di-TRUNCATE bersamaan dalam satu statement):
--   schools, roles
--     -> school_registration_requests, invitations, users
--       -> school_users
--         -> user_roles
--         -> subject_classes (lewat school_users sebagai guru)
--   academic_years -> terms -> classes
--     -> enrollments, subject_classes, feeds
--       -> chat_rooms (room_ref_id ke classes) -> chat_room_members, chat_messages
--                                                  -> chat_attachments, chat_read_receipts
--   subjects -> subject_classes, assessments_weights
--   assignment_categories -> assignments, assessments_weights
--   subject_classes -> materials, assignments
--     materials -> material_progress, student_notes, attachments
--     assignments -> submissions -> assessments
--   medias -> attachments, chat_attachments
--   (comments, notifications, logs bersifat generik/polymorphic, tidak
--    memiliki child table)
--
-- Cara pakai:
--   psql "<connection-string-supabase>" -f backend/scripts/reset_edv.sql
--   psql "<connection-string-supabase>" -f backend/scripts/seed_edv.sql
-- =============================================================================

TRUNCATE TABLE
    -- RBAC & sekolah
    "edv"."roles",
    "edv"."schools",
    "edv"."school_registration_requests",
    "edv"."invitations",
    "edv"."users",
    "edv"."school_users",
    "edv"."user_roles",

    -- struktur akademik
    "edv"."academic_years",
    "edv"."terms",
    "edv"."classes",
    "edv"."subjects",
    "edv"."subject_classes",
    "edv"."enrollments",

    -- konten belajar
    "edv"."materials",
    "edv"."material_progress",
    "edv"."student_notes",
    "edv"."medias",
    "edv"."attachments",

    -- tugas & penilaian
    "edv"."assignment_categories",
    "edv"."assignments",
    "edv"."submissions",
    "edv"."assessments",
    "edv"."assessments_weights",

    -- interaksi sosial
    "edv"."feeds",
    "edv"."comments",
    "edv"."notifications",

    -- chat
    "edv"."chat_rooms",
    "edv"."chat_room_members",
    "edv"."chat_messages",
    "edv"."chat_attachments",
    "edv"."chat_read_receipts",

    -- audit
    "edv"."logs"
RESTART IDENTITY CASCADE;
