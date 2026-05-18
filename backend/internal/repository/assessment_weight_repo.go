package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type AssessmentWeightRepository interface {
	Create(weight *domain.AssessmentWeight) error
	GetBySubject(subjectID string) ([]*domain.AssessmentWeight, error)
	DeleteBySubject(subjectID string) error
	GetTotalWeightBySubject(subjectID string) (float64, error)
}

type assessmentWeightRepository struct {
	db *gorm.DB
}

func NewAssessmentWeightRepository(db *gorm.DB) AssessmentWeightRepository {
	return &assessmentWeightRepository{db: db}
}

func (r *assessmentWeightRepository) Create(weight *domain.AssessmentWeight) error {
	return r.db.Create(weight).Error
}

func (r *assessmentWeightRepository) GetBySubject(subjectID string) ([]*domain.AssessmentWeight, error) {
	var weights []*domain.AssessmentWeight
	err := r.db.
		Preload("Subject").
		Preload("Category").
		Where("asw_sub_id = ?", subjectID).
		Find(&weights).Error
	return weights, err
}

func (r *assessmentWeightRepository) DeleteBySubject(subjectID string) error {
	return r.db.Where("asw_sub_id = ?", subjectID).Delete(&domain.AssessmentWeight{}).Error
}

func (r *assessmentWeightRepository) GetTotalWeightBySubject(subjectID string) (float64, error) {
	var total float64
	err := r.db.Model(&domain.AssessmentWeight{}).
		Where("asw_sub_id = ?", subjectID).
		Select("COALESCE(SUM(asw_weight), 0)").
		Scan(&total).Error
	return total, err
}
