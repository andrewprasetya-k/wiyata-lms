package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type LogRepository interface {
	Create(log *domain.Log) error
	GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error)
	// GetByCorrelationID returns every log row sharing a correlation_id,
	// oldest first. Used by Phase 10.5's bulk-import parent+child pattern
	// (Phase 10.2 §5) to reassemble a batch.
	GetByCorrelationID(correlationID string) ([]*domain.Log, error)
	// WithTx returns a repository instance bound to an existing transaction,
	// following the same convention as AttachmentRepository/AssignmentRepository.
	// Needed so bulk actions (e.g. CSV import) can write their audit rows
	// inside the same DB transaction as the business rows they describe.
	WithTx(tx *gorm.DB) LogRepository
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(log *domain.Log) error {
	return r.db.Create(log).Error
}

func (r *logRepository) GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error) {
	var logs []*domain.Log
	var total int64

	query := r.db.Model(&domain.Log{}).Preload("User").Where("log_sch_id = ?", schoolID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&logs).Error
	return logs, total, err
}

func (r *logRepository) GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error) {
	var logs []*domain.Log
	var total int64

	query := r.db.Model(&domain.Log{}).Where("log_usr_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&logs).Error
	return logs, total, err
}

func (r *logRepository) GetByCorrelationID(correlationID string) ([]*domain.Log, error) {
	var logs []*domain.Log
	err := r.db.Where("correlation_id = ?", correlationID).Order("created_at asc").Find(&logs).Error
	return logs, err
}

func (r *logRepository) WithTx(tx *gorm.DB) LogRepository {
	return &logRepository{db: tx}
}
