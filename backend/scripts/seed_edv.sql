-- =============================================================================
-- Wiyata — Seed data untuk schema "edv"
-- =============================================================================
--
-- Dibuat berdasarkan struktur tabel di full_backup.sql (schema-only dump).
-- File backup itu TIDAK berisi data sama sekali, jadi seluruh data di bawah ini
-- adalah data contoh (fake/sample) yang saya susun mengikuti struktur tabel,
-- tipe kolom, dan foreign key yang ada di "edv" — bukan hasil ekstraksi dari
-- data asli manapun.
--
-- Asumsi: tabel-tabel "edv" SUDAH ada di database (tidak ada CREATE TABLE di
-- sini), termasuk kolom PK yang sudah punya DEFAULT gen_random_uuid().
--
-- Skrip ini TIDAK meng-insert ID manapun secara eksplisit — semua primary key
-- dibiarkan digenerate otomatis oleh DEFAULT gen_random_uuid() pada kolomnya.
-- Karena banyak baris saling mereferensikan lewat foreign key, skrip ditulis
-- sebagai satu blok PL/pgSQL (DO $$ ... $$) yang menampung tiap ID hasil
-- generate ke variabel lokal lewat "RETURNING ... INTO", lalu variabel itu
-- dipakai untuk INSERT baris anaknya. Urutan insert tetap mengikuti urutan
-- dependency foreign key (parent dulu baru child).
--
-- Password semua user contoh di bawah ini: "Password123!"
-- (hash bcrypt di bawah digenerate langsung dari backend/internal/service/auth_service.go,
-- bcrypt.DefaultCost, jadi valid dipakai untuk login test di aplikasi ini).
--
-- Cara pakai:
--   psql "<connection-string-supabase>" -f backend/scripts/seed_edv.sql
-- =============================================================================

DO $$
DECLARE
    -- roles
    v_role_super_admin_id uuid;
    v_role_admin_id       uuid;
    v_role_teacher_id     uuid;
    v_role_student_id     uuid;

    -- school
    v_school_id uuid;

    -- users
    v_user_superadmin_id uuid;
    v_user_admin_id      uuid;
    v_user_teacher1_id   uuid; -- Andi Wijaya, guru Matematika
    v_user_teacher2_id   uuid; -- Rina Kartika, guru Fisika & B. Indonesia
    v_user_student1_id   uuid; -- Dewi Lestari
    v_user_student2_id   uuid; -- Fajar Ramadhan
    v_user_student3_id   uuid; -- Nadia Putri
    v_user_student4_id   uuid; -- Rizky Maulana
    v_user_student5_id   uuid; -- Salsabila Putri

    -- school_users
    v_scu_superadmin_id uuid;
    v_scu_admin_id      uuid;
    v_scu_teacher1_id   uuid;
    v_scu_teacher2_id   uuid;
    v_scu_student1_id   uuid;
    v_scu_student2_id   uuid;
    v_scu_student3_id   uuid;
    v_scu_student4_id   uuid;
    v_scu_student5_id   uuid;

    -- academic structure
    v_academic_year_id uuid;
    v_term_id          uuid;
    v_class1_id        uuid; -- 10 IPA 1
    v_class2_id        uuid; -- 10 IPA 2

    -- subjects & subject_classes
    v_subject_math_id    uuid;
    v_subject_physics_id uuid;
    v_subject_indo_id    uuid;
    v_subject_class_math_id    uuid; -- Matematika @ 10 IPA 1
    v_subject_class_physics_id uuid; -- Fisika @ 10 IPA 1
    v_subject_class_indo_id    uuid; -- B. Indonesia @ 10 IPA 2

    -- enrollments
    v_enr_teacher1_class1_id uuid;
    v_enr_teacher2_class1_id uuid;
    v_enr_teacher2_class2_id uuid;
    v_enr_student1_class1_id uuid;
    v_enr_student2_class1_id uuid;
    v_enr_student3_class1_id uuid;
    v_enr_student4_class2_id uuid;
    v_enr_student5_class2_id uuid;

    -- assignment categories
    v_category_daily_id uuid;
    v_category_exam_id  uuid;

    -- assignments, submissions
    v_assignment1_id uuid; -- Latihan Aljabar Dasar
    v_assignment2_id uuid; -- UTS Matematika
    v_submission1_id uuid; -- Dewi, sudah dinilai
    v_submission2_id uuid; -- Fajar, belum dinilai

    -- material & media
    v_material1_id uuid;
    v_media1_id    uuid;

    -- feed
    v_feed1_id uuid;

    -- chat
    v_room1_id    uuid; -- room kelas 10 IPA 1
    v_room2_id    uuid; -- DM Andi <-> Dewi
    v_message1_id uuid;

    -- =========================================================================
    -- Perluasan data demo (guru, subject, siswa, materi, tugas tambahan)
    -- Ditambahkan supaya tiap halaman Wiyata (dashboard, subject, grade report,
    -- review tugas) punya data yang cukup kaya untuk demo. Data lama di atas
    -- TIDAK diubah.
    -- =========================================================================

    -- guru baru
    v_user_teacher3_id uuid; -- Yusuf Hidayat, guru Bahasa Inggris
    v_scu_teacher3_id  uuid;

    -- subject baru
    v_subject_eng_id uuid; -- Bahasa Inggris

    -- siswa baru (perluasan roster 10 IPA 1 & 10 IPA 2)
    v_user_student6_id  uuid; -- Muhammad Iqbal (10 IPA 1)
    v_user_student7_id  uuid; -- Putri Ayu Lestari (10 IPA 1)
    v_user_student8_id  uuid; -- Bintang Saputra (10 IPA 1)
    v_user_student9_id  uuid; -- Aulia Rahmah (10 IPA 1)
    v_user_student10_id uuid; -- Agus Setiawan (10 IPA 2)
    v_user_student11_id uuid; -- Kirana Dewi (10 IPA 2)
    v_user_student12_id uuid; -- Reza Firmansyah (10 IPA 2)
    v_user_student13_id uuid; -- Yolanda Safitri (10 IPA 2)

    v_scu_student6_id  uuid;
    v_scu_student7_id  uuid;
    v_scu_student8_id  uuid;
    v_scu_student9_id  uuid;
    v_scu_student10_id uuid;
    v_scu_student11_id uuid;
    v_scu_student12_id uuid;
    v_scu_student13_id uuid;

    -- subject_classes tambahan (melengkapi tiap kelas dengan seluruh mapel)
    v_subject_class_indo_c1_id    uuid; -- B. Indonesia @ 10 IPA 1
    v_subject_class_math_c2_id    uuid; -- Matematika @ 10 IPA 2
    v_subject_class_physics_c2_id uuid; -- Fisika @ 10 IPA 2
    v_subject_class_eng_c1_id     uuid; -- B. Inggris @ 10 IPA 1
    v_subject_class_eng_c2_id     uuid; -- B. Inggris @ 10 IPA 2

    -- materials tambahan (variabel dipakai lagi untuk media/attachment/progress/notes)
    v_material2_id  uuid; -- Persamaan Linear Satu Variabel @ Matematika-10 IPA 1
    v_material3_id  uuid; -- SPLDV @ Matematika-10 IPA 1
    v_material4_id  uuid; -- Persamaan Linear Satu Variabel @ Matematika-10 IPA 2
    v_material5_id  uuid; -- Gerak Lurus Beraturan @ Fisika-10 IPA 1
    v_material6_id  uuid; -- Hukum Newton tentang Gerak @ Fisika-10 IPA 1
    v_material7_id  uuid; -- Gerak Lurus Beraturan @ Fisika-10 IPA 2
    v_material8_id  uuid; -- Hukum Newton tentang Gerak @ Fisika-10 IPA 2
    v_material9_id  uuid; -- Teks Eksposisi @ B. Indonesia-10 IPA 1
    v_material10_id uuid; -- Teks Argumentasi @ B. Indonesia-10 IPA 1
    v_material11_id uuid; -- Teks Eksposisi @ B. Indonesia-10 IPA 2
    v_material12_id uuid; -- Simple Present Tense @ B. Inggris-10 IPA 1
    v_material13_id uuid; -- Descriptive Text @ B. Inggris-10 IPA 1
    v_material14_id uuid; -- Simple Present Tense @ B. Inggris-10 IPA 2
    v_material15_id uuid; -- Descriptive Text @ B. Inggris-10 IPA 2

    v_media2_id  uuid;
    v_media3_id  uuid;
    v_media4_id  uuid;
    v_media5_id  uuid;
    v_media6_id  uuid;
    v_media7_id  uuid;
    v_media8_id  uuid;
    v_media9_id  uuid;
    v_media10_id uuid;
    v_media11_id uuid;
    v_media12_id uuid;
    v_media13_id uuid;
    v_media14_id uuid;
    v_media15_id uuid;

    -- assignments tambahan (2 per subject_class yang tadinya belum punya tugas)
    v_assignment3_id  uuid; -- Latihan Persamaan Linear @ Matematika-10 IPA 2
    v_assignment4_id  uuid; -- Ulangan Harian SPLDV @ Matematika-10 IPA 2
    v_assignment5_id  uuid; -- Latihan Soal Gerak Lurus Beraturan @ Fisika-10 IPA 1
    v_assignment6_id  uuid; -- Ulangan Bab Gerak Lurus @ Fisika-10 IPA 1
    v_assignment7_id  uuid; -- Tugas Analisis Gerak Lurus @ Fisika-10 IPA 2
    v_assignment8_id  uuid; -- Ulangan Hukum Newton @ Fisika-10 IPA 2
    v_assignment9_id  uuid; -- Menulis Teks Eksposisi @ B. Indonesia-10 IPA 1
    v_assignment10_id uuid; -- Ulangan Teks Argumentasi @ B. Indonesia-10 IPA 1
    v_assignment11_id uuid; -- Latihan Menyusun Teks Argumentasi @ B. Indonesia-10 IPA 2
    v_assignment12_id uuid; -- Ulangan Teks Eksposisi @ B. Indonesia-10 IPA 2
    v_assignment13_id uuid; -- Simple Present Tense Exercise @ B. Inggris-10 IPA 1
    v_assignment14_id uuid; -- Descriptive Text Writing Test @ B. Inggris-10 IPA 1
    v_assignment15_id uuid; -- Grammar Practice: Simple Present @ B. Inggris-10 IPA 2
    v_assignment16_id uuid; -- Writing Test: Descriptive Text @ B. Inggris-10 IPA 2

    -- feed & chat tambahan
    v_feed2_id    uuid;
    v_feed3_id    uuid;
    v_feed4_id    uuid;
    v_room3_id    uuid; -- room kelas 10 IPA 2
    v_message2_id uuid;
BEGIN

    -- -----------------------------------------------------------------------
    -- 1. roles (referensi RBAC platform: super_admin, admin, teacher, student)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."roles" ("rol_name") VALUES ('super_admin') RETURNING "rol_id" INTO v_role_super_admin_id;
    INSERT INTO "edv"."roles" ("rol_name") VALUES ('admin')       RETURNING "rol_id" INTO v_role_admin_id;
    INSERT INTO "edv"."roles" ("rol_name") VALUES ('teacher')     RETURNING "rol_id" INTO v_role_teacher_id;
    INSERT INTO "edv"."roles" ("rol_name") VALUES ('student')     RETURNING "rol_id" INTO v_role_student_id;

    -- -----------------------------------------------------------------------
    -- 2. schools
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."schools" ("sch_name", "sch_code", "sch_address", "sch_email", "sch_phone", "sch_website")
    VALUES ('SMA Wiyata Nusantara', 'WIYATA01', 'Jl. Pendidikan No. 1, Jakarta',
            'admin@wiyatanusantara.sch.id', '021-5550101', 'https://wiyatanusantara.sch.id')
    RETURNING "sch_id" INTO v_school_id;

    -- -----------------------------------------------------------------------
    -- 3. users (password semua: "Password123!")
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Budi Santoso', 'superadmin@wiyata.dev', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_superadmin_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Siti Aminah', 'admin@wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_admin_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Andi Wijaya', 'andi.wijaya@wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_teacher1_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Rina Kartika', 'rina.kartika@wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_teacher2_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Dewi Lestari', 'dewi.lestari@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student1_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Fajar Ramadhan', 'fajar.ramadhan@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student2_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Nadia Putri', 'nadia.putri@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student3_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Rizky Maulana', 'rizky.maulana@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student4_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Salsabila Putri', 'salsabila.putri@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student5_id;

    -- -----------------------------------------------------------------------
    -- 4. school_users (keanggotaan user di sekolah)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_superadmin_id, v_school_id) RETURNING "scu_id" INTO v_scu_superadmin_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_admin_id, v_school_id)      RETURNING "scu_id" INTO v_scu_admin_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_teacher1_id, v_school_id)   RETURNING "scu_id" INTO v_scu_teacher1_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_teacher2_id, v_school_id)   RETURNING "scu_id" INTO v_scu_teacher2_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student1_id, v_school_id)   RETURNING "scu_id" INTO v_scu_student1_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student2_id, v_school_id)   RETURNING "scu_id" INTO v_scu_student2_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student3_id, v_school_id)   RETURNING "scu_id" INTO v_scu_student3_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student4_id, v_school_id)   RETURNING "scu_id" INTO v_scu_student4_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student5_id, v_school_id)   RETURNING "scu_id" INTO v_scu_student5_id;

    -- -----------------------------------------------------------------------
    -- 5. user_roles (assign role RBAC ke tiap school_user)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."user_roles" ("urol_scu_id", "urol_rol_id") VALUES
        (v_scu_superadmin_id, v_role_super_admin_id),
        (v_scu_admin_id,      v_role_admin_id),
        (v_scu_teacher1_id,   v_role_teacher_id),
        (v_scu_teacher2_id,   v_role_teacher_id),
        (v_scu_student1_id,   v_role_student_id),
        (v_scu_student2_id,   v_role_student_id),
        (v_scu_student3_id,   v_role_student_id),
        (v_scu_student4_id,   v_role_student_id),
        (v_scu_student5_id,   v_role_student_id);

    -- -----------------------------------------------------------------------
    -- 6. academic_years & terms
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."academic_years" ("acy_sch_id", "acy_name", "is_active")
    VALUES (v_school_id, '2025/2026', true)
    RETURNING "acy_id" INTO v_academic_year_id;

    INSERT INTO "edv"."terms" ("trm_acy_id", "trm_name", "is_active")
    VALUES (v_academic_year_id, 'Semester Ganjil', true)
    RETURNING "trm_id" INTO v_term_id;

    -- -----------------------------------------------------------------------
    -- 7. classes
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."classes" ("cls_sch_id", "cls_trm_id", "cls_code", "cls_title", "cls_desc", "created_by", "is_active")
    VALUES (v_school_id, v_term_id, '10-IPA-1', 'Kelas 10 IPA 1', 'Rombongan belajar 10 IPA 1 tahun ajaran 2025/2026', v_user_admin_id, true)
    RETURNING "cls_id" INTO v_class1_id;

    INSERT INTO "edv"."classes" ("cls_sch_id", "cls_trm_id", "cls_code", "cls_title", "cls_desc", "created_by", "is_active")
    VALUES (v_school_id, v_term_id, '10-IPA-2', 'Kelas 10 IPA 2', 'Rombongan belajar 10 IPA 2 tahun ajaran 2025/2026', v_user_admin_id, true)
    RETURNING "cls_id" INTO v_class2_id;

    -- -----------------------------------------------------------------------
    -- 8. subjects
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."subjects" ("sub_sch_id", "sub_name", "sub_code", "sub_color")
    VALUES (v_school_id, 'Matematika', 'MTK', '#4f46e5') RETURNING "sub_id" INTO v_subject_math_id;

    INSERT INTO "edv"."subjects" ("sub_sch_id", "sub_name", "sub_code", "sub_color")
    VALUES (v_school_id, 'Fisika', 'FIS', '#0284c7') RETURNING "sub_id" INTO v_subject_physics_id;

    INSERT INTO "edv"."subjects" ("sub_sch_id", "sub_name", "sub_code", "sub_color")
    VALUES (v_school_id, 'Bahasa Indonesia', 'BIN', '#027a48') RETURNING "sub_id" INTO v_subject_indo_id;

    -- -----------------------------------------------------------------------
    -- 9. subject_classes (penempatan guru per mapel per kelas)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class1_id, v_subject_math_id, v_scu_teacher1_id) RETURNING "scl_id" INTO v_subject_class_math_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class1_id, v_subject_physics_id, v_scu_teacher2_id) RETURNING "scl_id" INTO v_subject_class_physics_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class2_id, v_subject_indo_id, v_scu_teacher2_id) RETURNING "scl_id" INTO v_subject_class_indo_id;

    -- -----------------------------------------------------------------------
    -- 10. enrollments (guru & siswa masuk ke kelas)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_teacher1_id, v_class1_id, 'teacher') RETURNING "enr_id" INTO v_enr_teacher1_class1_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_teacher2_id, v_class1_id, 'teacher') RETURNING "enr_id" INTO v_enr_teacher2_class1_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_teacher2_id, v_class2_id, 'teacher') RETURNING "enr_id" INTO v_enr_teacher2_class2_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_student1_id, v_class1_id, 'student') RETURNING "enr_id" INTO v_enr_student1_class1_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_student2_id, v_class1_id, 'student') RETURNING "enr_id" INTO v_enr_student2_class1_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_student3_id, v_class1_id, 'student') RETURNING "enr_id" INTO v_enr_student3_class1_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_student4_id, v_class2_id, 'student') RETURNING "enr_id" INTO v_enr_student4_class2_id;

    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role")
    VALUES (v_school_id, v_scu_student5_id, v_class2_id, 'student') RETURNING "enr_id" INTO v_enr_student5_class2_id;

    -- -----------------------------------------------------------------------
    -- 11. assignment_categories & assessments_weights
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."assignment_categories" ("asc_sch_id", "asc_name")
    VALUES (v_school_id, 'Tugas Harian') RETURNING "asc_id" INTO v_category_daily_id;

    INSERT INTO "edv"."assignment_categories" ("asc_sch_id", "asc_name")
    VALUES (v_school_id, 'Ujian') RETURNING "asc_id" INTO v_category_exam_id;

    INSERT INTO "edv"."assessments_weights" ("asw_sub_id", "asw_asc_id", "asw_weight") VALUES
        (v_subject_math_id, v_category_daily_id, 40.00),
        (v_subject_math_id, v_category_exam_id, 60.00);

    -- -----------------------------------------------------------------------
    -- 12. assignments
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late")
    VALUES (v_school_id, v_subject_class_math_id, v_category_daily_id,
            'Latihan Aljabar Dasar', 'Kerjakan soal latihan aljabar pada modul bab 1.', now() + interval '7 days', v_user_teacher1_id, false)
    RETURNING "asg_id" INTO v_assignment1_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late")
    VALUES (v_school_id, v_subject_class_math_id, v_category_exam_id,
            'Ujian Tengah Semester Matematika', 'UTS materi aljabar dan trigonometri dasar.', now() + interval '30 days', v_user_teacher1_id, true)
    RETURNING "asg_id" INTO v_assignment2_id;

    -- -----------------------------------------------------------------------
    -- 13. submissions & assessments
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at")
    VALUES (v_school_id, v_assignment1_id, v_user_student1_id, now() - interval '1 day')
    RETURNING "sbm_id" INTO v_submission1_id;

    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at")
    VALUES (v_school_id, v_assignment1_id, v_user_student2_id, now() - interval '2 hours')
    RETURNING "sbm_id" INTO v_submission2_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    VALUES (v_submission1_id, 88.50, 'Bagus, tingkatkan lagi ketelitian di soal nomor 3.', v_user_teacher1_id);
    -- Catatan: submission Fajar (v_submission2_id) sengaja dibiarkan belum dinilai
    -- supaya alur "pending review" di dashboard guru punya data untuk ditunjukkan.

    -- -----------------------------------------------------------------------
    -- 14. materials, medias, attachments, material_progress, student_notes
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by")
    VALUES (v_school_id, v_subject_class_math_id, 'Materi Aljabar Dasar', 'Ringkasan konsep dasar aljabar untuk kelas 10.', 'pdf', v_user_teacher1_id)
    RETURNING "mat_id" INTO v_material1_id;

    INSERT INTO "edv"."medias"
        ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'materi-aljabar-dasar.pdf', 245760, 'application/pdf',
            'schools/' || v_school_id || '/materials/materi-aljabar-dasar.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/materi-aljabar-dasar.pdf',
            true, 'material', v_material1_id)
    RETURNING "med_id" INTO v_media1_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material1_id, 'material', v_media1_id);

    INSERT INTO "edv"."material_progress" ("map_usr_id", "map_mat_id", "map_status", "last_opened_at")
    VALUES (v_user_student1_id, v_material1_id, 'completed', now() - interval '3 hours');

    INSERT INTO "edv"."student_notes" ("snt_sch_id", "snt_usr_id", "snt_mat_id", "snt_content")
    VALUES (v_school_id, v_user_student1_id, v_material1_id, 'Ingat: rumus faktorisasi selisih kuadrat a^2 - b^2 = (a+b)(a-b).');

    -- -----------------------------------------------------------------------
    -- 15. feeds & comments
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."feeds" ("fds_sch_id", "fds_cls_id", "fds_content", "created_by")
    VALUES (v_school_id, v_class1_id, 'Selamat pagi anak-anak, jangan lupa kumpulkan tugas aljabar paling lambat hari Jumat ya.', v_user_teacher1_id)
    RETURNING "fds_id" INTO v_feed1_id;

    INSERT INTO "edv"."comments" ("cmn_sch_id", "cmn_source_type", "cmn_source_id", "cmn_usr_id", "cmn_content")
    VALUES (v_school_id, 'feed', v_feed1_id, v_user_student1_id, 'Baik, Pak. Siap dikumpulkan hari ini.');

    -- -----------------------------------------------------------------------
    -- 16. chat_rooms, chat_room_members, chat_messages, chat_read_receipts
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."chat_rooms" ("room_sch_id", "room_name", "room_type", "room_ref_type", "room_ref_id", "created_by")
    VALUES (v_school_id, 'Kelas 10 IPA 1', 'class', 'class', v_class1_id, v_user_teacher1_id)
    RETURNING "room_id" INTO v_room1_id;

    INSERT INTO "edv"."chat_rooms" ("room_sch_id", "room_name", "room_type", "room_ref_type", "room_ref_id", "created_by")
    VALUES (v_school_id, NULL, 'dm', NULL, NULL, v_user_student1_id)
    RETURNING "room_id" INTO v_room2_id;

    INSERT INTO "edv"."chat_room_members" ("crm_room_id", "crm_usr_id", "crm_enr_id", "crm_role") VALUES
        (v_room1_id, v_user_teacher1_id, v_enr_teacher1_class1_id, 'admin'),
        (v_room1_id, v_user_student1_id, v_enr_student1_class1_id, 'member'),
        (v_room1_id, v_user_student2_id, v_enr_student2_class1_id, 'member'),
        (v_room2_id, v_user_teacher1_id, NULL, 'member'),
        (v_room2_id, v_user_student1_id, NULL, 'member');

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room1_id, v_user_teacher1_id, 'Selamat siang semua, ada yang mau ditanyakan soal tugas aljabar?', 'text')
    RETURNING "msg_id" INTO v_message1_id;

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room2_id, v_user_student1_id, 'Selamat siang Pak, izin bertanya soal nomor 4.', 'text');

    INSERT INTO "edv"."chat_read_receipts" ("rct_room_id", "rct_usr_id", "last_read_msg_id", "last_read_at")
    VALUES (v_room1_id, v_user_student1_id, v_message1_id, now());

    -- -----------------------------------------------------------------------
    -- 17. notifications
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."notifications" ("ntf_usr_id", "ntf_type", "ntf_title", "ntf_message", "ntf_link", "ntf_related_id", "is_read") VALUES
        (v_user_student1_id, 'assignment_created', 'Tugas baru', 'Latihan Aljabar Dasar',
         '/student/subjects/' || v_subject_class_math_id || '/assignments/' || v_assignment1_id, v_assignment1_id, false),
        (v_user_student1_id, 'assignment_graded', 'Tugas sudah dinilai', 'Nilai Anda sudah tersedia: 88.50',
         '/student/grades', v_submission1_id, false);

    -- -----------------------------------------------------------------------
    -- 18. invitations (contoh undangan yang masih pending)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."invitations"
        ("inv_school_id", "inv_email", "inv_role", "inv_token_hash", "inv_invited_by", "inv_expires_at", "inv_full_name", "inv_class_id")
    VALUES (v_school_id, 'siswa.baru@contoh.sch.id', 'student', 'seed-placeholder-token-hash-0001',
            v_user_admin_id, now() + interval '7 days', 'Calon Siswa Baru', v_class1_id);

    -- -----------------------------------------------------------------------
    -- 19. school_registration_requests (contoh, independen dari sekolah di atas)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."school_registration_requests"
        ("srr_school_name", "srr_npsn", "srr_pic_name", "srr_pic_email", "srr_pic_phone", "srr_pic_role", "srr_message", "srr_status")
    VALUES ('SMK Cendekia Bangsa', '12345678', 'Hendra Gunawan', 'hendra@cendekiabangsa.sch.id', '081234567890',
            'Kepala Sekolah', 'Ingin menggunakan Wiyata untuk seluruh kelas TKJ.', 'pending');

    -- -----------------------------------------------------------------------
    -- 20. logs
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."logs" ("log_sch_id", "log_usr_id", "log_action", "log_metadata")
    VALUES (v_school_id, v_user_teacher1_id, 'assignment.created',
            jsonb_build_object('assignmentId', v_assignment1_id, 'title', 'Latihan Aljabar Dasar'));

    -- =========================================================================
    -- PERLUASAN DATA DEMO
    -- Menambah guru, subject, siswa, materi, tugas, submission, dan interaksi
    -- lain supaya setiap halaman Wiyata (dashboard, subject, grade report,
    -- review tugas, chat, notifikasi) terlihat hidup saat demo. Tidak ada baris
    -- lama yang diubah atau dihapus.
    -- =========================================================================

    -- -----------------------------------------------------------------------
    -- 21. guru baru: Yusuf Hidayat (Bahasa Inggris)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Yusuf Hidayat', 'yusuf.hidayat@wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_teacher3_id;

    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id")
    VALUES (v_user_teacher3_id, v_school_id) RETURNING "scu_id" INTO v_scu_teacher3_id;

    INSERT INTO "edv"."user_roles" ("urol_scu_id", "urol_rol_id") VALUES (v_scu_teacher3_id, v_role_teacher_id);

    -- -----------------------------------------------------------------------
    -- 22. subject baru: Bahasa Inggris
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."subjects" ("sub_sch_id", "sub_name", "sub_code", "sub_color")
    VALUES (v_school_id, 'Bahasa Inggris', 'ENG', '#b45309') RETURNING "sub_id" INTO v_subject_eng_id;

    -- -----------------------------------------------------------------------
    -- 23. siswa baru (perluasan roster 10 IPA 1 & 10 IPA 2)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Muhammad Iqbal', 'muhammad.iqbal@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student6_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Putri Ayu Lestari', 'putri.lestari@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student7_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Bintang Saputra', 'bintang.saputra@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student8_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Aulia Rahmah', 'aulia.rahmah@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student9_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Agus Setiawan', 'agus.setiawan@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student10_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Kirana Dewi', 'kirana.dewi@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student11_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Reza Firmansyah', 'reza.firmansyah@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student12_id;

    INSERT INTO "edv"."users" ("usr_nama_lengkap", "usr_email", "usr_password", "is_active")
    VALUES ('Yolanda Safitri', 'yolanda.safitri@siswa.wiyatanusantara.sch.id', '$2a$10$S50b3T3q7C34gBYIKdpLv.ykChekCY6mCi8uK/WFV0aATsVM3zJ.S', true)
    RETURNING "usr_id" INTO v_user_student13_id;

    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student6_id, v_school_id)  RETURNING "scu_id" INTO v_scu_student6_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student7_id, v_school_id)  RETURNING "scu_id" INTO v_scu_student7_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student8_id, v_school_id)  RETURNING "scu_id" INTO v_scu_student8_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student9_id, v_school_id)  RETURNING "scu_id" INTO v_scu_student9_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student10_id, v_school_id) RETURNING "scu_id" INTO v_scu_student10_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student11_id, v_school_id) RETURNING "scu_id" INTO v_scu_student11_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student12_id, v_school_id) RETURNING "scu_id" INTO v_scu_student12_id;
    INSERT INTO "edv"."school_users" ("scu_usr_id", "scu_sch_id") VALUES (v_user_student13_id, v_school_id) RETURNING "scu_id" INTO v_scu_student13_id;

    INSERT INTO "edv"."user_roles" ("urol_scu_id", "urol_rol_id") VALUES
        (v_scu_student6_id,  v_role_student_id),
        (v_scu_student7_id,  v_role_student_id),
        (v_scu_student8_id,  v_role_student_id),
        (v_scu_student9_id,  v_role_student_id),
        (v_scu_student10_id, v_role_student_id),
        (v_scu_student11_id, v_role_student_id),
        (v_scu_student12_id, v_role_student_id),
        (v_scu_student13_id, v_role_student_id);

    -- -----------------------------------------------------------------------
    -- 24. subject_classes tambahan — tiap kelas dilengkapi seluruh mapel
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class1_id, v_subject_indo_id, v_scu_teacher2_id) RETURNING "scl_id" INTO v_subject_class_indo_c1_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class2_id, v_subject_math_id, v_scu_teacher1_id) RETURNING "scl_id" INTO v_subject_class_math_c2_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class2_id, v_subject_physics_id, v_scu_teacher2_id) RETURNING "scl_id" INTO v_subject_class_physics_c2_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class1_id, v_subject_eng_id, v_scu_teacher3_id) RETURNING "scl_id" INTO v_subject_class_eng_c1_id;

    INSERT INTO "edv"."subject_classes" ("scl_cls_id", "scl_sub_id", "scl_scu_id")
    VALUES (v_class2_id, v_subject_eng_id, v_scu_teacher3_id) RETURNING "scl_id" INTO v_subject_class_eng_c2_id;

    -- -----------------------------------------------------------------------
    -- 25. enrollments tambahan (guru & siswa baru masuk ke kelas)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."enrollments" ("enr_sch_id", "enr_scu_id", "enr_cls_id", "enr_role") VALUES
        (v_school_id, v_scu_teacher1_id,   v_class2_id, 'teacher'),
        (v_school_id, v_scu_teacher3_id,   v_class1_id, 'teacher'),
        (v_school_id, v_scu_teacher3_id,   v_class2_id, 'teacher'),
        (v_school_id, v_scu_student6_id,   v_class1_id, 'student'),
        (v_school_id, v_scu_student7_id,   v_class1_id, 'student'),
        (v_school_id, v_scu_student8_id,   v_class1_id, 'student'),
        (v_school_id, v_scu_student9_id,   v_class1_id, 'student'),
        (v_school_id, v_scu_student10_id,  v_class2_id, 'student'),
        (v_school_id, v_scu_student11_id,  v_class2_id, 'student'),
        (v_school_id, v_scu_student12_id,  v_class2_id, 'student'),
        (v_school_id, v_scu_student13_id,  v_class2_id, 'student');

    -- -----------------------------------------------------------------------
    -- 26. assessments_weights tambahan (Fisika, B. Indonesia, B. Inggris)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."assessments_weights" ("asw_sub_id", "asw_asc_id", "asw_weight") VALUES
        (v_subject_physics_id, v_category_daily_id, 50.00),
        (v_subject_physics_id, v_category_exam_id,  50.00),
        (v_subject_indo_id,    v_category_daily_id, 30.00),
        (v_subject_indo_id,    v_category_exam_id,  70.00),
        (v_subject_eng_id,     v_category_daily_id, 50.00),
        (v_subject_eng_id,     v_category_exam_id,  50.00);

    -- -----------------------------------------------------------------------
    -- 27. materials tambahan, medias, attachments (2 per subject_class, judul
    --     relevan per mapel, tanggal dibuat disebar 1 bulan terakhir, kombinasi
    --     PDF/gambar/video/ppt supaya lampiran bervariasi)
    -- -----------------------------------------------------------------------

    -- Matematika @ 10 IPA 1 (menambah dari 1 materi existing menjadi 3)
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_math_id, 'Persamaan Linear Satu Variabel',
            'Konsep dasar dan contoh soal persamaan linear satu variabel.', 'pdf', v_user_teacher1_id, now() - interval '21 days')
    RETURNING "mat_id" INTO v_material2_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'persamaan-linear-satu-variabel.pdf', 198400, 'application/pdf',
            'schools/' || v_school_id || '/materials/persamaan-linear-satu-variabel.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/persamaan-linear-satu-variabel.pdf',
            true, 'material', v_material2_id)
    RETURNING "med_id" INTO v_media2_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material2_id, 'material', v_media2_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_math_id, 'Sistem Persamaan Linear Dua Variabel (SPLDV)',
            'Metode substitusi dan eliminasi untuk menyelesaikan SPLDV.', 'ppt', v_user_teacher1_id, now() - interval '10 days')
    RETURNING "mat_id" INTO v_material3_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'spldv-slide.pptx', 812300, 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
            'schools/' || v_school_id || '/materials/spldv-slide.pptx',
            'https://storage.example.com/schools/' || v_school_id || '/materials/spldv-slide.pptx',
            true, 'material', v_material3_id)
    RETURNING "med_id" INTO v_media3_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material3_id, 'material', v_media3_id);

    -- Matematika @ 10 IPA 2
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_math_c2_id, 'Persamaan Linear Satu Variabel',
            'Konsep dasar dan contoh soal persamaan linear satu variabel.', 'pdf', v_user_teacher1_id, now() - interval '25 days')
    RETURNING "mat_id" INTO v_material4_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'persamaan-linear-satu-variabel.pdf', 198400, 'application/pdf',
            'schools/' || v_school_id || '/materials/persamaan-linear-satu-variabel-c2.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/persamaan-linear-satu-variabel-c2.pdf',
            true, 'material', v_material4_id)
    RETURNING "med_id" INTO v_media4_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material4_id, 'material', v_media4_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_math_c2_id, 'Sistem Persamaan Linear Dua Variabel (SPLDV)',
            'Metode substitusi dan eliminasi untuk menyelesaikan SPLDV.', 'other', v_user_teacher1_id, now() - interval '12 days');

    -- Fisika @ 10 IPA 1
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_physics_id, 'Gerak Lurus Beraturan (GLB)',
            'Animasi dan penjelasan konsep gerak lurus beraturan.', 'video', v_user_teacher2_id, now() - interval '30 days')
    RETURNING "mat_id" INTO v_material5_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'glb-animasi.mp4', 8432000, 'video/mp4',
            'schools/' || v_school_id || '/materials/glb-animasi.mp4',
            'https://storage.example.com/schools/' || v_school_id || '/materials/glb-animasi.mp4',
            true, 'material', v_material5_id)
    RETURNING "med_id" INTO v_media5_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material5_id, 'material', v_media5_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_physics_id, 'Hukum Newton tentang Gerak',
            'Ringkasan hukum I, II, dan III Newton beserta contoh penerapannya.', 'pdf', v_user_teacher2_id, now() - interval '14 days')
    RETURNING "mat_id" INTO v_material6_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'hukum-newton-ringkasan.pdf', 256700, 'application/pdf',
            'schools/' || v_school_id || '/materials/hukum-newton-ringkasan.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/hukum-newton-ringkasan.pdf',
            true, 'material', v_material6_id)
    RETURNING "med_id" INTO v_media6_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material6_id, 'material', v_media6_id);

    -- Fisika @ 10 IPA 2
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_physics_c2_id, 'Gerak Lurus Beraturan (GLB)',
            'Animasi dan penjelasan konsep gerak lurus beraturan.', 'video', v_user_teacher2_id, now() - interval '28 days')
    RETURNING "mat_id" INTO v_material7_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'glb-animasi.mp4', 8432000, 'video/mp4',
            'schools/' || v_school_id || '/materials/glb-animasi-c2.mp4',
            'https://storage.example.com/schools/' || v_school_id || '/materials/glb-animasi-c2.mp4',
            true, 'material', v_material7_id)
    RETURNING "med_id" INTO v_media7_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material7_id, 'material', v_media7_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_physics_c2_id, 'Hukum Newton tentang Gerak',
            'Ringkasan hukum I, II, dan III Newton beserta contoh penerapannya.', 'pdf', v_user_teacher2_id, now() - interval '7 days')
    RETURNING "mat_id" INTO v_material8_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'hukum-newton-catatan.jpg', 184200, 'image/jpeg',
            'schools/' || v_school_id || '/materials/hukum-newton-catatan.jpg',
            'https://storage.example.com/schools/' || v_school_id || '/materials/hukum-newton-catatan.jpg',
            true, 'material', v_material8_id)
    RETURNING "med_id" INTO v_media8_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material8_id, 'material', v_media8_id);

    -- Bahasa Indonesia @ 10 IPA 1
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_indo_c1_id, 'Teks Eksposisi',
            'Struktur dan ciri kebahasaan teks eksposisi.', 'pdf', v_user_teacher2_id, now() - interval '20 days')
    RETURNING "mat_id" INTO v_material9_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'teks-eksposisi.pdf', 172900, 'application/pdf',
            'schools/' || v_school_id || '/materials/teks-eksposisi-c1.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/teks-eksposisi-c1.pdf',
            true, 'material', v_material9_id)
    RETURNING "med_id" INTO v_media9_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material9_id, 'material', v_media9_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_indo_c1_id, 'Teks Argumentasi',
            'Cara menyusun argumen yang didukung fakta dan data.', 'other', v_user_teacher2_id, now() - interval '5 days')
    RETURNING "mat_id" INTO v_material10_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'infografis-teks-argumentasi.png', 96500, 'image/png',
            'schools/' || v_school_id || '/materials/infografis-teks-argumentasi.png',
            'https://storage.example.com/schools/' || v_school_id || '/materials/infografis-teks-argumentasi.png',
            true, 'material', v_material10_id)
    RETURNING "med_id" INTO v_media10_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material10_id, 'material', v_media10_id);

    -- Bahasa Indonesia @ 10 IPA 2
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_indo_id, 'Teks Eksposisi',
            'Struktur dan ciri kebahasaan teks eksposisi.', 'pdf', v_user_teacher2_id, now() - interval '18 days')
    RETURNING "mat_id" INTO v_material11_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'teks-eksposisi.pdf', 172900, 'application/pdf',
            'schools/' || v_school_id || '/materials/teks-eksposisi-c2.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/teks-eksposisi-c2.pdf',
            true, 'material', v_material11_id)
    RETURNING "med_id" INTO v_media11_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material11_id, 'material', v_media11_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_indo_id, 'Teks Argumentasi',
            'Cara menyusun argumen yang didukung fakta dan data.', 'other', v_user_teacher2_id, now() - interval '3 days');

    -- Bahasa Inggris @ 10 IPA 1
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c1_id, 'Simple Present Tense',
            'Rumus dan penggunaan simple present tense dalam kalimat sehari-hari.', 'ppt', v_user_teacher3_id, now() - interval '15 days')
    RETURNING "mat_id" INTO v_material12_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'simple-present-tense-slide.pptx', 654200, 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
            'schools/' || v_school_id || '/materials/simple-present-tense-slide.pptx',
            'https://storage.example.com/schools/' || v_school_id || '/materials/simple-present-tense-slide.pptx',
            true, 'material', v_material12_id)
    RETURNING "med_id" INTO v_media12_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material12_id, 'material', v_media12_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c1_id, 'Descriptive Text',
            'Struktur dan contoh descriptive text tentang orang dan tempat.', 'pdf', v_user_teacher3_id, now() - interval '1 day')
    RETURNING "mat_id" INTO v_material13_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'contoh-descriptive-text.jpg', 143800, 'image/jpeg',
            'schools/' || v_school_id || '/materials/contoh-descriptive-text.jpg',
            'https://storage.example.com/schools/' || v_school_id || '/materials/contoh-descriptive-text.jpg',
            true, 'material', v_material13_id)
    RETURNING "med_id" INTO v_media13_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material13_id, 'material', v_media13_id);

    -- Bahasa Inggris @ 10 IPA 2
    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c2_id, 'Simple Present Tense',
            'Rumus dan penggunaan simple present tense dalam kalimat sehari-hari.', 'ppt', v_user_teacher3_id, now() - interval '17 days')
    RETURNING "mat_id" INTO v_material14_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'simple-present-tense-handout.pdf', 133100, 'application/pdf',
            'schools/' || v_school_id || '/materials/simple-present-tense-handout.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/simple-present-tense-handout.pdf',
            true, 'material', v_material14_id)
    RETURNING "med_id" INTO v_media14_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material14_id, 'material', v_media14_id);

    INSERT INTO "edv"."materials" ("mat_sch_id", "mat_scl_id", "mat_title", "mat_desc", "mat_types", "created_by", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c2_id, 'Descriptive Text',
            'Struktur dan contoh descriptive text tentang orang dan tempat.', 'pdf', v_user_teacher3_id, now())
    RETURNING "mat_id" INTO v_material15_id;

    INSERT INTO "edv"."medias" ("med_sch_id", "med_name", "med_file_size", "med_mime_type", "med_storage_path", "med_file_url", "is_public", "med_owner_type", "med_owner_id")
    VALUES (v_school_id, 'descriptive-text-writing-guide.pdf', 205600, 'application/pdf',
            'schools/' || v_school_id || '/materials/descriptive-text-writing-guide.pdf',
            'https://storage.example.com/schools/' || v_school_id || '/materials/descriptive-text-writing-guide.pdf',
            true, 'material', v_material15_id)
    RETURNING "med_id" INTO v_media15_id;

    INSERT INTO "edv"."attachments" ("att_sch_id", "att_source_id", "att_source_type", "att_med_id")
    VALUES (v_school_id, v_material15_id, 'material', v_media15_id);

    -- -----------------------------------------------------------------------
    -- 28. material_progress tambahan
    -- Catatan: enum status_progress di database hanya punya 'not_started' dan
    -- 'completed' (tidak ada 'in_progress'), jadi variasi dibatasi ke 2 nilai
    -- tersebut mengikuti skema yang sebenarnya.
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."material_progress" ("map_usr_id", "map_mat_id", "map_status", "last_opened_at") VALUES
        (v_user_student1_id,  v_material2_id,  'completed',   now() - interval '18 days'),
        (v_user_student1_id,  v_material3_id,  'not_started', NULL),
        (v_user_student2_id,  v_material2_id,  'completed',   now() - interval '15 days'),
        (v_user_student3_id,  v_material5_id,  'completed',   now() - interval '25 days'),
        (v_user_student6_id,  v_material5_id,  'not_started', NULL),
        (v_user_student7_id,  v_material9_id,  'completed',   now() - interval '12 days'),
        (v_user_student4_id,  v_material4_id,  'completed',   now() - interval '20 days'),
        (v_user_student5_id,  v_material7_id,  'not_started', NULL),
        (v_user_student10_id, v_material11_id, 'completed',   now() - interval '10 days'),
        (v_user_student6_id,  v_material12_id, 'completed',   now() - interval '8 days'),
        (v_user_student11_id, v_material14_id, 'not_started', NULL);

    -- -----------------------------------------------------------------------
    -- 29. student_notes tambahan
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."student_notes" ("snt_sch_id", "snt_usr_id", "snt_mat_id", "snt_content") VALUES
        (v_school_id, v_user_student1_id,  v_material2_id,  'Contoh soal: 2x + 5 = 15, maka x = 5. Ingat pindahkan angka ke ruas kanan dulu.'),
        (v_school_id, v_user_student4_id,  v_material4_id,  'Eliminasi salah satu variabel dulu sebelum substitusi ke persamaan lain.'),
        (v_school_id, v_user_student3_id,  v_material5_id,  'Rumus GLB: jarak = kecepatan x waktu. Kecepatan pada GLB selalu konstan.'),
        (v_school_id, v_user_student7_id,  v_material9_id,  'Teks eksposisi harus memuat fakta, bukan opini pribadi penulis.'),
        (v_school_id, v_user_student6_id,  v_material12_id, 'Simple present dipakai untuk kebiasaan dan fakta umum. Tambahkan -s/-es untuk subjek orang ketiga tunggal.');

    -- -----------------------------------------------------------------------
    -- 30. assignments tambahan (2 per subject_class: Tugas Harian + Ujian)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_math_c2_id, v_category_daily_id, 'Latihan Persamaan Linear',
            'Kerjakan latihan soal persamaan linear satu variabel pada buku paket halaman 24.', now() - interval '3 days', v_user_teacher1_id, true, now() - interval '14 days')
    RETURNING "asg_id" INTO v_assignment3_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_math_c2_id, v_category_exam_id, 'Ulangan Harian SPLDV',
            'Ulangan harian materi sistem persamaan linear dua variabel.', now() + interval '14 days', v_user_teacher1_id, false, now() - interval '7 days')
    RETURNING "asg_id" INTO v_assignment4_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_physics_id, v_category_daily_id, 'Latihan Soal Gerak Lurus Beraturan',
            'Kerjakan 5 soal hitungan GLB pada lembar kerja siswa.', now() + interval '1 day', v_user_teacher2_id, true, now() - interval '5 days')
    RETURNING "asg_id" INTO v_assignment5_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_physics_id, v_category_exam_id, 'Ulangan Bab Gerak Lurus',
            'Ulangan bab gerak lurus meliputi GLB dan GLBB.', now() + interval '21 days', v_user_teacher2_id, true, now() - interval '3 days')
    RETURNING "asg_id" INTO v_assignment6_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_physics_c2_id, v_category_daily_id, 'Tugas Analisis Gerak Lurus',
            'Analisis grafik jarak-waktu pada kasus gerak lurus beraturan.', now() - interval '10 days', v_user_teacher2_id, false, now() - interval '20 days')
    RETURNING "asg_id" INTO v_assignment7_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_physics_c2_id, v_category_exam_id, 'Ulangan Hukum Newton',
            'Ulangan hukum I, II, dan III Newton.', now() + interval '7 days', v_user_teacher2_id, true, now() - interval '2 days')
    RETURNING "asg_id" INTO v_assignment8_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_indo_c1_id, v_category_daily_id, 'Menulis Teks Eksposisi',
            'Tulis teks eksposisi singkat bertema lingkungan hidup, minimal 3 paragraf.', now() + interval '2 days', v_user_teacher2_id, true, now() - interval '4 days')
    RETURNING "asg_id" INTO v_assignment9_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_indo_c1_id, v_category_exam_id, 'Ulangan Teks Argumentasi',
            'Ulangan struktur dan ciri kebahasaan teks argumentasi.', now() + interval '25 days', v_user_teacher2_id, false, now() - interval '1 day')
    RETURNING "asg_id" INTO v_assignment10_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_indo_id, v_category_daily_id, 'Latihan Menyusun Teks Argumentasi',
            'Susun kerangka teks argumentasi dengan tema pilihan sendiri.', now() - interval '1 day', v_user_teacher2_id, true, now() - interval '10 days')
    RETURNING "asg_id" INTO v_assignment11_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_indo_id, v_category_exam_id, 'Ulangan Teks Eksposisi',
            'Ulangan struktur dan kaidah kebahasaan teks eksposisi.', now() + interval '10 days', v_user_teacher2_id, true, now() - interval '6 days')
    RETURNING "asg_id" INTO v_assignment12_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c1_id, v_category_daily_id, 'Simple Present Tense Exercise',
            'Complete the worksheet about simple present tense, 20 sentences.', now() + interval '3 days', v_user_teacher3_id, true, now() - interval '8 days')
    RETURNING "asg_id" INTO v_assignment13_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c1_id, v_category_exam_id, 'Descriptive Text Writing Test',
            'Write a descriptive text about your favorite place, minimum 150 words.', now() + interval '18 days', v_user_teacher3_id, false, now() - interval '2 days')
    RETURNING "asg_id" INTO v_assignment14_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c2_id, v_category_daily_id, 'Grammar Practice: Simple Present',
            'Answer the grammar practice questions on simple present tense.', now() - interval '5 days', v_user_teacher3_id, true, now() - interval '15 days')
    RETURNING "asg_id" INTO v_assignment15_id;

    INSERT INTO "edv"."assignments" ("asg_sch_id", "asg_scl_id", "asg_asc_id", "asg_title", "asg_desc", "asg_deadline", "created_by", "asg_allowed_late", "created_at")
    VALUES (v_school_id, v_subject_class_eng_c2_id, v_category_exam_id, 'Writing Test: Descriptive Text',
            'Write a descriptive text describing a person you admire.', now() + interval '15 days', v_user_teacher3_id, true, now() - interval '1 day')
    RETURNING "asg_id" INTO v_assignment16_id;

    -- -----------------------------------------------------------------------
    -- 31. submissions tambahan untuk assignment lama (Latihan Aljabar Dasar,
    --     UTS Matematika) — hanya menambah baris baru dari siswa baru,
    --     baris submission lama tidak disentuh.
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment1_id, v_user_student3_id, now() - interval '6 hours'),
        (v_school_id, v_assignment1_id, v_user_student6_id, now() - interval '2 days'),
        (v_school_id, v_assignment1_id, v_user_student7_id, now() - interval '1 day'),
        (v_school_id, v_assignment2_id, v_user_student1_id, now() - interval '3 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 95.00, 'Pengerjaan rapi, sudah tepat.', v_user_teacher1_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment1_id AND sbm_usr_id = v_user_student6_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 72.00, 'Masih ada kesalahan di soal nomor 2, tolong dicek lagi.', v_user_teacher1_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment1_id AND sbm_usr_id = v_user_student7_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 92.00, 'Bagus, konsep substitusi sudah dikuasai.', v_user_teacher1_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment2_id AND sbm_usr_id = v_user_student1_id;
    -- Catatan: submission Nadia di assignment1 sengaja dibiarkan belum dinilai.

    -- -----------------------------------------------------------------------
    -- 32. submissions & assessments untuk assignment baru
    -- -----------------------------------------------------------------------

    -- #3 Latihan Persamaan Linear (Matematika 10 IPA 2) — rate tinggi, 5/6 submit
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment3_id, v_user_student4_id,  now() - interval '5 days'),
        (v_school_id, v_assignment3_id, v_user_student5_id,  now() - interval '4 days'),
        (v_school_id, v_assignment3_id, v_user_student10_id, now() - interval '1 day'),   -- telat, deadline -3 hari
        (v_school_id, v_assignment3_id, v_user_student11_id, now() - interval '4 days'),
        (v_school_id, v_assignment3_id, v_user_student13_id, now() - interval '2 days');  -- telat

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 88.00, 'Baik, lanjutkan.', v_user_teacher1_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment3_id AND sbm_usr_id = v_user_student4_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 76.00, 'Perhatikan lagi tanda operasi.', v_user_teacher1_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment3_id AND sbm_usr_id = v_user_student5_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 65.00, 'Beberapa langkah penyelesaian belum tepat.', v_user_teacher1_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment3_id AND sbm_usr_id = v_user_student11_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 60.00, 'Karena terlambat, nilai dikurangi. Perbaiki manajemen waktu.', v_user_teacher1_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment3_id AND sbm_usr_id = v_user_student13_id;
    -- Catatan: submission Agus (student10) sengaja belum dinilai (pending review). Reza (student12) belum submit.

    -- #4 Ulangan Harian SPLDV (Matematika 10 IPA 2) — rate rendah, deadline masih jauh
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment4_id, v_user_student4_id, now() - interval '1 day');
    -- Catatan: sengaja belum dinilai, 5 siswa lain belum mengumpulkan karena deadline masih 14 hari lagi.

    -- #5 Latihan Soal GLB (Fisika 10 IPA 1) — rate sedang, deadline besok
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment5_id, v_user_student1_id, now() - interval '1 day'),
        (v_school_id, v_assignment5_id, v_user_student2_id, now() - interval '12 hours'),
        (v_school_id, v_assignment5_id, v_user_student6_id, now() - interval '6 hours');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 100.00, 'Sempurna!', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment5_id AND sbm_usr_id = v_user_student1_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 84.00, 'Bagus, tinggal rapikan satuan.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment5_id AND sbm_usr_id = v_user_student6_id;
    -- Catatan: submission Fajar (student2) sengaja belum dinilai.

    -- #6 Ulangan Bab Gerak Lurus (Fisika 10 IPA 1) — belum ada yang submit, deadline masih 3 minggu lagi

    -- #7 Tugas Analisis Gerak Lurus (Fisika 10 IPA 2) — rate tinggi, sudah lama lewat deadline
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment7_id, v_user_student4_id,  now() - interval '12 days'),
        (v_school_id, v_assignment7_id, v_user_student5_id,  now() - interval '11 days'),
        (v_school_id, v_assignment7_id, v_user_student10_id, now() - interval '11 days'),
        (v_school_id, v_assignment7_id, v_user_student11_id, now() - interval '13 days'),
        (v_school_id, v_assignment7_id, v_user_student13_id, now() - interval '12 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 92.00, 'Analisis grafik sudah tepat.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment7_id AND sbm_usr_id = v_user_student4_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 80.00, 'Cukup baik, lengkapi kesimpulan.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment7_id AND sbm_usr_id = v_user_student5_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 68.00, 'Grafik kurang presisi.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment7_id AND sbm_usr_id = v_user_student11_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 55.00, 'Perlu belajar lagi cara membaca grafik jarak-waktu.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment7_id AND sbm_usr_id = v_user_student13_id;
    -- Catatan: submission Agus (student10) sengaja belum dinilai. Reza (student12) tidak pernah mengumpulkan.

    -- #8 Ulangan Hukum Newton (Fisika 10 IPA 2) — rate sedang
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment8_id, v_user_student4_id,  now() - interval '2 days'),
        (v_school_id, v_assignment8_id, v_user_student5_id,  now() - interval '1 day'),
        (v_school_id, v_assignment8_id, v_user_student11_id, now() - interval '1 day');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 76.00, 'Cukup baik.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment8_id AND sbm_usr_id = v_user_student4_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 88.00, 'Pemahaman hukum Newton sudah baik.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment8_id AND sbm_usr_id = v_user_student5_id;
    -- Catatan: submission Kirana (student11) sengaja belum dinilai.

    -- #9 Menulis Teks Eksposisi (B. Indonesia 10 IPA 1) — rate sedang
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment9_id, v_user_student1_id, now() - interval '1 day'),
        (v_school_id, v_assignment9_id, v_user_student3_id, now() - interval '8 hours'),
        (v_school_id, v_assignment9_id, v_user_student7_id, now() - interval '1 day'),
        (v_school_id, v_assignment9_id, v_user_student6_id, now() - interval '3 hours');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 95.00, 'Struktur eksposisi sudah lengkap dan jelas.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment9_id AND sbm_usr_id = v_user_student1_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 84.00, 'Bagus, tambahkan lebih banyak data pendukung.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment9_id AND sbm_usr_id = v_user_student7_id;
    -- Catatan: submission Nadia (student3) dan Iqbal (student6) sengaja belum dinilai.

    -- #10 Ulangan Teks Argumentasi (B. Indonesia 10 IPA 1) — belum ada yang submit

    -- #11 Latihan Menyusun Teks Argumentasi (B. Indonesia 10 IPA 2) — rate tinggi, baru lewat kemarin
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment11_id, v_user_student4_id,  now() - interval '3 days'),
        (v_school_id, v_assignment11_id, v_user_student5_id,  now() - interval '2 days'),
        (v_school_id, v_assignment11_id, v_user_student10_id, now() - interval '6 hours'),  -- telat
        (v_school_id, v_assignment11_id, v_user_student11_id, now() - interval '2 days'),
        (v_school_id, v_assignment11_id, v_user_student12_id, now() - interval '12 hours'); -- telat

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 98.00, 'Kerangka argumen sangat runtut.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment11_id AND sbm_usr_id = v_user_student4_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 72.00, 'Argumen cukup baik, perbaiki penutup.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment11_id AND sbm_usr_id = v_user_student5_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 65.00, 'Perlu data pendukung yang lebih kuat.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment11_id AND sbm_usr_id = v_user_student11_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 55.00, 'Karena terlambat dan argumen belum kuat, nilai masih rendah.', v_user_teacher2_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment11_id AND sbm_usr_id = v_user_student12_id;
    -- Catatan: submission Agus (student10) sengaja belum dinilai. Yolanda (student13) belum submit.

    -- #12 Ulangan Teks Eksposisi (B. Indonesia 10 IPA 2) — rate rendah
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment12_id, v_user_student5_id, now() - interval '1 day');
    -- Catatan: sengaja belum dinilai.

    -- #13 Simple Present Tense Exercise (B. Inggris 10 IPA 1) — rate sedang
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment13_id, v_user_student1_id, now() - interval '2 days'),
        (v_school_id, v_assignment13_id, v_user_student2_id, now() - interval '1 day'),
        (v_school_id, v_assignment13_id, v_user_student8_id, now() - interval '12 hours');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 100.00, 'Excellent work!', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment13_id AND sbm_usr_id = v_user_student1_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 68.00, 'Review the third-person singular rule again.', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment13_id AND sbm_usr_id = v_user_student2_id;
    -- Catatan: submission Bintang (student8) sengaja belum dinilai.

    -- #14 Descriptive Text Writing Test (B. Inggris 10 IPA 1) — belum ada yang submit

    -- #15 Grammar Practice: Simple Present (B. Inggris 10 IPA 2) — rate tinggi, sudah lewat deadline
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment15_id, v_user_student4_id,  now() - interval '7 days'),
        (v_school_id, v_assignment15_id, v_user_student5_id,  now() - interval '6 days'),
        (v_school_id, v_assignment15_id, v_user_student10_id, now() - interval '4 days'), -- telat
        (v_school_id, v_assignment15_id, v_user_student11_id, now() - interval '6 days'),
        (v_school_id, v_assignment15_id, v_user_student13_id, now() - interval '3 days'); -- telat

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 88.00, 'Well done.', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment15_id AND sbm_usr_id = v_user_student4_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 72.00, 'Check your spelling on number 4 and 7.', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment15_id AND sbm_usr_id = v_user_student5_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 60.00, 'Please review the material again.', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment15_id AND sbm_usr_id = v_user_student11_id;
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 55.00, 'Late submission, please manage your time better.', v_user_teacher3_id FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment15_id AND sbm_usr_id = v_user_student13_id;
    -- Catatan: submission Agus (student10) sengaja belum dinilai. Reza (student12) belum submit.

    -- #16 Writing Test: Descriptive Text (B. Inggris 10 IPA 2) — rate rendah
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment16_id, v_user_student11_id, now() - interval '1 day');
    -- Catatan: sengaja belum dinilai.

    -- -----------------------------------------------------------------------
    -- 37. Perbaikan Grade Report — tambah submission & nilai kategori "Ujian"
    --     di 6 subject_class yang tadinya 0 nilai Ujian (Matematika-10IPA2,
    --     Fisika-10IPA1, B.Indonesia-10IPA1&2, B.Inggris-10IPA1&2), supaya
    --     nilai akhir tidak mentok di persentase bobot Tugas Harian saja untuk
    --     semua siswa. Bobot & rumus tidak diubah, hanya data.
    -- -----------------------------------------------------------------------

    -- Matematika 10 IPA 2 — Ulangan Harian SPLDV (assignment4): Rizky sudah
    -- submit tapi belum dinilai; tambah 1 penilaian lagi (Salsabila).
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 84.00, 'Konsep eliminasi sudah tepat.', v_user_teacher1_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment4_id AND sbm_usr_id = v_user_student4_id;

    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment4_id, v_user_student5_id, now() - interval '2 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 72.00, 'Cukup baik, perhatikan tanda saat eliminasi.', v_user_teacher1_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment4_id AND sbm_usr_id = v_user_student5_id;

    -- Fisika 10 IPA 1 — Ulangan Bab Gerak Lurus (assignment6): tadinya 0 submission
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment6_id, v_user_student1_id, now() - interval '3 days'),
        (v_school_id, v_assignment6_id, v_user_student6_id, now() - interval '2 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 95.00, 'Pemahaman GLB dan GLBB sudah sangat baik.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment6_id AND sbm_usr_id = v_user_student1_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 68.00, 'Konsep GLBB masih perlu diperdalam.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment6_id AND sbm_usr_id = v_user_student6_id;

    -- B. Indonesia 10 IPA 1 — Ulangan Teks Argumentasi (assignment10): tadinya 0 submission
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment10_id, v_user_student1_id, now() - interval '4 days'),
        (v_school_id, v_assignment10_id, v_user_student7_id, now() - interval '3 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 88.00, 'Struktur argumen jelas dan didukung data yang cukup.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment10_id AND sbm_usr_id = v_user_student1_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 76.00, 'Argumen cukup baik, tambahkan data pendukung.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment10_id AND sbm_usr_id = v_user_student7_id;

    -- B. Indonesia 10 IPA 2 — Ulangan Teks Eksposisi (assignment12): Salsabila
    -- sudah submit tapi belum dinilai; tambah 1 penilaian lagi (Rizky).
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 65.00, 'Struktur eksposisi masih perlu diperjelas.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment12_id AND sbm_usr_id = v_user_student5_id;

    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment12_id, v_user_student4_id, now() - interval '1 day');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 92.00, 'Sangat baik, data dan fakta pendukung lengkap.', v_user_teacher2_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment12_id AND sbm_usr_id = v_user_student4_id;

    -- B. Inggris 10 IPA 1 — Descriptive Text Writing Test (assignment14): tadinya 0 submission
    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment14_id, v_user_student1_id, now() - interval '2 days'),
        (v_school_id, v_assignment14_id, v_user_student2_id, now() - interval '1 day');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 100.00, 'Very well written, great use of adjectives.', v_user_teacher3_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment14_id AND sbm_usr_id = v_user_student1_id;

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 60.00, 'Please add more descriptive details.', v_user_teacher3_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment14_id AND sbm_usr_id = v_user_student2_id;

    -- B. Inggris 10 IPA 2 — Writing Test: Descriptive Text (assignment16): Kirana
    -- sudah submit tapi belum dinilai; tambah 1 penilaian lagi (Rizky).
    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 80.00, 'Good description, minor grammar issues.', v_user_teacher3_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment16_id AND sbm_usr_id = v_user_student11_id;

    INSERT INTO "edv"."submissions" ("sbm_sch_id", "sbm_asg_id", "sbm_usr_id", "submitted_at") VALUES
        (v_school_id, v_assignment16_id, v_user_student4_id, now() - interval '2 days');

    INSERT INTO "edv"."assessments" ("asm_sbm_id", "asm_score", "asm_feedback", "assessed_by")
    SELECT sbm_id, 88.00, 'Well organized and descriptive.', v_user_teacher3_id
    FROM "edv"."submissions" WHERE sbm_asg_id = v_assignment16_id AND sbm_usr_id = v_user_student4_id;

    -- -----------------------------------------------------------------------
    -- 33. feeds & comments tambahan
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."feeds" ("fds_sch_id", "fds_cls_id", "fds_content", "created_by", "created_at")
    VALUES (v_school_id, v_class1_id, 'Pengumuman: try out matematika tingkat sekolah akan dilaksanakan minggu depan. Persiapkan diri kalian ya!', v_user_teacher1_id, now() - interval '4 days')
    RETURNING "fds_id" INTO v_feed2_id;

    INSERT INTO "edv"."feeds" ("fds_sch_id", "fds_cls_id", "fds_content", "created_by", "created_at")
    VALUES (v_school_id, v_class2_id, 'Reminder: tugas Grammar Practice sudah bisa dikumpulkan di kelas B. Inggris. Jangan lupa dikumpulkan tepat waktu.', v_user_teacher3_id, now() - interval '2 days')
    RETURNING "fds_id" INTO v_feed3_id;

    INSERT INTO "edv"."feeds" ("fds_sch_id", "fds_cls_id", "fds_content", "created_by", "created_at")
    VALUES (v_school_id, v_class1_id, 'Informasi: jadwal ulangan Fisika bab Gerak Lurus diundur menjadi dua minggu lagi karena ada acara sekolah.', v_user_teacher2_id, now() - interval '1 day')
    RETURNING "fds_id" INTO v_feed4_id;

    INSERT INTO "edv"."comments" ("cmn_sch_id", "cmn_source_type", "cmn_source_id", "cmn_usr_id", "cmn_content", "created_at") VALUES
        (v_school_id, 'feed', v_feed2_id, v_user_student6_id, 'Siap Pak, akan belajar dari sekarang.', now() - interval '4 days' + interval '3 hours'),
        (v_school_id, 'feed', v_feed3_id, v_user_student10_id, 'Baik Pak, siap dikumpulkan.', now() - interval '2 days' + interval '1 hour'),
        (v_school_id, 'feed', v_feed4_id, v_user_student1_id, 'Terima kasih infonya, Bu.', now() - interval '1 day' + interval '2 hours');

    -- -----------------------------------------------------------------------
    -- 38. komentar tambahan di Materi & Tugas (fitur ini sebelumnya 100% kosong
    --     karena seed lama cuma mengisi komentar di feed)
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."comments" ("cmn_sch_id", "cmn_source_type", "cmn_source_id", "cmn_usr_id", "cmn_content", "created_at") VALUES
        (v_school_id, 'material',   v_material2_id,  v_user_student1_id,  'Pak, untuk soal nomor 3 apakah metodenya sama dengan contoh di halaman 2?', now() - interval '18 days'),
        (v_school_id, 'material',   v_material2_id,  v_user_teacher1_id,  'Betul, gunakan metode yang sama. Coba kerjakan dulu, nanti kita bahas bersama di kelas.', now() - interval '17 days'),
        (v_school_id, 'material',   v_material9_id,  v_user_student7_id,  'Bu, apakah teks eksposisi boleh dibuka dengan kalimat tanya?', now() - interval '14 days'),
        (v_school_id, 'material',   v_material5_id,  v_user_student3_id,  'Bu, rumus GLB ini berlaku juga untuk gerak melingkar beraturan, bukan?', now() - interval '25 days'),
        (v_school_id, 'material',   v_material5_id,  v_user_teacher2_id,  'Tidak, gerak melingkar beraturan rumusnya berbeda. Kita bahas minggu depan ya.', now() - interval '24 days'),
        (v_school_id, 'assignment', v_assignment1_id, v_user_student6_id, 'Pak, hasil akhir soal nomor 2 harus dalam bentuk pecahan atau desimal?', now() - interval '1 day'),
        (v_school_id, 'assignment', v_assignment1_id, v_user_teacher1_id, 'Boleh keduanya, tapi usahakan tulis dalam bentuk pecahan paling sederhana.', now() - interval '20 hours'),
        (v_school_id, 'assignment', v_assignment9_id, v_user_student1_id, 'Bu, minimal 3 paragraf itu termasuk paragraf penutup atau tidak?', now() - interval '3 days');

    -- -----------------------------------------------------------------------
    -- 34. chat tambahan — room kelas 10 IPA 2
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."chat_rooms" ("room_sch_id", "room_name", "room_type", "room_ref_type", "room_ref_id", "created_by")
    VALUES (v_school_id, 'Kelas 10 IPA 2', 'class', 'class', v_class2_id, v_user_teacher1_id)
    RETURNING "room_id" INTO v_room3_id;

    INSERT INTO "edv"."chat_room_members" ("crm_room_id", "crm_usr_id", "crm_enr_id", "crm_role") VALUES
        (v_room3_id, v_user_teacher1_id,   (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_teacher1_id AND enr_cls_id = v_class2_id), 'admin'),
        (v_room3_id, v_user_teacher3_id,   (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_teacher3_id AND enr_cls_id = v_class2_id), 'member'),
        (v_room3_id, v_user_student4_id,   (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_student4_id AND enr_cls_id = v_class2_id), 'member'),
        (v_room3_id, v_user_student5_id,   (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_student5_id AND enr_cls_id = v_class2_id), 'member'),
        (v_room3_id, v_user_student10_id,  (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_student10_id AND enr_cls_id = v_class2_id), 'member');

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room3_id, v_user_teacher1_id, 'Selamat siang 10 IPA 2, jangan lupa kumpulkan Latihan Persamaan Linear ya.', 'text');

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room3_id, v_user_student4_id, 'Baik Pak, sudah saya kumpulkan tadi.', 'text');

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room3_id, v_user_teacher3_id, 'For English class, please submit your Grammar Practice before Friday.', 'text')
    RETURNING "msg_id" INTO v_message2_id;

    INSERT INTO "edv"."chat_read_receipts" ("rct_room_id", "rct_usr_id", "last_read_msg_id", "last_read_at")
    VALUES (v_room3_id, v_user_student4_id, v_message2_id, now());

    -- -----------------------------------------------------------------------
    -- 39. perbaikan coverage chat — Rina (guru dengan subject_class terbanyak)
    --     tadinya tidak jadi member room manapun. Tambahkan dia + beberapa
    --     siswa 10 IPA 1 lagi ke room1, supaya guru utama & mayoritas siswa
    --     tidak melihat halaman Chat kosong. Tidak semua user perlu aktif.
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."chat_room_members" ("crm_room_id", "crm_usr_id", "crm_enr_id", "crm_role") VALUES
        (v_room1_id, v_user_teacher2_id, v_enr_teacher2_class1_id, 'member'),
        (v_room1_id, v_user_student3_id, v_enr_student3_class1_id, 'member'),
        (v_room1_id, v_user_student6_id, (SELECT enr_id FROM "edv"."enrollments" WHERE enr_scu_id = v_scu_student6_id AND enr_cls_id = v_class1_id), 'member'),
        (v_room3_id, v_user_teacher2_id, v_enr_teacher2_class2_id, 'member');

    INSERT INTO "edv"."chat_messages" ("msg_room_id", "msg_usr_id", "msg_content", "msg_type")
    VALUES (v_room1_id, v_user_teacher2_id, 'Halo semua, saya izin bergabung di grup ini juga untuk update kelas Fisika dan B. Indonesia ya.', 'text');

    -- -----------------------------------------------------------------------
    -- 35. notifications tambahan — hanya pakai 5 type yang didukung backend
    --     (assignment_created, assignment_graded, material_added, feed_posted,
    --     comment_added) dengan format link yang sama dengan kode aplikasi.
    -- -----------------------------------------------------------------------
    INSERT INTO "edv"."notifications" ("ntf_usr_id", "ntf_type", "ntf_title", "ntf_message", "ntf_link", "ntf_related_id", "is_read", "created_at") VALUES
        (v_user_student4_id,  'assignment_created', 'Tugas baru', 'Latihan Persamaan Linear',
         '/student/subjects/' || v_subject_class_math_c2_id || '/assignments/' || v_assignment3_id, v_assignment3_id, true, now() - interval '14 days'),
        (v_user_student1_id,  'assignment_created', 'Tugas baru', 'Latihan Soal Gerak Lurus Beraturan',
         '/student/subjects/' || v_subject_class_physics_id || '/assignments/' || v_assignment5_id, v_assignment5_id, true, now() - interval '5 days'),
        (v_user_student1_id,  'assignment_created', 'Tugas baru', 'Menulis Teks Eksposisi',
         '/student/subjects/' || v_subject_class_indo_c1_id || '/assignments/' || v_assignment9_id, v_assignment9_id, false, now() - interval '4 days'),
        (v_user_student6_id,  'assignment_created', 'Tugas baru', 'Simple Present Tense Exercise',
         '/student/subjects/' || v_subject_class_eng_c1_id || '/assignments/' || v_assignment13_id, v_assignment13_id, false, now() - interval '8 days'),
        (v_user_student4_id,  'assignment_graded', 'Tugas sudah dinilai', 'Nilai Anda sudah tersedia: 88.00',
         '/student/grades', v_assignment7_id, true, now() - interval '11 days'),
        (v_user_student1_id,  'assignment_graded', 'Tugas sudah dinilai', 'Nilai Anda sudah tersedia: 100.00',
         '/student/grades', v_assignment5_id, false, now() - interval '1 day'),
        (v_user_student7_id,  'assignment_graded', 'Tugas sudah dinilai', 'Nilai Anda sudah tersedia: 84.00',
         '/student/grades', v_assignment9_id, false, now() - interval '1 day'),
        (v_user_student1_id,  'material_added', 'Materi baru', 'Persamaan Linear Satu Variabel',
         '/student/subjects/' || v_subject_class_math_id || '/materials/' || v_material2_id, v_material2_id, true, now() - interval '21 days'),
        (v_user_student3_id,  'material_added', 'Materi baru', 'Gerak Lurus Beraturan (GLB)',
         '/student/subjects/' || v_subject_class_physics_id || '/materials/' || v_material5_id, v_material5_id, false, now() - interval '30 days'),
        (v_user_student6_id,  'material_added', 'Materi baru', 'Descriptive Text',
         '/student/subjects/' || v_subject_class_eng_c1_id || '/materials/' || v_material13_id, v_material13_id, false, now() - interval '1 day'),
        (v_user_student1_id,  'feed_posted', 'Pengumuman kelas baru', 'Pengumuman: try out matematika tingkat sekolah minggu depan.',
         '/student/feed', v_feed2_id, true, now() - interval '4 days'),
        (v_user_student2_id,  'feed_posted', 'Pengumuman kelas baru', 'Informasi: jadwal ulangan Fisika diundur dua minggu.',
         '/student/feed', v_feed4_id, false, now() - interval '1 day'),
        (v_user_student10_id, 'feed_posted', 'Pengumuman kelas baru', 'Reminder: tugas Grammar Practice sudah bisa dikumpulkan.',
         '/student/feed', v_feed3_id, false, now() - interval '2 days'),
        (v_user_teacher1_id,  'comment_added', 'Komentar baru', 'Ada komentar baru di diskusi.',
         '/teacher/feed', v_feed2_id, false, now() - interval '4 days' + interval '3 hours'),
        (v_user_teacher3_id,  'comment_added', 'Komentar baru', 'Ada komentar baru di diskusi.',
         '/teacher/feed', v_feed3_id, true, now() - interval '2 days' + interval '1 hour');

    -- -----------------------------------------------------------------------
    -- 36. logs tambahan (seperlunya saja — tabel ini tidak dipakai handler
    --     manapun saat ini, jadi tidak difokuskan)
    -- -----------------------------------------------------------------------
    -- Catatan: created_at diisi eksplisit (sebelumnya default now() di transaksi
    -- yang sama sehingga semua log numpuk di satu titik waktu yang identik) —
    -- supaya widget "Recent Activities" di Admin Dashboard menampilkan urutan
    -- kronologis yang wajar, bukan sekumpulan baris dengan timestamp sama persis.
    INSERT INTO "edv"."logs" ("log_sch_id", "log_usr_id", "log_action", "log_metadata", "created_at") VALUES
        (v_school_id, v_user_teacher3_id, 'material.created', jsonb_build_object('materialId', v_material12_id, 'title', 'Simple Present Tense'), now() - interval '15 days'),
        (v_school_id, v_user_teacher1_id, 'assignment.created', jsonb_build_object('assignmentId', v_assignment3_id, 'title', 'Latihan Persamaan Linear'), now() - interval '14 days'),
        (v_school_id, v_user_teacher2_id, 'assessment.created', jsonb_build_object('assignmentId', v_assignment7_id, 'studentId', v_user_student4_id), now() - interval '11 days'),
        (v_school_id, v_user_admin_id,    'user.created', jsonb_build_object('note', 'Penambahan roster siswa & guru baru untuk data demo'), now() - interval '10 days'),
        (v_school_id, v_user_teacher3_id, 'assignment.created', jsonb_build_object('assignmentId', v_assignment15_id, 'title', 'Grammar Practice: Simple Present'), now() - interval '8 days'),
        (v_school_id, v_user_teacher1_id, 'feed.created', jsonb_build_object('feedId', v_feed2_id), now() - interval '4 days'),
        (v_school_id, v_user_teacher2_id, 'material.created', jsonb_build_object('materialId', v_material9_id, 'title', 'Teks Eksposisi'), now() - interval '20 days'),
        (v_school_id, v_user_teacher2_id, 'assignment.created', jsonb_build_object('assignmentId', v_assignment6_id, 'title', 'Ulangan Bab Gerak Lurus'), now() - interval '3 days'),
        (v_school_id, v_user_teacher1_id, 'assessment.created', jsonb_build_object('assignmentId', v_assignment1_id, 'studentId', v_user_student6_id), now() - interval '2 days'),
        (v_school_id, v_user_teacher3_id, 'feed.created', jsonb_build_object('feedId', v_feed3_id), now() - interval '2 days'),
        (v_school_id, v_user_admin_id,    'school_user.created', jsonb_build_object('note', 'Penambahan guru Bahasa Inggris'), now() - interval '18 days');

END $$;
