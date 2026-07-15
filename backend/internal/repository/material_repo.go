package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type MaterialRepository interface {
	Create(mat *domain.Material) error
	FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error)
	GetByID(id string) (*domain.Material, error)
	Update(mat *domain.Material) error
	Delete(id string) error

	// Progress
	UpsertProgress(prog *domain.MaterialProgress) error
	GetProgress(userID, matID string) (*domain.MaterialProgress, error)
}

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) Create(mat *domain.Material) error {
	return r.db.Create(mat).Error
}

func (r *materialRepository) FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error) {
	var materials []*domain.Material
	var total int64

	query := r.db.Model(&domain.Material{}).
		Preload("SubjectClass.Subject", func(db *gorm.DB) *gorm.DB {
			return db.Select("sub_id", "sub_name", "sub_color")
		}).
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("usr_id", "usr_nama_lengkap")
		})

	if subjectClassID != "" {
		query = query.Where("mat_scl_id = ?", subjectClassID)
	}
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("mat_title ILIKE ?", searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&materials).Error
	return materials, total, err
}

func (r *materialRepository) GetByID(id string) (*domain.Material, error) {
	var mat domain.Material
	err := r.db.Preload("SubjectClass.Subject", func(db *gorm.DB) *gorm.DB {
		return db.Select("sub_id", "sub_name", "sub_color")
	}).
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("usr_id", "usr_nama_lengkap")
		}).
		Where("mat_id = ?", id).First(&mat).Error
	return &mat, err
}

func (r *materialRepository) Update(mat *domain.Material) error {
	result := r.db.Model(&domain.Material{}).
		Where("mat_id = ?", mat.ID).
		Select("mat_title", "mat_desc", "mat_types", "updated_at").
		Updates(mat)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *materialRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Material{}, "mat_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *materialRepository) UpsertProgress(prog *domain.MaterialProgress) error {
	return r.db.Save(prog).Error
}

func (r *materialRepository) GetProgress(userID, matID string) (*domain.MaterialProgress, error) {
	var prog domain.MaterialProgress
	err := r.db.Where("map_usr_id = ? AND map_mat_id = ?", userID, matID).First(&prog).Error
	return &prog, err
}
