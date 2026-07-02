package repository

import (
	"backend/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type SchoolRegistrationRequestRepository interface {
	Create(request *domain.SchoolRegistrationRequest) error
	HasPendingDuplicate(schoolName string, picEmail string) (bool, error)
}

type schoolRegistrationRequestRepository struct {
	db *gorm.DB
}

func NewSchoolRegistrationRequestRepository(db *gorm.DB) SchoolRegistrationRequestRepository {
	return &schoolRegistrationRequestRepository{db: db}
}

func (r *schoolRegistrationRequestRepository) Create(request *domain.SchoolRegistrationRequest) error {
	return r.db.Create(request).Error
}

func (r *schoolRegistrationRequestRepository) HasPendingDuplicate(schoolName string, picEmail string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SchoolRegistrationRequest{}).
		Where("srr_status = ?", domain.SchoolRegistrationPending).
		Where(
			"LOWER(TRIM(srr_school_name)) = ? OR LOWER(TRIM(srr_pic_email)) = ?",
			strings.ToLower(strings.TrimSpace(schoolName)),
			strings.ToLower(strings.TrimSpace(picEmail)),
		).
		Count(&count).Error
	return count > 0, err
}
