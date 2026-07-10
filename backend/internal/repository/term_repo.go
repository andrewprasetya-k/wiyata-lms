package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type TermRepository interface {
	Create(term *domain.Term) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Term, int64, error)
	GetByAcademicYear(acyID string, schoolID string) ([]*domain.Term, error)
	GetByID(id string) (*domain.Term, error)
	Update(term *domain.Term) error
	Delete(id string) error
	DeactivateAllExcept(acyID string, activeID string) error
	SetActiveStatus(id string, isActive bool) error
	CheckDuplicateName(acyID string, name string, excludeID string) (bool, error)
	HasClasses(id string) (bool, error)
}

type termRepository struct {
	db *gorm.DB
}

func NewTermRepository(db *gorm.DB) TermRepository {
	return &termRepository{db: db}
}

func (r *termRepository) Create(term *domain.Term) error {
	return r.db.Create(term).Error
}

func (r *termRepository) FindAll(schoolID string, search string, page int, limit int) ([]*domain.Term, int64, error) {
	var terms []*domain.Term
	var total int64

	query := r.db.Model(&domain.Term{}).Preload("AcademicYear.School").
		Joins("JOIN edv.academic_years ON edv.academic_years.acy_id = edv.terms.trm_acy_id").
		Where("edv.academic_years.acy_sch_id = ?", schoolID)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("edv.terms.trm_name ILIKE ?", searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("edv.terms.created_at desc").Find(&terms).Error
	return terms, total, err
}

func (r *termRepository) GetByAcademicYear(acyID string, schoolID string) ([]*domain.Term, error) {
	var terms []*domain.Term
	err := r.db.Preload("AcademicYear.School").
		Joins("JOIN edv.academic_years ON edv.academic_years.acy_id = edv.terms.trm_acy_id").
		Where("edv.terms.trm_acy_id = ? AND edv.academic_years.acy_sch_id = ?", acyID, schoolID).
		Order("edv.terms.created_at asc").Find(&terms).Error
	return terms, err
}

func (r *termRepository) GetByID(id string) (*domain.Term, error) {
	var term domain.Term
	err := r.db.Preload("AcademicYear.School").Where("trm_id = ?", id).First(&term).Error
	return &term, err
}

func (r *termRepository) Update(term *domain.Term) error {
	result := r.db.Save(term)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *termRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Term{}, "trm_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *termRepository) DeactivateAllExcept(acyID string, activeID string) error {
	return r.db.Model(&domain.Term{}).
		Where("trm_acy_id = ? AND trm_id != ?", acyID, activeID).
		Update("is_active", false).Error
}

func (r *termRepository) SetActiveStatus(id string, isActive bool) error {
	result := r.db.Model(&domain.Term{}).Where("trm_id = ?", id).Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *termRepository) CheckDuplicateName(acyID string, name string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Term{}).Where("trm_acy_id = ? AND trm_name = ?", acyID, name)
	if excludeID != "" {
		query = query.Where("trm_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *termRepository) HasClasses(id string) (bool, error) {
	var count int64
	// cek tabel edv.classes
	err := r.db.Table("edv.classes").Where("cls_trm_id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}
