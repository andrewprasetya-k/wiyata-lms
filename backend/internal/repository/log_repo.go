package repository

import (
	"backend/internal/domain"
	"strings"
	"time"

	"gorm.io/gorm"
)

type LogFilter struct {
	SchoolID      *string
	Scope         string
	Action        string
	EntityType    string
	Severity      string
	ActorUserID   string
	DateFrom      *time.Time
	DateTo        *time.Time
	CorrelationID string
	Search        string
	Page          int
	Limit         int
}

type LogRepository interface {
	Create(log *domain.Log) error
	GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByCorrelationID(correlationID string) ([]*domain.Log, error)
	Search(filter LogFilter) ([]*domain.Log, int64, error)

	GetByID(id string) (*domain.Log, error)

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

func (r *logRepository) Search(filter LogFilter) ([]*domain.Log, int64, error) {
	var logs []*domain.Log
	var total int64

	query := r.db.Model(&domain.Log{}).Preload("User").Preload("School")

	if filter.SchoolID != nil {
		query = query.Where("log_sch_id = ?", *filter.SchoolID)
	}
	if filter.Scope != "" {
		query = query.Where("scope = ?", filter.Scope)
	}
	if filter.Action != "" {
		query = query.Where("log_action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.ActorUserID != "" {
		query = query.Where("log_usr_id = ?", filter.ActorUserID)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}
	if filter.CorrelationID != "" {
		query = query.Where("correlation_id = ?", filter.CorrelationID)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Joins("LEFT JOIN edv.users ON edv.users.usr_id = logs.log_usr_id").
			Where("logs.log_action ILIKE ? OR logs.entity_type ILIKE ? OR edv.users.usr_nama_lengkap ILIKE ? OR edv.users.usr_email ILIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	err := query.Order("logs.created_at desc").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *logRepository) GetByID(id string) (*domain.Log, error) {
	var log domain.Log
	err := r.db.Preload("User").Preload("School").Where("log_id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *logRepository) WithTx(tx *gorm.DB) LogRepository {
	return &logRepository{db: tx}
}
