package repository

import (
	"time"

	"gorm.io/gorm"
)

type ActivityRepository interface {
	GetStudentAssignmentDue(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetStudentMaterialCreated(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetStudentFeedPosted(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetStudentAssignmentGraded(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetTeacherSubmissionReceived(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetTeacherSubmissionPendingReview(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetTeacherAssignmentDue(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
	GetTeacherFeedComments(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error)
}

type ActivityRow struct {
	SourceID       string    `gorm:"column:source_id"`
	ActivityType   string    `gorm:"column:activity_type"`
	Title          string    `gorm:"column:title"`
	Description    string    `gorm:"column:description"`
	EventAt        time.Time `gorm:"column:event_at"`
	Priority       string    `gorm:"column:priority"`
	SubjectID      string    `gorm:"column:subject_id"`
	SubjectName    string    `gorm:"column:subject_name"`
	SubjectCode    string    `gorm:"column:subject_code"`
	SubjectColor   string    `gorm:"column:subject_color"`
	ClassID        string    `gorm:"column:class_id"`
	ClassName      string    `gorm:"column:class_name"`
	ClassCode      string    `gorm:"column:class_code"`
	Link           string    `gorm:"column:link"`
	AssignmentID   string    `gorm:"column:assignment_id"`
	SubjectClassID string    `gorm:"column:subject_class_id"`
	MaterialID     string    `gorm:"column:material_id"`
	FeedID         string    `gorm:"column:feed_id"`
	CommentID      string    `gorm:"column:comment_id"`
	SubmissionID   string    `gorm:"column:submission_id"`
	StudentID      string    `gorm:"column:student_id"`
	StudentName    string    `gorm:"column:student_name"`
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) GetStudentAssignmentDue(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			a.asg_id AS source_id,
			'assignment_due' AS activity_type,
			a.asg_title AS title,
			CONCAT('Tenggat tugas ', sub.sub_name, ' · ', c.cls_title) AS description,
			a.asg_deadline AS event_at,
			'high' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/student/subjects/', sc.scl_id, '/assignments/', a.asg_id) AS link,
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id
		FROM edv.assignments a
		JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id
		JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL
		WHERE a.asg_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND scu.scu_usr_id = ?
			AND scu.scu_sch_id = ?
			AND e.enr_sch_id = ?
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND a.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND a.asg_deadline >= ?
			AND a.asg_deadline < ?
			AND NOT EXISTS (
				SELECT 1
				FROM edv.submissions s
				WHERE s.sbm_asg_id = a.asg_id
					AND s.sbm_usr_id = ?
					AND s.sbm_sch_id = ?
					AND s.deleted_at IS NULL
			)
	`, schoolID, schoolID, schoolID, userID, schoolID, schoolID, from, to, userID, schoolID).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetStudentMaterialCreated(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			m.mat_id AS source_id,
			'material_created' AS activity_type,
			m.mat_title AS title,
			CONCAT('Materi baru ', sub.sub_name, ' · ', c.cls_title) AS description,
			m.created_at AS event_at,
			'normal' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/student/subjects/', sc.scl_id, '/materials/', m.mat_id) AS link,
			sc.scl_id AS subject_class_id,
			m.mat_id AS material_id
		FROM edv.materials m
		JOIN edv.subject_classes sc ON sc.scl_id = m.mat_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id
		JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL
		WHERE m.mat_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND scu.scu_usr_id = ?
			AND scu.scu_sch_id = ?
			AND e.enr_sch_id = ?
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND m.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND m.created_at >= ?
			AND m.created_at < ?
	`, schoolID, schoolID, schoolID, userID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetStudentFeedPosted(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			f.fds_id AS source_id,
			'feed_posted' AS activity_type,
			'Pengumuman kelas' AS title,
			CONCAT('Pengumuman baru di ', c.cls_title) AS description,
			f.created_at AS event_at,
			'normal' AS priority,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			'/student/feed' AS link,
			f.fds_id AS feed_id
		FROM edv.feeds f
		JOIN edv.classes c ON c.cls_id = f.fds_cls_id
		JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id
		JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL
		WHERE f.fds_sch_id = ?
			AND c.cls_sch_id = ?
			AND scu.scu_usr_id = ?
			AND scu.scu_sch_id = ?
			AND e.enr_sch_id = ?
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND f.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND f.created_at >= ?
			AND f.created_at < ?
	`, schoolID, schoolID, userID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetStudentAssignmentGraded(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			asm.asm_id AS source_id,
			'assignment_graded' AS activity_type,
			a.asg_title AS title,
			CONCAT('Nilai tugas ', sub.sub_name, ' sudah tersedia') AS description,
			asm.assessed_at AS event_at,
			'normal' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/student/subjects/', sc.scl_id, '/assignments/', a.asg_id) AS link,
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
			s.sbm_id AS submission_id
		FROM edv.assessments asm
		JOIN edv.submissions s ON s.sbm_id = asm.asm_sbm_id
		JOIN edv.assignments a ON a.asg_id = s.sbm_asg_id
		JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id
		JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL
		WHERE s.sbm_usr_id = ?
			AND s.sbm_sch_id = ?
			AND a.asg_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND scu.scu_usr_id = ?
			AND scu.scu_sch_id = ?
			AND e.enr_sch_id = ?
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND s.deleted_at IS NULL
			AND a.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND asm.assessed_at >= ?
			AND asm.assessed_at < ?
	`, userID, schoolID, schoolID, schoolID, schoolID, userID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetTeacherSubmissionReceived(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			s.sbm_id AS source_id,
			'submission_received' AS activity_type,
			a.asg_title AS title,
			CONCAT(student.usr_full_name, ' mengumpulkan tugas ', sub.sub_name) AS description,
			s.submitted_at AS event_at,
			'normal' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/teacher/assignments/', a.asg_id, '/review') AS link,
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
			s.sbm_id AS submission_id,
			student.usr_id AS student_id,
			student.usr_full_name AS student_name
		FROM edv.submissions s
		JOIN edv.users student ON student.usr_id = s.sbm_usr_id AND student.deleted_at IS NULL
		JOIN edv.assignments a ON a.asg_id = s.sbm_asg_id
		JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = c.cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		WHERE teacher_scu.scu_usr_id = ?
			AND teacher_scu.scu_sch_id = ?
			AND teacher_e.enr_sch_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND s.sbm_sch_id = ?
			AND a.asg_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND s.deleted_at IS NULL
			AND a.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND s.submitted_at >= ?
			AND s.submitted_at < ?
			AND EXISTS (SELECT 1 FROM edv.assessments asm WHERE asm.asm_sbm_id = s.sbm_id)
	`, userID, schoolID, schoolID, schoolID, schoolID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetTeacherSubmissionPendingReview(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			s.sbm_id AS source_id,
			'submission_pending_review' AS activity_type,
			a.asg_title AS title,
			CONCAT(student.usr_full_name, ' menunggu penilaian') AS description,
			s.submitted_at AS event_at,
			'high' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/teacher/assignments/', a.asg_id, '/review') AS link,
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
			s.sbm_id AS submission_id,
			student.usr_id AS student_id,
			student.usr_full_name AS student_name
		FROM edv.submissions s
		JOIN edv.users student ON student.usr_id = s.sbm_usr_id AND student.deleted_at IS NULL
		JOIN edv.assignments a ON a.asg_id = s.sbm_asg_id
		JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = c.cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		WHERE teacher_scu.scu_usr_id = ?
			AND teacher_scu.scu_sch_id = ?
			AND teacher_e.enr_sch_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND s.sbm_sch_id = ?
			AND a.asg_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND s.deleted_at IS NULL
			AND a.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND s.submitted_at >= ?
			AND s.submitted_at < ?
			AND NOT EXISTS (SELECT 1 FROM edv.assessments asm WHERE asm.asm_sbm_id = s.sbm_id)
	`, userID, schoolID, schoolID, schoolID, schoolID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetTeacherAssignmentDue(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT
			a.asg_id AS source_id,
			'assignment_due' AS activity_type,
			a.asg_title AS title,
			CONCAT('Tenggat tugas ', sub.sub_name, ' · ', c.cls_title) AS description,
			a.asg_deadline AS event_at,
			'high' AS priority,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			CONCAT('/teacher/assignments/', a.asg_id, '/review') AS link,
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id
		FROM edv.assignments a
		JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = c.cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		WHERE teacher_scu.scu_usr_id = ?
			AND teacher_scu.scu_sch_id = ?
			AND teacher_e.enr_sch_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND a.asg_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.cls_sch_id = ?
			AND a.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND a.asg_deadline >= ?
			AND a.asg_deadline < ?
	`, userID, schoolID, schoolID, schoolID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}

func (r *activityRepository) GetTeacherFeedComments(userID string, schoolID string, from time.Time, to time.Time) ([]ActivityRow, error) {
	var rows []ActivityRow
	err := r.db.Raw(`
		SELECT DISTINCT ON (cmn.cmn_id)
			cmn.cmn_id AS source_id,
			'feed_comment' AS activity_type,
			'Komentar feed kelas' AS title,
			CONCAT(commenter.usr_full_name, ' berkomentar di ', c.cls_title) AS description,
			cmn.created_at AS event_at,
			'normal' AS priority,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			'/teacher/feed' AS link,
			f.fds_id AS feed_id,
			cmn.cmn_id AS comment_id
		FROM edv.comments cmn
		JOIN edv.users commenter ON commenter.usr_id = cmn.cmn_usr_id AND commenter.deleted_at IS NULL
		JOIN edv.feeds f ON f.fds_id = cmn.cmn_source_id
		JOIN edv.classes c ON c.cls_id = f.fds_cls_id
		JOIN edv.subject_classes sc ON sc.scl_cls_id = c.cls_id
		JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = c.cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		WHERE cmn.cmn_source_type = 'feed'
			AND cmn.cmn_sch_id = ?
			AND f.fds_sch_id = ?
			AND c.cls_sch_id = ?
			AND teacher_scu.scu_usr_id = ?
			AND teacher_scu.scu_sch_id = ?
			AND teacher_e.enr_sch_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND cmn.deleted_at IS NULL
			AND f.deleted_at IS NULL
			AND c.deleted_at IS NULL
			AND cmn.created_at >= ?
			AND cmn.created_at < ?
		ORDER BY cmn.cmn_id, cmn.created_at ASC
	`, schoolID, schoolID, schoolID, userID, schoolID, schoolID, from, to).Scan(&rows).Error
	return rows, err
}
