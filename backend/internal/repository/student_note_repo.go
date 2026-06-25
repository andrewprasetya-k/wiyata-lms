package repository

import (
	"backend/internal/domain"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentNoteRepository interface {
	GetByUserMaterialInSchool(schoolID string, userID string, materialID string) (*domain.StudentNote, error)
	GetByUserSubjectClassInSchool(schoolID string, userID string, subjectClassID string) ([]domain.StudentNoteWithMaterial, error)
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

func (r *studentNoteRepository) GetByUserSubjectClassInSchool(schoolID string, userID string, subjectClassID string) ([]domain.StudentNoteWithMaterial, error) {
	notes := make([]domain.StudentNoteWithMaterial, 0)
	err := r.db.
		Table("edv.student_notes sn").
		Select(`
			sn.snt_id,
			sn.snt_mat_id,
			m.mat_title AS material_title,
			sn.snt_content,
			sn.created_at,
			sn.updated_at
		`).
		Joins("JOIN edv.materials m ON m.mat_id = sn.snt_mat_id").
		Where("sn.snt_sch_id = ? AND sn.snt_usr_id = ?", schoolID, userID).
		Where("m.mat_scl_id = ? AND m.mat_sch_id = ?", subjectClassID, schoolID).
		Where("m.deleted_at IS NULL").
		Order("sn.updated_at DESC").
		Scan(&notes).Error
	return notes, err
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
