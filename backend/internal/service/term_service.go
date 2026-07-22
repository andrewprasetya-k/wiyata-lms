package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type TermService interface {
	Create(actor domain.ActorContext, term *domain.Term) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Term, int64, error)
	GetByAcademicYear(acyID string, schoolID string) ([]*domain.Term, error)
	GetByID(id string) (*domain.Term, error)
	Update(actor domain.ActorContext, term *domain.Term) error
	Delete(actor domain.ActorContext, id string) error
	Activate(actor domain.ActorContext, id string) error
	Deactivate(id string) error
}

type termService struct {
	repo       repository.TermRepository
	logService LogService
}

func NewTermService(repo repository.TermRepository, logService LogService) TermService {
	return &termService{repo: repo, logService: logService}
}

func (s *termService) Create(actor domain.ActorContext, term *domain.Term) error {
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
	if err := s.repo.Create(term); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "term.created", "term", strPtr(term.ID), domain.LogSeverityMedium, map[string]any{
		"term_name":     term.Name,
		"academic_year": term.AcademicYearID,
	})

	return nil
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

func (s *termService) Update(actor domain.ActorContext, term *domain.Term) error {
	term.Name = strings.TrimSpace(term.Name)

	// 1. Validasi Duplikasi Nama
	exists, err := s.repo.CheckDuplicateName(term.AcademicYearID, term.Name, term.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("semester '%s' sudah terdaftar di tahun ajaran ini", term.Name)
	}

	if err := s.repo.Update(term); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "term.updated", "term", strPtr(term.ID), domain.LogSeverityMedium, map[string]any{
		"term_name":     term.Name,
		"academic_year": term.AcademicYearID,
	})

	return nil
}

func (s *termService) Delete(actor domain.ActorContext, id string) error {
	term, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Proteksi: Cek apakah masih ada Kelas yang bergantung
	hasClasses, err := s.repo.HasClasses(id)
	if err != nil {
		return err
	}
	if hasClasses {
		return fmt.Errorf("semester tidak bisa dihapus karena masih memiliki data kelas")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "term.deleted", "term", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"term_name":     term.Name,
		"academic_year": term.AcademicYearID,
	})

	return nil
}

func (s *termService) Activate(actor domain.ActorContext, id string) error {
	term, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Aktifkan semester ini
	if err := s.repo.SetActiveStatus(id, true); err != nil {
		return err
	}

	// 2. Nonaktifkan yang lainnya di tahun ajaran yang sama
	if err := s.repo.DeactivateAllExcept(term.AcademicYearID, id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "term.activated", "term", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"term_name":     term.Name,
		"academic_year": term.AcademicYearID,
	})

	return nil
}

func (s *termService) Deactivate(id string) error {
	return s.repo.SetActiveStatus(id, false)
}
