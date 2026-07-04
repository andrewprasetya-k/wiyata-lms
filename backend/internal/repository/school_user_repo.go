package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type SchoolUserRepository interface {
	Create(scu *domain.SchoolUser) error
	GetBySchool(schoolID string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error)
	GetBySchoolWithDeleted(schoolID string, search string, role string, includeDeleted bool, page int, limit int) ([]*domain.SchoolUser, int64, error)
	GetByUser(userID string) ([]*domain.SchoolUser, error)
	Delete(id string) error
	SoftDeleteByIDInSchool(schoolUserID string, schoolID string) error
	RestoreByIDInSchool(schoolUserID string, schoolID string) error
	IsEnrolled(userID string, schoolID string) (bool, error)
	BelongsToSchool(schoolUserID string, schoolID string) (bool, error)
	FindByUserAndSchoolIncludingDeleted(userID string, schoolID string) (*domain.SchoolUser, error)
}

type schoolUserRepository struct {
	db *gorm.DB
}

func NewSchoolUserRepository(db *gorm.DB) SchoolUserRepository {
	return &schoolUserRepository{db: db}
}

func (r *schoolUserRepository) Create(scu *domain.SchoolUser) error {
	return r.db.Create(scu).Error
}

func (r *schoolUserRepository) GetBySchool(schoolID string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error) {
	return r.GetBySchoolWithDeleted(schoolID, search, "", false, page, limit)
}

func (r *schoolUserRepository) GetBySchoolWithDeleted(schoolID string, search string, role string, includeDeleted bool, page int, limit int) ([]*domain.SchoolUser, int64, error) {
	var members []*domain.SchoolUser
	var total int64

	query := r.db.Model(&domain.SchoolUser{}).
		Preload("User").
		Preload("Roles.Role").
		Where("scu_sch_id = ?", schoolID)
	if includeDeleted {
		query = query.Unscoped()
	}

	// Search by user name or email
	if search != "" {
		query = query.Joins("JOIN edv.users ON users.usr_id = school_users.scu_usr_id AND users.deleted_at IS NULL").
			Where("users.usr_nama_lengkap ILIKE ? OR users.usr_email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role != "" {
		query = query.
			Joins("JOIN edv.user_roles ur ON ur.urol_scu_id = school_users.scu_id").
			Joins("JOIN edv.roles r ON r.rol_id = ur.urol_rol_id").
			Where("r.rol_name = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&members).Error
	return members, total, err
}

func (r *schoolUserRepository) GetByUser(userID string) ([]*domain.SchoolUser, error) {
	var schools []*domain.SchoolUser
	err := r.db.Preload("School").
		Preload("Roles.Role").
		Where("scu_usr_id = ?", userID).
		Where("school_users.deleted_at IS NULL").
		Find(&schools).Error
	return schools, err
}

func (r *schoolUserRepository) Delete(userId string) error {
	result := r.db.Delete(&domain.SchoolUser{}, "scu_usr_id = ?", userId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolUserRepository) SoftDeleteByIDInSchool(schoolUserID string, schoolID string) error {
	result := r.db.Where("scu_id = ? AND scu_sch_id = ?", schoolUserID, schoolID).Delete(&domain.SchoolUser{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolUserRepository) RestoreByIDInSchool(schoolUserID string, schoolID string) error {
	result := r.db.Unscoped().Model(&domain.SchoolUser{}).
		Where("scu_id = ? AND scu_sch_id = ?", schoolUserID, schoolID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolUserRepository) IsEnrolled(userID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SchoolUser{}).
		Where("scu_usr_id = ? AND scu_sch_id = ?", userID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *schoolUserRepository) BelongsToSchool(schoolUserID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SchoolUser{}).
		Where("scu_id = ? AND scu_sch_id = ?", schoolUserID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *schoolUserRepository) FindByUserAndSchoolIncludingDeleted(userID string, schoolID string) (*domain.SchoolUser, error) {
	var schoolUser domain.SchoolUser
	err := r.db.Unscoped().
		Where("scu_usr_id = ? AND scu_sch_id = ?", userID, schoolID).
		First(&schoolUser).Error
	return &schoolUser, err
}
