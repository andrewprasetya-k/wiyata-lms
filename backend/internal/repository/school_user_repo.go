package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type SchoolUserRepository interface {
	Create(scu *domain.SchoolUser) error
	GetBySchool(schoolID string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error)
	GetByUser(userID string) ([]*domain.SchoolUser, error)
	Delete(id string) error
	IsEnrolled(userID string, schoolID string) (bool, error)
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
	var members []*domain.SchoolUser
	var total int64

	query := r.db.Model(&domain.SchoolUser{}).
		Preload("User").
		Preload("Roles.Role").
		Where("scu_sch_id = ?", schoolID)

	// Search by user name or email
	if search != "" {
		query = query.Joins("JOIN edv.users ON users.usr_id = school_users.scu_usr_id").
			Where("users.usr_nama_lengkap ILIKE ? OR users.usr_email ILIKE ?", "%"+search+"%", "%"+search+"%")
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

func (r *schoolUserRepository) IsEnrolled(userID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SchoolUser{}).
		Where("scu_usr_id = ? AND scu_sch_id = ?", userID, schoolID).
		Count(&count).Error
	return count > 0, err
}
