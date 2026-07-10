package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type SubjectRepository interface {
	Create(subject *domain.Subject) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Subject, int64, error)
	GetBySchool(schoolID string) ([]*domain.Subject, error)
	GetByID(id string) (*domain.Subject, error)
	GetByCode(schoolID string, code string) (*domain.Subject, error)
	Update(subject *domain.Subject) error
	Delete(id string) error
	CheckDuplicateCode(schoolID string, code string, excludeID string) (bool, error)
	CountSubjectClassesBySubject(subjectID string) (int64, error)
}

type subjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) SubjectRepository {
	return &subjectRepository{db: db}
}

func (r *subjectRepository) Create(subject *domain.Subject) error {
	return r.db.Create(subject).Error
}

func (r *subjectRepository) FindAll(schoolID string, search string, page int, limit int) ([]*domain.Subject, int64, error) {
	var subjects []*domain.Subject
	var total int64

	query := r.db.Model(&domain.Subject{}).Preload("School").Where("sub_sch_id = ?", schoolID)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("sub_name ILIKE ? OR sub_code ILIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("edv.subjects.created_at desc").Find(&subjects).Error
	return subjects, total, err
}

func (r *subjectRepository) GetBySchool(schoolID string) ([]*domain.Subject, error) {
	var subjects []*domain.Subject
	err := r.db.Preload("School").Where("sub_sch_id = ?", schoolID).Order("created_at asc").Find(&subjects).Error
	return subjects, err
}

func (r *subjectRepository) GetByID(id string) (*domain.Subject, error) {
	var subject domain.Subject
	err := r.db.Preload("School").Where("sub_id = ?", id).First(&subject).Error
	return &subject, err
}

func (r *subjectRepository) GetByCode(schoolID string, code string) (*domain.Subject, error) {
	var subject domain.Subject
	err := r.db.Preload("School").Where("sub_sch_id = ? AND sub_code = ?", schoolID, code).First(&subject).Error
	return &subject, err
}

func (r *subjectRepository) Update(subject *domain.Subject) error {
	result := r.db.Save(subject)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Subject{}, "sub_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectRepository) CheckDuplicateCode(schoolID string, code string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Subject{}).Where("sub_sch_id = ? AND sub_code = ?", schoolID, code)
	if excludeID != "" {
		query = query.Where("sub_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *subjectRepository) CountSubjectClassesBySubject(subjectID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.SubjectClass{}).Where("scl_sub_id = ?", subjectID).Count(&count).Error
	return count, err
}
