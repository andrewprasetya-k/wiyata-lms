package repository

import (
	"backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type EnrollmentRepository interface {
	Create(enr *domain.Enrollment) error
	GetByID(id string) (*domain.Enrollment, error)
	GetByClassAndSchoolUser(classID string, schoolUserID string) (*domain.Enrollment, error)
	GetByClass(classID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error)
	GetByMember(schoolUserID string) ([]*domain.Enrollment, error)
	Update(id string, role string) error
	Reactivate(id string, role string) error
	SoftDelete(id string) error
	CheckExists(classID, schoolUserID string) (bool, error)
	BelongsToSchool(enrollmentID string, schoolID string) (bool, error)
	ActiveBelongsToSchool(enrollmentID string, schoolID string) (bool, error)
	HasTeacherSubjectClassAssignment(classID string, schoolUserID string, schoolID string) (bool, error)
	GetStudentUserIDsByClass(classID string) ([]string, error)
	GetMemberUserIDsByClass(classID string) ([]string, error)
	UserEnrolledInClassAsRole(userID string, schoolID string, classID string, role string) (bool, error)
	BulkCloseBySchoolUser(schoolUserID string, leftAt time.Time) error
}

type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) Create(enr *domain.Enrollment) error {
	return r.db.Create(enr).Error
}

func (r *enrollmentRepository) GetByID(id string) (*domain.Enrollment, error) {
	var enr domain.Enrollment
	err := r.db.Preload("SchoolUser.User").Preload("Class").
		Where("enr_id = ?", id).First(&enr).Error
	return &enr, err
}

func (r *enrollmentRepository) GetByClassAndSchoolUser(classID string, schoolUserID string) (*domain.Enrollment, error) {
	var enr domain.Enrollment
	err := r.db.
		Where("enr_cls_id = ? AND enr_scu_id = ?", classID, schoolUserID).
		First(&enr).Error
	return &enr, err
}

func (r *enrollmentRepository) GetByClass(classID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error) {
	var results []*domain.Enrollment
	var total int64

	query := r.db.Model(&domain.Enrollment{}).
		Preload("SchoolUser.User").
		Where("enr_cls_id = ? AND left_at IS NULL", classID)

	// Search by user name or email
	if search != "" {
		query = query.Joins("JOIN edv.school_users ON school_users.scu_id = enrollments.enr_scu_id AND school_users.deleted_at IS NULL").
			Joins("JOIN edv.users ON users.usr_id = school_users.scu_usr_id").
			Where("users.usr_nama_lengkap ILIKE ? OR users.usr_email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("joined_at desc").Find(&results).Error
	return results, total, err
}

func (r *enrollmentRepository) GetByMember(schoolUserID string) ([]*domain.Enrollment, error) {
	var results []*domain.Enrollment
	err := r.db.Preload("Class.School").Preload("Class.Term.AcademicYear").
		Where("enr_scu_id = ? AND left_at IS NULL", schoolUserID).Find(&results).Error
	return results, err
}

func (r *enrollmentRepository) Update(id string, role string) error {
	result := r.db.Model(&domain.Enrollment{}).Where("enr_id = ? AND left_at IS NULL", id).Update("enr_role", role)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *enrollmentRepository) Reactivate(id string, role string) error {
	result := r.db.Model(&domain.Enrollment{}).
		Where("enr_id = ?", id).
		Updates(map[string]any{
			"enr_role": role,
			"left_at":  nil,
		})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *enrollmentRepository) SoftDelete(id string) error {
	result := r.db.Model(&domain.Enrollment{}).
		Where("enr_id = ? AND left_at IS NULL", id).
		Update("left_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *enrollmentRepository) CheckExists(classID, schoolUserID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("enr_cls_id = ? AND enr_scu_id = ? AND left_at IS NULL", classID, schoolUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *enrollmentRepository) BelongsToSchool(enrollmentID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("enr_id = ? AND enr_sch_id = ?", enrollmentID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *enrollmentRepository) ActiveBelongsToSchool(enrollmentID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("enr_id = ? AND enr_sch_id = ? AND left_at IS NULL", enrollmentID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *enrollmentRepository) HasTeacherSubjectClassAssignment(classID string, schoolUserID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Where("sc.scl_cls_id = ? AND sc.scl_scu_id = ?", classID, schoolUserID).
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *enrollmentRepository) GetStudentUserIDsByClass(classID string) ([]string, error) {
	var userIDs []string
	err := r.db.Model(&domain.Enrollment{}).
		Joins("JOIN edv.school_users ON school_users.scu_id = enrollments.enr_scu_id AND school_users.deleted_at IS NULL").
		Where("enrollments.enr_cls_id = ? AND enrollments.enr_role = ? AND enrollments.left_at IS NULL", classID, "student").
		Pluck("school_users.scu_usr_id", &userIDs).Error
	return userIDs, err
}

func (r *enrollmentRepository) GetMemberUserIDsByClass(classID string) ([]string, error) {
	var userIDs []string
	err := r.db.Model(&domain.Enrollment{}).
		Joins("JOIN edv.school_users ON school_users.scu_id = enrollments.enr_scu_id AND school_users.deleted_at IS NULL").
		Where("enrollments.enr_cls_id = ? AND enrollments.left_at IS NULL", classID).
		Pluck("school_users.scu_usr_id", &userIDs).Error
	return userIDs, err
}

func (r *enrollmentRepository) BulkCloseBySchoolUser(schoolUserID string, leftAt time.Time) error {
	return r.db.Model(&domain.Enrollment{}).
		Where("enr_scu_id = ? AND left_at IS NULL", schoolUserID).
		Update("left_at", leftAt).Error
}

func (r *enrollmentRepository) UserEnrolledInClassAsRole(userID string, schoolID string, classID string, role string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Joins("JOIN edv.school_users ON school_users.scu_id = enrollments.enr_scu_id AND school_users.deleted_at IS NULL").
		Where("school_users.scu_usr_id = ? AND school_users.scu_sch_id = ? AND school_users.deleted_at IS NULL", userID, schoolID).
		Where("enrollments.enr_sch_id = ? AND enrollments.enr_cls_id = ?", schoolID, classID).
		Where("enrollments.enr_role = ? AND enrollments.left_at IS NULL", role).
		Count(&count).Error
	return count > 0, err
}
