package repository

import (
	"backend/internal/domain"
	"backend/internal/utils"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentNoteRepository interface {
	GetByUserMaterialInSchool(schoolID string, userID string, materialID string) (*domain.StudentNote, error)
	GetByUserSubjectClassInSchool(schoolID string, userID string, subjectClassID string) ([]domain.StudentNoteWithMaterial, error)
	GetAccessibleByUserInSchool(schoolID string, userID string) ([]domain.StudentGlobalNote, error)
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

func (r *studentNoteRepository) GetAccessibleByUserInSchool(schoolID string, userID string) ([]domain.StudentGlobalNote, error) {
	notes := make([]domain.StudentGlobalNote, 0)
	err := r.db.
		Table("edv.student_notes sn").
		Select(`
			sn.snt_id,
			sn.snt_mat_id,
			m.mat_title AS material_title,
			m.mat_types AS material_type,
			sc.scl_id AS subject_class_id,
			sub.sub_id AS subject_id,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			sn.snt_content,
			sn.created_at,
			sn.updated_at
		`).
		Joins("JOIN edv.materials m ON m.mat_id = sn.snt_mat_id").
		Joins("JOIN edv.subject_classes sc ON sc.scl_id = m.mat_scl_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id").
		Joins("JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL").
		Where("sn.snt_sch_id = ? AND sn.snt_usr_id = ?", schoolID, userID).
		Where("m.mat_sch_id = ? AND m.deleted_at IS NULL", schoolID).
		Where("sub.sub_sch_id = ?", schoolID).
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Where("e.enr_sch_id = ? AND e.enr_role = ? AND e.left_at IS NULL", schoolID, "student").
		Where("scu.scu_usr_id = ? AND scu.scu_sch_id = ? AND scu.deleted_at IS NULL", userID, schoolID).
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
			"updated_at":  utils.NowJakarta(),
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
