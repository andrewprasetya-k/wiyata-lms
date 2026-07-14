package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type AssessmentWeightRepository interface {
	Create(weight *domain.AssessmentWeight) error
	GetBySubject(subjectID string) ([]*domain.AssessmentWeight, error)
	GetBySubjects(subjectIDs []string) (map[string][]*domain.AssessmentWeight, error)
	DeleteBySubject(subjectID string) error
	ReplaceBySubject(subjectID string, weights []*domain.AssessmentWeight) error
	GetTotalWeightBySubject(subjectID string) (float64, error)
	SubjectBelongsToSchool(subjectID string, schoolID string) (bool, error)
	CountCategoriesInSchool(categoryIDs []string, schoolID string) (int64, error)
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

func (r *assessmentWeightRepository) GetBySubjects(subjectIDs []string) (map[string][]*domain.AssessmentWeight, error) {
	result := make(map[string][]*domain.AssessmentWeight, len(subjectIDs))
	if len(subjectIDs) == 0 {
		return result, nil
	}

	var weights []*domain.AssessmentWeight
	err := r.db.
		Preload("Subject").
		Preload("Category").
		Where("asw_sub_id IN ?", subjectIDs).
		Find(&weights).Error
	if err != nil {
		return nil, err
	}
	for _, w := range weights {
		result[w.SubjectID] = append(result[w.SubjectID], w)
	}
	return result, nil
}

func (r *assessmentWeightRepository) DeleteBySubject(subjectID string) error {
	return r.db.Where("asw_sub_id = ?", subjectID).Delete(&domain.AssessmentWeight{}).Error
}

func (r *assessmentWeightRepository) ReplaceBySubject(subjectID string, weights []*domain.AssessmentWeight) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("asw_sub_id = ?", subjectID).Delete(&domain.AssessmentWeight{}).Error; err != nil {
			return err
		}
		for _, w := range weights {
			if err := tx.Create(w).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *assessmentWeightRepository) GetTotalWeightBySubject(subjectID string) (float64, error) {
	var total float64
	err := r.db.Model(&domain.AssessmentWeight{}).
		Where("asw_sub_id = ?", subjectID).
		Select("COALESCE(SUM(asw_weight), 0)").
		Scan(&total).Error
	return total, err
}

func (r *assessmentWeightRepository) SubjectBelongsToSchool(subjectID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Subject{}).
		Where("sub_id = ? AND sub_sch_id = ?", subjectID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *assessmentWeightRepository) CountCategoriesInSchool(categoryIDs []string, schoolID string) (int64, error) {
	if len(categoryIDs) == 0 {
		return 0, nil
	}

	var count int64
	err := r.db.Model(&domain.AssignmentCategory{}).
		Where("asc_id IN ? AND asc_sch_id = ?", categoryIDs, schoolID).
		Count(&count).Error
	return count, err
}
