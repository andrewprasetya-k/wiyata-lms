package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type ClassRepository interface {
	Create(class *domain.Class) error
	FindAll(search string, schoolID string, termID string, page int, limit int) ([]*domain.Class, int64, error)
	GetByID(id string) (*domain.Class, error)
	Update(class *domain.Class) error
	Delete(id string) error
	CountEnrollmentsByClass(classID string) (int64, error)
	CountSubjectClassesByClass(classID string) (int64, error)
	CheckDuplicateCode(schoolID string, termID string, code string, excludeID string) (bool, error)
	GetSchoolIDByClass(classID string) (string, error)
}

type classRepository struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) Create(class *domain.Class) error {
	return r.db.Create(class).Error
}

func (r *classRepository) FindAll(search string, schoolID string, termID string, page int, limit int) ([]*domain.Class, int64, error) {
	var classes []*domain.Class
	var total int64

	query := r.db.Model(&domain.Class{}).
		Preload("School").
		Preload("Term.AcademicYear").
		Preload("Creator")

	if schoolID != "" {
		query = query.Where("cls_sch_id = ?", schoolID)
	}
	if termID != "" {
		query = query.Where("cls_trm_id = ?", termID)
	}
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("cls_title ILIKE ? OR cls_code ILIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&classes).Error
	return classes, total, err
}

func (r *classRepository) GetByID(id string) (*domain.Class, error) {
	var class domain.Class
	err := r.db.Preload("School").
		Preload("Term.AcademicYear").
		Preload("Creator").
		Where("cls_id = ?", id).First(&class).Error
	return &class, err
}

func (r *classRepository) Update(class *domain.Class) error {
	result := r.db.Save(class)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *classRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Class{}, "cls_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *classRepository) CountEnrollmentsByClass(classID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).Where("enr_cls_id = ?", classID).Count(&count).Error
	return count, err
}

func (r *classRepository) CountSubjectClassesByClass(classID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.SubjectClass{}).Where("scl_cls_id = ?", classID).Count(&count).Error
	return count, err
}

func (r *classRepository) CheckDuplicateCode(schoolID string, termID string, code string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Class{}).
		Where("cls_sch_id = ? AND cls_trm_id = ? AND cls_code = ?", schoolID, termID, code)
	if excludeID != "" {
		query = query.Where("cls_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *classRepository) GetSchoolIDByClass(classID string) (string, error) {
	var schoolID string
	err := r.db.Model(&domain.Class{}).
		Where("cls_id = ?", classID).
		Pluck("cls_sch_id", &schoolID).Error
	if err != nil {
		return "", err
	}
	if schoolID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return schoolID, nil
}
