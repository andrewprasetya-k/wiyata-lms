package repository

import (
	"backend/internal/domain"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentNoteRepository interface {
	GetByUserMaterialInSchool(schoolID string, userID string, materialID string) (*domain.StudentNote, error)
	Upsert(note *domain.StudentNote) (*domain.StudentNote, error)
	DeleteByUserMaterialInSchool(schoolID string, userID string, materialID string) error
}

type studentNoteRepository struct {
	db *gorm.DB
}

func NewStudentNoteRepository(db *gorm.DB) StudentNoteRepository {
	return &studentNoteRepository{db: db}
}

func (r *studentNoteRepository) GetByUserMaterialInSchool(schoolID string, userID string, materialID string) (*domain.StudentNote, error) {
	var note domain.StudentNote
	err := r.db.
		Where("snt_sch_id = ? AND snt_usr_id = ? AND snt_mat_id = ?", schoolID, userID, materialID).
		First(&note).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &note, err
}

func (r *studentNoteRepository) Upsert(note *domain.StudentNote) (*domain.StudentNote, error) {
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "snt_usr_id"},
			{Name: "snt_mat_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"snt_sch_id":  note.SchoolID,
			"snt_content": note.Content,
			"updated_at":  gorm.Expr("now()"),
		}),
	}).Create(note).Error
	if err != nil {
		return nil, err
	}

	return r.GetByUserMaterialInSchool(note.SchoolID, note.UserID, note.MaterialID)
}

func (r *studentNoteRepository) DeleteByUserMaterialInSchool(schoolID string, userID string, materialID string) error {
	return r.db.
		Where("snt_sch_id = ? AND snt_usr_id = ? AND snt_mat_id = ?", schoolID, userID, materialID).
		Delete(&domain.StudentNote{}).Error
}
