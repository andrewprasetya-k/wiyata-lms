package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(media *domain.Media) error
	GetByID(id string) (*domain.Media, error)
	GetByIDs(ids []string) ([]*domain.Media, error)
	GetByOwner(ownerType domain.OwnerType, ownerID string) ([]*domain.Media, error)
	Delete(id string) error
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(media *domain.Media) error {
	return r.db.Create(media).Error
}

func (r *mediaRepository) GetByID(id string) (*domain.Media, error) {
	var media domain.Media
	err := r.db.Where("med_id = ?", id).First(&media).Error
	return &media, err
}

func (r *mediaRepository) GetByIDs(ids []string) ([]*domain.Media, error) {
	var results []*domain.Media
	err := r.db.Where("med_id IN ?", ids).Find(&results).Error
	return results, err
}

func (r *mediaRepository) GetByOwner(ownerType domain.OwnerType, ownerID string) ([]*domain.Media, error) {
	var results []*domain.Media
	err := r.db.Where("med_owner_type = ? AND med_owner_id = ?", ownerType, ownerID).Find(&results).Error
	return results, err
}

func (r *mediaRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Media{}, "med_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
