package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type ClassService interface {
	Create(class *domain.Class) error
	FindAll(search string, schoolCode string, termID string, page int, limit int) ([]*domain.Class, int64, error)
	GetByID(id string) (*domain.Class, error)
	Update(class *domain.Class) error
	Delete(id string) error
}

type classService struct {
	repo          repository.ClassRepository
	schoolService SchoolService
}

func NewClassService(repo repository.ClassRepository, schoolService SchoolService) ClassService {
	return &classService{
		repo:          repo,
		schoolService: schoolService,
	}
}

func (s *classService) Create(class *domain.Class) error {
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

	return s.repo.Create(class)
}

func (s *classService) FindAll(search string, schoolID string, termID string, page int, limit int) ([]*domain.Class, int64, error) {
	return s.repo.FindAll(search, schoolID, termID, page, limit)
}

func (s *classService) GetByID(id string) (*domain.Class, error) {
	return s.repo.GetByID(id)
}

func (s *classService) Update(class *domain.Class) error {
	class.Title = strings.TrimSpace(class.Title)
	// Catatan: Kode kelas biasanya tidak diubah setelah dibuat, tapi kita tetap sediakan validasi jika diperlukan
	return s.repo.Update(class)
}

func (s *classService) Delete(id string) error {
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

	return s.repo.Delete(id)
}
