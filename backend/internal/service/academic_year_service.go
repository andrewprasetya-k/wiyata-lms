package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type AcademicYearService interface {
	Create(actor domain.ActorContext, acy *domain.AcademicYear) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.AcademicYear, int64, error)
	GetBySchool(schoolCode string) ([]*domain.AcademicYear, error)
	GetByID(id string) (*domain.AcademicYear, error)
	Update(actor domain.ActorContext, acy *domain.AcademicYear) error
	Delete(actor domain.ActorContext, id string) error
	Activate(actor domain.ActorContext, id string) error
	Deactivate(id string) error
}

type academicYearService struct {
	repo          repository.AcademicYearRepository
	schoolService SchoolService
	logService    LogService
}

func NewAcademicYearService(repo repository.AcademicYearRepository, schoolService SchoolService, logService LogService) AcademicYearService {
	return &academicYearService{
		repo:          repo,
		schoolService: schoolService,
		logService:    logService,
	}
}

func (s *academicYearService) Create(actor domain.ActorContext, acy *domain.AcademicYear) error {
	acy.Name = strings.TrimSpace(acy.Name)

	// 1. Validasi Duplikasi Nama di Sekolah yang sama
	exists, err := s.repo.CheckDuplicateName(acy.SchoolID, acy.Name, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tahun ajaran '%s' sudah terdaftar di sekolah ini", acy.Name)
	}

	acy.IsActive = false // Selalu default false saat baru dibuat
	if err := s.repo.Create(acy); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "academic_year.created", "academic_year", strPtr(acy.ID), domain.LogSeverityMedium, map[string]any{
		"year_name": acy.Name,
	})

	return nil
}

func (s *academicYearService) FindAll(schoolID string, search string, page int, limit int) ([]*domain.AcademicYear, int64, error) {
	return s.repo.FindAll(schoolID, search, page, limit)
}

func (s *academicYearService) GetBySchool(schoolCode string) ([]*domain.AcademicYear, error) {
	schoolID, err := s.schoolService.ConvertCodeToID(schoolCode)
	if err != nil {
		return nil, err
	}
	return s.repo.GetBySchool(schoolID)
}

func (s *academicYearService) GetByID(id string) (*domain.AcademicYear, error) {
	return s.repo.GetByID(id)
}

func (s *academicYearService) Update(actor domain.ActorContext, acy *domain.AcademicYear) error {
	acy.Name = strings.TrimSpace(acy.Name)

	// 1. Validasi Duplikasi Nama
	exists, err := s.repo.CheckDuplicateName(acy.SchoolID, acy.Name, acy.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tahun ajaran '%s' sudah terdaftar di sekolah ini", acy.Name)
	}

	if err := s.repo.Update(acy); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "academic_year.updated", "academic_year", strPtr(acy.ID), domain.LogSeverityMedium, map[string]any{
		"year_name": acy.Name,
	})

	return nil
}

func (s *academicYearService) Delete(actor domain.ActorContext, id string) error {
	acy, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Proteksi: Cek apakah masih ada Terms (Semester) yang bergantung
	hasTerms, err := s.repo.HasTerms(id)
	if err != nil {
		return err
	}
	if hasTerms {
		return fmt.Errorf("tahun ajaran tidak bisa dihapus karena masih memiliki data semester (terms)")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "academic_year.deleted", "academic_year", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"year_name": acy.Name,
	})

	return nil
}

func (s *academicYearService) Activate(actor domain.ActorContext, id string) error {
	acy, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	activeBefore := acy.IsActive

	// 1. Aktifkan tahun ajaran ini
	if err := s.repo.SetActiveStatus(id, true); err != nil {
		return err
	}

	// 2. Nonaktifkan yang lainnya di sekolah yang sama
	if err := s.repo.DeactivateAllExcept(acy.SchoolID, id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "academic_year.activated", "academic_year", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"year_name":     acy.Name,
		"active_before": activeBefore,
		"active_after":  true,
	})

	return nil
}

func (s *academicYearService) Deactivate(id string) error {
	return s.repo.SetActiveStatus(id, false)
}
