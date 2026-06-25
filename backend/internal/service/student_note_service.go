package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"errors"
	"fmt"
	"strings"
)

const studentNoteMaxLength = 10000

var (
	ErrStudentNoteContentRequired = errors.New("student note content is required")
	ErrStudentNoteContentTooLong  = errors.New("student note content exceeds 10000 characters")
)

type StudentNoteService interface {
	GetMaterialNote(materialID string, schoolID string, userID string) (*domain.StudentNote, error)
	GetSubjectClassNotes(subjectClassID string, schoolID string, userID string) ([]domain.StudentNoteWithMaterial, error)
	SaveMaterialNote(materialID string, schoolID string, userID string, content string) (*domain.StudentNote, error)
	DeleteMaterialNote(materialID string, schoolID string, userID string) error
}

type studentNoteMaterialRepository interface {
	GetByID(id string) (*domain.Material, error)
}

type studentNoteSubjectClassRepository interface {
	UserEnrolledInSubjectClassAsRole(userID string, schoolID string, subjectClassID string, role string) (bool, error)
}

type studentNoteService struct {
	repo             repository.StudentNoteRepository
	materialRepo     studentNoteMaterialRepository
	subjectClassRepo studentNoteSubjectClassRepository
}

func NewStudentNoteService(repo repository.StudentNoteRepository, materialRepo studentNoteMaterialRepository, subjectClassRepo studentNoteSubjectClassRepository) StudentNoteService {
	return &studentNoteService{
		repo:             repo,
		materialRepo:     materialRepo,
		subjectClassRepo: subjectClassRepo,
	}
}

func (s *studentNoteService) GetMaterialNote(materialID string, schoolID string, userID string) (*domain.StudentNote, error) {
	if err := s.ensureStudentCanAccessMaterial(materialID, schoolID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByUserMaterialInSchool(schoolID, userID, materialID)
}

func (s *studentNoteService) GetSubjectClassNotes(subjectClassID string, schoolID string, userID string) ([]domain.StudentNoteWithMaterial, error) {
	allowed, err := s.subjectClassRepo.UserEnrolledInSubjectClassAsRole(userID, schoolID, subjectClassID, "student")
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: student cannot access subject class notes")
	}

	return s.repo.GetByUserSubjectClassInSchool(schoolID, userID, subjectClassID)
}

func (s *studentNoteService) SaveMaterialNote(materialID string, schoolID string, userID string, content string) (*domain.StudentNote, error) {
	if err := s.ensureStudentCanAccessMaterial(materialID, schoolID, userID); err != nil {
		return nil, err
	}

	normalized, err := normalizeStudentNoteContent(content)
	if err != nil {
		return nil, err
	}

	return s.repo.Upsert(&domain.StudentNote{
		SchoolID:   schoolID,
		UserID:     userID,
		MaterialID: materialID,
		Content:    normalized,
	})
}

func (s *studentNoteService) DeleteMaterialNote(materialID string, schoolID string, userID string) error {
	if err := s.ensureStudentCanAccessMaterial(materialID, schoolID, userID); err != nil {
		return err
	}
	return s.repo.DeleteByUserMaterialInSchool(schoolID, userID, materialID)
}

func (s *studentNoteService) ensureStudentCanAccessMaterial(materialID string, schoolID string, userID string) error {
	material, err := s.materialRepo.GetByID(materialID)
	if err != nil {
		return err
	}
	if material.SchoolID != schoolID {
		return fmt.Errorf("forbidden: material note does not belong to active school")
	}

	allowed, err := s.subjectClassRepo.UserEnrolledInSubjectClassAsRole(userID, schoolID, material.SubjectClassID, "student")
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden: student cannot access material note")
	}
	return nil
}

func normalizeStudentNoteContent(content string) (string, error) {
	normalized := strings.TrimSpace(content)
	if normalized == "" {
		return "", ErrStudentNoteContentRequired
	}
	if len([]rune(normalized)) > studentNoteMaxLength {
		return "", ErrStudentNoteContentTooLong
	}
	return normalized, nil
}
