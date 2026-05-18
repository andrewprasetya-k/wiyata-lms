package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type GradeRepository interface {
	GetAssessmentsByStudentAndSubject(userID string, subjectID string) ([]*domain.Assessment, error)
	GetStudentsBySubjectClass(subjectClassID string) ([]*domain.User, error)
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
		Distinct().
		Find(&users).Error

	return users, err
}
