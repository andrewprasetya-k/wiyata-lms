package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type TermService interface {
	Create(term *domain.Term) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Term, int64, error)
	GetByAcademicYear(acyID string, schoolID string) ([]*domain.Term, error)
	GetByID(id string) (*domain.Term, error)
	Update(term *domain.Term) error
	Delete(id string) error
	Activate(id string) error
	Deactivate(id string) error
}

type termService struct {
	repo repository.TermRepository
}

func NewTermService(repo repository.TermRepository) TermService {
	return &termService{repo: repo}
}

func (s *termService) Create(term *domain.Term) error {
	term.Name = strings.TrimSpace(term.Name)

	// 1. Validasi Duplikasi Nama di Tahun Ajaran yang sama
	exists, err := s.repo.CheckDuplicateName(term.AcademicYearID, term.Name, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("semester '%s' sudah terdaftar di tahun ajaran ini", term.Name)
	}

	term.IsActive = false // Default draft
	return s.repo.Create(term)
}

func (s *termService) FindAll(schoolID string, search string, page int, limit int) ([]*domain.Term, int64, error) {
	return s.repo.FindAll(schoolID, search, page, limit)
}

func (s *termService) GetByAcademicYear(acyID string, schoolID string) ([]*domain.Term, error) {
	return s.repo.GetByAcademicYear(acyID, schoolID)
}

func (s *termService) GetByID(id string) (*domain.Term, error) {
	return s.repo.GetByID(id)
}

func (s *termService) Update(term *domain.Term) error {
	term.Name = strings.TrimSpace(term.Name)

	// 1. Validasi Duplikasi Nama
	exists, err := s.repo.CheckDuplicateName(term.AcademicYearID, term.Name, term.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("semester '%s' sudah terdaftar di tahun ajaran ini", term.Name)
	}

	return s.repo.Update(term)
}

func (s *termService) Delete(id string) error {
	// 1. Proteksi: Cek apakah masih ada Kelas yang bergantung
	hasClasses, err := s.repo.HasClasses(id)
	if err != nil {
		return err
	}
	if hasClasses {
		return fmt.Errorf("semester tidak bisa dihapus karena masih memiliki data kelas")
	}

	return s.repo.Delete(id)
}

func (s *termService) Activate(id string) error {
	term, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Aktifkan semester ini
	if err := s.repo.SetActiveStatus(id, true); err != nil {
		return err
	}

	// 2. Nonaktifkan yang lainnya di tahun ajaran yang sama
	return s.repo.DeactivateAllExcept(term.AcademicYearID, id)
}

func (s *termService) Deactivate(id string) error {
	return s.repo.SetActiveStatus(id, false)
}
