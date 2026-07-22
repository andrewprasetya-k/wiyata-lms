package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type ClassService interface {
	Create(actor domain.ActorContext, class *domain.Class) error
	FindAll(search string, schoolCode string, termID string, page int, limit int) ([]*domain.Class, int64, error)
	GetByID(id string) (*domain.Class, error)
	Update(actor domain.ActorContext, class *domain.Class) error
	Delete(actor domain.ActorContext, id string) error
}

type classService struct {
	repo          repository.ClassRepository
	schoolService SchoolService
	logService    LogService
}

func NewClassService(repo repository.ClassRepository, schoolService SchoolService, logService LogService) ClassService {
	return &classService{
		repo:          repo,
		schoolService: schoolService,
		logService:    logService,
	}
}

func (s *classService) Create(actor domain.ActorContext, class *domain.Class) error {
	class.Title = strings.TrimSpace(class.Title)
	class.Code = strings.ToUpper(strings.TrimSpace(class.Code))

	// 1. Validasi Duplikasi Kode di Sekolah & Semester yang sama
	exists, err := s.repo.CheckDuplicateCode(class.SchoolID, class.TermID, class.Code, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode kelas '%s' sudah terdaftar untuk periode ini", class.Code)
	}

	if err := s.repo.Create(class); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "class.created", "class", strPtr(class.ID), domain.LogSeverityMedium, map[string]any{
		"class_name": class.Title,
		"created_by": class.CreatedBy,
	})

	return nil
}

func (s *classService) FindAll(search string, schoolID string, termID string, page int, limit int) ([]*domain.Class, int64, error) {
	return s.repo.FindAll(search, schoolID, termID, page, limit)
}

func (s *classService) GetByID(id string) (*domain.Class, error) {
	return s.repo.GetByID(id)
}

func (s *classService) Update(actor domain.ActorContext, class *domain.Class) error {
	class.Title = strings.TrimSpace(class.Title)
	// Catatan: Kode kelas biasanya tidak diubah setelah dibuat, tapi kita tetap sediakan validasi jika diperlukan
	if err := s.repo.Update(class); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "class.updated", "class", strPtr(class.ID), domain.LogSeverityMedium, map[string]any{
		"class_name": class.Title,
		"created_by": class.CreatedBy,
	})

	return nil
}

func (s *classService) Delete(actor domain.ActorContext, id string) error {
	class, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	enrollmentCount, err := s.repo.CountEnrollmentsByClass(id)
	if err != nil {
		return err
	}
	if enrollmentCount > 0 {
		return fmt.Errorf("class cannot be deleted because it still has enrollments")
	}

	subjectClassCount, err := s.repo.CountSubjectClassesByClass(id)
	if err != nil {
		return err
	}
	if subjectClassCount > 0 {
		return fmt.Errorf("class cannot be deleted because it still has subject assignments")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "class.deleted", "class", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"class_name": class.Title,
		"created_by": class.CreatedBy,
	})

	return nil
}
