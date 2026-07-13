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

END $$;
