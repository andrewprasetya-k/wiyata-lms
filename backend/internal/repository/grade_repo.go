package repository

import (
	"backend/internal/domain"
	"backend/internal/dto"

	"gorm.io/gorm"
)

type GradeRepository interface {
	GetAssessmentsByStudentAndSubject(userID string, subjectID string) ([]*domain.Assessment, error)
	GetStudentsBySubjectClass(subjectClassID string) ([]*domain.User, error)
	GetStudentGradebookClass(userID string, schoolID string, classID string) (*dto.StudentGradebookClassRow, error)
	GetStudentGradebookRows(userID string, schoolID string, classID string) ([]dto.StudentGradebookRow, error)
}

type gradeRepository struct {
	db *gorm.DB
}

func NewGradeRepository(db *gorm.DB) GradeRepository {
	return &gradeRepository{db: db}
}

func (r *gradeRepository) GetAssessmentsByStudentAndSubject(userID string, subjectID string) ([]*domain.Assessment, error) {
	var assessments []*domain.Assessment

	err := r.db.
		Joins("JOIN edv.submissions ON submissions.sbm_id = assessments.asm_sbm_id").
		Joins("JOIN edv.assignments ON assignments.asg_id = submissions.sbm_asg_id").
		Joins("JOIN edv.subject_classes ON subject_classes.scl_id = assignments.asg_scl_id").
		Preload("Submission.Assignment.Category").
		Where("submissions.sbm_usr_id = ?", userID).
		Where("subject_classes.scl_sub_id = ?", subjectID).
		Where("submissions.deleted_at IS NULL").
		Where("assignments.deleted_at IS NULL").
		Find(&assessments).Error

	return assessments, err
}

func (r *gradeRepository) GetStudentsBySubjectClass(subjectClassID string) ([]*domain.User, error) {
	var users []*domain.User

	err := r.db.
		Joins("JOIN edv.school_users ON school_users.scu_usr_id = users.usr_id").
		Joins("JOIN edv.enrollments ON enrollments.enr_scu_id = school_users.scu_id").
		Joins("JOIN edv.subject_classes ON subject_classes.scl_cls_id = enrollments.enr_cls_id").
		Where("subject_classes.scl_id = ?", subjectClassID).
		Where("enrollments.enr_role = ?", "student").
		Where("enrollments.left_at IS NULL").
		Distinct().
		Find(&users).Error

	return users, err
}

func (r *gradeRepository) GetStudentGradebookClass(userID string, schoolID string, classID string) (*dto.StudentGradebookClassRow, error) {
	var row dto.StudentGradebookClassRow
	result := r.db.Table("edv.classes c").
		Select("c.cls_id AS class_id, c.cls_title AS class_name, c.cls_code AS class_code").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id").
		Joins("JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id").
		Where("c.cls_id = ? AND c.cls_sch_id = ?", classID, schoolID).
		Where("scu.scu_usr_id = ? AND scu.scu_sch_id = ?", userID, schoolID).
		Where("e.enr_sch_id = ?", schoolID).
		Where("e.enr_role = ?", "student").
		Where("e.left_at IS NULL").
		Where("c.deleted_at IS NULL").
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *gradeRepository) GetStudentGradebookRows(userID string, schoolID string, classID string) ([]dto.StudentGradebookRow, error) {
	var rows []dto.StudentGradebookRow
	err := r.db.Table("edv.subject_classes sc").
		Select(`
			sc.scl_id AS subject_class_id,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			a.asg_id AS assignment_id,
			a.asg_title AS assignment_title,
			a.asg_asc_id AS category_id,
			ac.asc_name AS category_name,
			a.asg_deadline AS deadline,
			s.sbm_id AS submission_id,
			s.submitted_at AS submitted_at,
			asm.asm_score AS score,
			asm.asm_feedback AS feedback,
			asm.assessed_at AS assessed_at,
			assessor.usr_nama_lengkap AS assessor_name
		`).
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("LEFT JOIN edv.assignments a ON a.asg_scl_id = sc.scl_id AND a.asg_sch_id = ? AND a.deleted_at IS NULL", schoolID).
		Joins("LEFT JOIN edv.assignment_categories ac ON ac.asc_id = a.asg_asc_id").
		Joins("LEFT JOIN edv.submissions s ON s.sbm_asg_id = a.asg_id AND s.sbm_usr_id = ? AND s.sbm_sch_id = ? AND s.deleted_at IS NULL", userID, schoolID).
		Joins(`LEFT JOIN LATERAL (
			SELECT *
			FROM edv.assessments latest_asm
			WHERE latest_asm.asm_sbm_id = s.sbm_id
			ORDER BY latest_asm.assessed_at DESC, latest_asm.asm_id DESC
			LIMIT 1
		) asm ON true`).
		Joins("LEFT JOIN edv.users assessor ON assessor.usr_id = asm.assessed_by").
		Where("sc.scl_cls_id = ?", classID).
		Where("c.cls_sch_id = ? AND sub.sub_sch_id = ?", schoolID, schoolID).
		Where("c.deleted_at IS NULL").
		Order("sub.sub_name asc, a.asg_deadline asc nulls last, a.created_at asc").
		Scan(&rows).Error
	return rows, err
}
