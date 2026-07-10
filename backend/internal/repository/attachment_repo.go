package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type AttachmentRepository interface {
	Create(att *domain.Attachment) error
	GetBySource(sourceType domain.SourceType, sourceID string) ([]*domain.Attachment, error)
	GetBySources(sourceType domain.SourceType, sourceIDs []string) ([]*domain.Attachment, error)
	Delete(id string) error
	DeleteBySource(sourceType domain.SourceType, sourceID string) error
	ReplaceBySource(schoolID string, sourceType domain.SourceType, sourceID string, mediaIDs []string) error
	// WithTx returns a repository instance bound to an existing transaction, so
	// callers can compose multiple repository operations into one atomic unit.
	WithTx(tx *gorm.DB) AttachmentRepository
}

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) WithTx(tx *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: tx}
}

func (r *attachmentRepository) Create(att *domain.Attachment) error {
	return r.db.Create(att).Error
}

func (r *attachmentRepository) GetBySource(sourceType domain.SourceType, sourceID string) ([]*domain.Attachment, error) {
	var results []*domain.Attachment
	err := r.db.Preload("Media").
		Where("att_source_type = ? AND att_source_id = ?", sourceType, sourceID).
		Find(&results).Error
	return results, err
}

func (r *attachmentRepository) GetBySources(sourceType domain.SourceType, sourceIDs []string) ([]*domain.Attachment, error) {
	results := make([]*domain.Attachment, 0)
	if len(sourceIDs) == 0 {
		return results, nil
	}

	activeMediaIDs := r.db.Model(&domain.Media{}).Select("med_id")
	err := r.db.
		Preload("Media").
		Where("att_source_type = ? AND att_source_id IN ?", sourceType, sourceIDs).
		Where("att_med_id IN (?)", activeMediaIDs).
		Order("att_source_id ASC, created_at ASC").
		Find(&results).Error
	return results, err
}

func (r *attachmentRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Attachment{}, "att_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *attachmentRepository) DeleteBySource(sourceType domain.SourceType, sourceID string) error {
	return r.db.Where("att_source_type = ? AND att_source_id = ?", sourceType, sourceID).
		Delete(&domain.Attachment{}).Error
}

func (r *attachmentRepository) ReplaceBySource(schoolID string, sourceType domain.SourceType, sourceID string, mediaIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("att_source_type = ? AND att_source_id = ?", sourceType, sourceID).
			Delete(&domain.Attachment{}).Error; err != nil {
			return err
		}

		for _, mediaID := range mediaIDs {
			att := &domain.Attachment{
				SchoolID:   schoolID,
				SourceID:   sourceID,
				SourceType: sourceType,
				MediaID:    mediaID,
			}
			if err := tx.Create(att).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
