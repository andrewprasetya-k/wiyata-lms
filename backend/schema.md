// Use modern DBML syntax
// Project: Edv (Learning Management System)

Enum material_type {
video
pdf
ppt
other
}

Enum source_type {
material
assignment
feed
submission
comment
}

Enum owner_type {
user
material
assignment
feed
submission
comment
school
system
}

Enum class_role {
teacher
student
}

Enum status_progress {
not_started
completed
}

Table schools {
sch_id uuid [pk, default: `gen_random_uuid()`]
sch_name varchar(150)
sch_code varchar(50) [unique]
sch_address text
sch_email text
sch_phone text
sch_website text
sch_logo uuid
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]
deleted_at timestamptz
}

Table school_registration_requests {
srr_id uuid [pk, default: `gen_random_uuid()`]
srr_school_name text [not null]
srr_npsn text
srr_pic_name text [not null]
srr_pic_email text [not null]
srr_pic_phone text
srr_pic_role text
srr_message text
srr_status text [not null, default: `'pending'`]
srr_reviewed_by uuid [ref: > users.usr_id]
srr_reviewed_at timestamptz
srr_review_note text
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]

indexes {
(srr_status, created_at) [name: 'idx_school_registration_requests_status']
(lower(srr_pic_email)) [unique, name: 'idx_school_registration_requests_pending_email', note: 'partial index — only enforced WHERE srr_status = \'pending\'; prevents duplicate pending requests from the same email']
(lower(srr_school_name)) [unique, name: 'idx_school_registration_requests_pending_school_name', note: 'partial index — only enforced WHERE srr_status = \'pending\'; prevents duplicate pending requests for the same school name']
}
}

Table invitations {
inv_id uuid [pk, default: `gen_random_uuid()`]
inv_school_id uuid [not null, ref: > schools.sch_id]
inv_email text [not null]
inv_role text [not null]
inv_full_name text
inv_class_id uuid [ref: > classes.cls_id]
inv_token_hash text [not null, unique]
inv_invited_by uuid [not null, ref: > users.usr_id]
inv_target_user_id uuid [ref: > users.usr_id]
inv_expires_at timestamptz [not null]
inv_accepted_at timestamptz
inv_revoked_at timestamptz
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]

indexes {
(inv_school_id, inv_email, inv_role) [name: 'idx_invitations_school_email_role']
(inv_email, inv_accepted_at, inv_revoked_at) [name: 'idx_invitations_email_status']
(inv_expires_at) [name: 'idx_invitations_expires_at']
}
}

Table academic_years {
acy_id uuid [pk, default: `gen_random_uuid()`]
acy_sch_id uuid [ref: > schools.sch_id]
acy_name varchar(20)
is_active boolean [default: false]
created_at timestamptz [default: `now()`]

indexes {
(acy_sch_id, acy_name) [unique]
}
}

Table terms {
trm_id uuid [pk, default: `gen_random_uuid()`]
trm_acy_id uuid [ref: > academic_years.acy_id]
trm_name varchar(50)
is_active boolean [default: false]
created_at timestamptz [default: `now()`]

indexes {
(trm_acy_id, trm_name) [unique]
}
}

Table users {
usr_id uuid [pk, default: `gen_random_uuid()`]
usr_nama_lengkap varchar(150)
usr_email varchar(150) [not null]
usr_password varchar(255)
is_active boolean [default: true]
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(usr_email, deleted_at) [unique]
}
}

Table school_users {
scu_id uuid [pk, default: `gen_random_uuid()`]
scu_usr_id uuid [ref: > users.usr_id]
scu_sch_id uuid [ref: > schools.sch_id]
created_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(scu_usr_id, scu_sch_id) [unique]
}
}

Table roles {
rol_id uuid [pk, default: `gen_random_uuid()`]
rol_name varchar(50)
created_at timestamptz [default: `now()`]

// NOTE: rol_name has no unique constraint at the database level today.
// Application code must not assume role names are unique.
}

Table user_roles {
urol_id uuid [pk, default: `gen_random_uuid()`]
urol_scu_id uuid [ref: > school_users.scu_id]
urol_rol_id uuid [ref: > roles.rol_id]
created_at timestamptz [default: `now()`]

indexes {
(urol_scu_id, urol_rol_id) [unique]
(urol_scu_id) [name: 'idx_user_roles_scu']
}
}

Table medias {
med_id uuid [pk, default: `gen_random_uuid()`]
med_sch_id uuid [ref: > schools.sch_id]
med_name varchar(255)
med_file_size bigint
med_mime_type varchar(100)
med_storage_path text
med_file_url text
med_thumbnail_url text
is_public boolean [default: true]
med_owner_type owner_type
med_owner_id uuid
created_at timestamptz [default: `now()`]
deleted_at timestamptz
}

// Relationship for school logo
Ref: schools.sch_logo > medias.med_id

Table attachments {
att_id uuid [pk, default: `gen_random_uuid()`]
att_sch_id uuid [ref: > schools.sch_id]
att_source_id uuid
att_source_type source_type
att_med_id uuid [ref: > medias.med_id]
created_at timestamptz [default: `now()`]

indexes {
(att_source_type, att_source_id)
}
}

Table subjects {
sub_id uuid [pk, default: `gen_random_uuid()`]
sub_sch_id uuid [ref: > schools.sch_id]
sub_name varchar(100)
sub_code varchar(20)
sub_color varchar [not null, note: 'unconstrained varchar at the type level; CHECK (length(sub_color) < 10) enforces the effective 9-character cap instead of a typed varchar(9)']
created_at timestamptz [default: `now()`]
updated_at timestamptz

indexes {
(sub_sch_id, sub_code) [unique]
}
}

Table classes {
cls_id uuid [pk, default: `gen_random_uuid()`]
cls_sch_id uuid [ref: > schools.sch_id]
cls_trm_id uuid [ref: > terms.trm_id]
cls_code text
cls_title text
cls_desc text
created_by uuid [ref: > users.usr_id]
is_active boolean [default: true]
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(cls_sch_id, cls_code, cls_trm_id) [unique]
(cls_sch_id, deleted_at) [name: 'idx_classes_active']
}
}

//subject_classes untuk kelas per mata pelajaran
Table subject_classes {
scl_id uuid [pk, default: `gen_random_uuid()`]
scl_cls_id uuid [ref: > classes.cls_id]
scl_sub_id uuid [ref: > subjects.sub_id]
scl_scu_id uuid [ref: > school_users.scu_id]

indexes {
(scl_cls_id, scl_sub_id, scl_scu_id) [unique]
}
}

//enrollments untuk kelas (misal 12 IPA)
Table enrollments {
enr_id uuid [pk, default: `gen_random_uuid()`]
enr_sch_id uuid [ref: > schools.sch_id]
enr_scu_id uuid [ref: > school_users.scu_id]
enr_cls_id uuid [ref: > classes.cls_id]
enr_role class_role
joined_at timestamptz [default: `now()`]
left_at timestamptz

indexes {
(enr_scu_id, enr_cls_id) [unique]
(enr_cls_id) [name: 'idx_enrollments_class']
(enr_cls_id, enr_role) [name: 'idx_enrollments_class_role']
(enr_scu_id) [name: 'idx_enrollments_user']
}
}

Table materials {
mat_id uuid [pk, default: `gen_random_uuid()`]
mat_sch_id uuid [ref: > schools.sch_id]

// Ubah dari cls_id menjadi scl_id
mat_scl_id uuid [ref: > subject_classes.scl_id]

mat_title varchar(150)
mat_desc text
mat_types material_type
created_by uuid [ref: > users.usr_id]
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(mat_sch_id, deleted_at) [name: 'idx_materials_active']
(mat_scl_id) [name: 'idx_materials_class']
}
}

Table material_progress {
map_id uuid [pk, default: `gen_random_uuid()`]
map_usr_id uuid [ref: > users.usr_id]
map_mat_id uuid [ref: > materials.mat_id]
map_status status_progress
last_opened_at timestamptz

indexes {
(map_usr_id, map_mat_id) [unique]
(map_mat_id) [name: 'idx_material_progress_material']
(map_usr_id) [name: 'idx_material_progress_user']
}
}

Table feeds {
fds_id uuid [pk, default: `gen_random_uuid()`]
fds_sch_id uuid [ref: > schools.sch_id]
fds_cls_id uuid [ref: > classes.cls_id]
fds_content text
created_by uuid [ref: > users.usr_id]
created_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(fds_sch_id, deleted_at) [name: 'idx_feeds_active']
}
}

Table comments {
cmn_id uuid [pk, default: `gen_random_uuid()`]
cmn_sch_id uuid [ref: > schools.sch_id]
cmn_source_type source_type
cmn_source_id uuid
cmn_usr_id uuid [ref: > users.usr_id]
cmn_content text
created_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(cmn_source_type, cmn_source_id)
}
}

Table assignment_categories {
asc_id uuid [pk, default: `gen_random_uuid()`]
asc_sch_id uuid [ref: > schools.sch_id]
asc_name varchar(50)
created_at timestamptz [default: `now()`]

indexes {
(asc_sch_id, asc_name) [unique]
}
}

Table assignments {
asg_id uuid [pk, default: `gen_random_uuid()`]
asg_sch_id uuid [ref: > schools.sch_id]
asg_scl_id uuid [ref: > subject_classes.scl_id]
asg_asc_id uuid [ref: > assignment_categories.asc_id]
asg_title varchar(150)
asg_desc text
asg_deadline timestamptz
asg_allowed_late bool [default: true]
created_by uuid [ref: > users.usr_id]
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(asg_sch_id, deleted_at) [name: 'idx_assignments_active']
(asg_scl_id) [name: 'idx_assignments_class']
}
}

Table submissions {
sbm_id uuid [pk, default: `gen_random_uuid()`]
sbm_sch_id uuid [ref: > schools.sch_id]
sbm_asg_id uuid [ref: > assignments.asg_id]
sbm_usr_id uuid [ref: > users.usr_id]
submitted_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(sbm_asg_id, sbm_usr_id) [unique]
(sbm_asg_id, sbm_usr_id) [name: 'idx_submissions_asg_user', note: 'redundant plain btree alongside the unique constraint above — both exist in the live database']
(sbm_asg_id) [name: 'idx_submissions_assignment']
(sbm_usr_id) [name: 'idx_submissions_user']
}
}

Table assessments {
asm_id uuid [pk, default: `gen_random_uuid()`]
asm_sbm_id uuid [ref: > submissions.sbm_id]
asm_score decimal(5,2)
asm_feedback text
assessed_by uuid [ref: > users.usr_id]
assessed_at timestamptz [default: `now()`]

indexes {
(asm_sbm_id) [name: 'idx_assessments_submission', note: 'plain btree index, NOT unique at the database level — one submission is not guaranteed to have only one assessment row. UpsertAssessment() in assignment_repo.go compensates for this in application code by keeping only the newest row per submission and deleting duplicates on every write.']
}
}

Table assessments_weights {
asw_id uuid [pk, default: `gen_random_uuid()`]
asw_sub_id uuid [ref: > subjects.sub_id]
asw_asc_id uuid [ref: > assignment_categories.asc_id]
asw_weight decimal(5,2)

indexes {
(asw_sub_id, asw_asc_id) [unique]
}
}

Table logs {
log_id uuid [pk, default: `gen_random_uuid()`]
log_sch_id uuid [ref: > schools.sch_id]
log_usr_id uuid [ref: > users.usr_id]
log_action varchar(150)
log_metadata jsonb
created_at timestamptz [default: `now()`]
}

Table notifications {
ntf_id uuid [pk, default: `gen_random_uuid()`]
ntf_usr_id uuid [ref: > users.usr_id]
ntf_type varchar(50)
ntf_title varchar(255)
ntf_message text
ntf_link text
ntf_related_id uuid
is_read boolean [default: false]
created_at timestamptz [default: `now()`]

indexes {
(ntf_usr_id, is_read, created_at) [name: 'idx_notifications_user']
}
}

# chat app schema

Enum chat_room_type {
class  
 subject  
 dm  
 group // NEW — grup bebas antar user dalam satu sekolah
}

Enum chat_message_type {
text
file
system // "Materi baru ditambahkan", "Siswa baru bergabung"
}

Table chat_rooms {
room_id uuid [pk, default: `gen_random_uuid()`]
room_sch_id uuid [ref: > schools.sch_id]
room_name varchar(150)
room_type chat_room_type

// pointer ke konteks akademik Wiyata
room_ref_type varchar(20) // 'class' | 'subject' | null
room_ref_id uuid // cls_id atau scl_id, null untuk DM

created_by uuid [ref: > users.usr_id]
created_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(room_sch_id, room_ref_type, room_ref_id) [unique, name: 'chat_rooms_room_sch_id_room_ref_type_room_ref_id_key', note: 'implemented as an auto-named UNIQUE table constraint in the live database, not a manually-named index']
}
}

Table chat_room_members {
crm_id uuid [pk, default: `gen_random_uuid()`]
crm_room_id uuid [ref: > chat_rooms.room_id]
crm_usr_id uuid [ref: > users.usr_id]
crm_enr_id uuid [ref: > enrollments.enr_id]
crm_role varchar(20) [default: 'member'] // NEW — 'admin' | 'member'
joined_at timestamptz [default: `now()`]
left_at timestamptz

indexes {
(crm_room_id, crm_usr_id) [unique]
(crm_usr_id) [name: 'idx_chat_members_user']
}
}

Table chat_messages {
msg_id uuid [pk, default: `gen_random_uuid()`]
msg_room_id uuid [ref: > chat_rooms.room_id]
msg_usr_id uuid [ref: > users.usr_id]
msg_content text
msg_type chat_message_type

// thread/reply
msg_reply_to uuid [ref: > chat_messages.msg_id]

// link ke konteks akademik (opsional)
// guru share tugas/materi langsung dari chat
msg_ref_type source_type // pakai enum source_type yang sudah ada
msg_ref_id uuid

created_at timestamptz [default: `now()`]
deleted_at timestamptz

indexes {
(msg_room_id, created_at) [name: 'idx_chat_messages_room', note: 'created_at ordered DESC in the live database']
}
}

Table chat_attachments {
cat_id uuid [pk, default: `gen_random_uuid()`]
cat_msg_id uuid [ref: > chat_messages.msg_id]

// pakai medias yang sudah ada di Wiyata
cat_med_id uuid [ref: > medias.med_id]

created_at timestamptz [default: `now()`]
}

Table chat_read_receipts {
rct_id uuid [pk, default: `gen_random_uuid()`]
rct_room_id uuid [ref: > chat_rooms.room_id]
rct_usr_id uuid [ref: > users.usr_id]
last_read_msg_id uuid [ref: > chat_messages.msg_id]
last_read_at timestamptz

indexes {
(rct_room_id, rct_usr_id) [unique]
(rct_usr_id) [name: 'idx_chat_receipts_user']
}
}

Table student_notes {
snt_id uuid [pk, default: `gen_random_uuid()`]
snt_sch_id uuid [ref: > schools.sch_id]
snt_usr_id uuid [ref: > users.usr_id]
snt_mat_id uuid [ref: > materials.mat_id]
snt_content text // private plain-text material note, max 10,000 characters at API layer
created_at timestamptz [default: `now()`]
updated_at timestamptz [default: `now()`]

indexes {
(snt_usr_id, snt_mat_id) [unique]
(snt_usr_id) [name: 'idx_student_notes_user']
}
}

// Material-only Notes MVP: one private note per student per material.
// Teacher/admin access, assignment notes, global notes, and soft delete are intentionally unsupported.
