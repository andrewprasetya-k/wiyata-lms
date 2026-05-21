package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type SubjectClassRepository interface {
	Create(scl *domain.SubjectClass) error
	GetByClass(classID string) ([]*domain.SubjectClass, error)
	GetByID(id string) (*domain.SubjectClass, error)
	Update(scl *domain.SubjectClass) error
	Delete(id string) error
	CheckExists(classID, subjectID, schoolUserID string) (bool, error)
	GetClassIDBySubjectClass(subjectClassID string) (string, error)
	TeacherTeachesInClass(schoolUserID string, classID string) (bool, error)
}

type subjectClassRepository struct {
	db *gorm.DB
}

func NewSubjectClassRepository(db *gorm.DB) SubjectClassRepository {
	return &subjectClassRepository{db: db}
}

func (r *subjectClassRepository) Create(scl *domain.SubjectClass) error {
	return r.db.Create(scl).Error
}

func (r *subjectClassRepository) GetByClass(classID string) ([]*domain.SubjectClass, error) {
	var results []*domain.SubjectClass
	err := r.db.Preload("Subject").Preload("Teacher.User").
		Where("scl_cls_id = ?", classID).Find(&results).Error
	return results, err
}

func (r *subjectClassRepository) GetByID(id string) (*domain.SubjectClass, error) {
	var scl domain.SubjectClass
	err := r.db.Preload("Subject").Preload("Teacher.User").Preload("Class").
		Where("scl_id = ?", id).First(&scl).Error
	return &scl, err
}

func (r *subjectClassRepository) Update(scl *domain.SubjectClass) error {
	result := r.db.Save(scl)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectClassRepository) Delete(id string) error {
	result := r.db.Delete(&domain.SubjectClass{}, "scl_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectClassRepository) CheckExists(classID, subjectID, schoolUserID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_cls_id = ? AND scl_sub_id = ? AND scl_scu_id = ?", classID, subjectID, schoolUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) GetClassIDBySubjectClass(subjectClassID string) (string, error) {
	var classID string
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_id = ?", subjectClassID).
		Pluck("scl_cls_id", &classID).Error
	return classID, err
}

func (r *subjectClassRepository) TeacherTeachesInClass(schoolUserID string, classID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_scu_id = ? AND scl_cls_id = ?", schoolUserID, classID).
		Count(&count).Error
	return count > 0, err
}
