package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SchoolService interface {
	CreateSchool(actor domain.ActorContext, school *domain.School, creatorUserID string) (*domain.SchoolUser, error)
	GetSchools(search string, status string, page int, limit int, sortBy string, order string) ([]*domain.School, int64, error)
	GetSchoolByCode(schoolCode string) (*domain.School, error)
	GetSchoolByID(schoolID string) (*domain.School, error)
	RestoreDeletedSchool(actor domain.ActorContext, schoolCode string) error
	UpdateSchool(actor domain.ActorContext, school *domain.School) error
	DeleteSchool(actor domain.ActorContext, schoolCode string) error
	HardDeleteSchool(actor domain.ActorContext, schoolCode string) error
	GetSchoolSummary() (*dto.SchoolSummaryDTO, error)
	CheckCodeAvailability(schoolCode string) (bool, error)

	//functional methods
	ConvertCodeToID(schoolCode string) (string, error)
}

type schoolService struct {
	repo       repository.SchoolRepository
	logService LogService
}

// constructor
func NewSchoolService(repo repository.SchoolRepository, logService LogService) SchoolService {
	return &schoolService{repo: repo, logService: logService}
}

// diffSchoolFields returns the domain.School field names (snake_case, to
// match log_dto convention elsewhere) that differ between before and after —
// the {changed_fields} metadata shape for school.updated (Phase 10.2).
func diffSchoolFields(before *domain.School, after *domain.School) []string {
	changed := []string{}
	if before.Name != after.Name {
		changed = append(changed, "name")
	}
	if before.Code != after.Code {
		changed = append(changed, "code")
	}
	if !stringPtrEqual(before.LogoID, after.LogoID) {
		changed = append(changed, "logo_id")
	}
	if before.Address != after.Address {
		changed = append(changed, "address")
	}
	if before.Email != after.Email {
		changed = append(changed, "email")
	}
	if before.Phone != after.Phone {
		changed = append(changed, "phone")
	}
	if !stringPtrEqual(before.Website, after.Website) {
		changed = append(changed, "website")
	}
	return changed
}

func stringPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *schoolService) CreateSchool(actor domain.ActorContext, school *domain.School, creatorUserID string) (*domain.SchoolUser, error) {
	if strings.TrimSpace(creatorUserID) == "" {
		return nil, fmt.Errorf("creator user is required")
	}

	s.sanitizeInput(school)

	// 1. Validasi duplikasi email — dilewati jika kosong (opsional untuk self-service).
	if school.Email != "" {
		exists, err := s.repo.CheckEmailExists(school.Email, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("email sekolah '%s' sudah terdaftar", school.Email)
		}
	}

	// 2. Validasi duplikasi telepon — dilewati jika kosong (opsional untuk self-service).
	if school.Phone != "" {
		exists, err := s.repo.CheckPhoneExists(school.Phone, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("nomor telepon sekolah '%s' sudah terdaftar", school.Phone)
		}
	}

	// 3. Jika code kosong, generate otomatis dengan pengecekan keunikan.
	// Nama sekolah sengaja tidak divalidasi unik (boleh sama).
	if school.Code == "" {
		code, err := s.repo.GenerateUniqueCode()
		if err != nil {
			return nil, err
		}
		school.Code = code
	} else {
		_, err := s.repo.GetSchoolByCode(school.Code)
		if err == nil {
			return nil, fmt.Errorf("school code '%s' already exists", school.Code)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	schoolUser, err := s.repo.CreateSchoolWithAdmin(school, creatorUserID)
	if err != nil {
		return nil, err
	}

	// The actor had no active school context yet (the school didn't exist
	// until this call), but the resulting event belongs to the school just
	// created — so scope/school_id are set from the result, not the request.
	schoolActor := actor
	schoolActor.Scope = domain.LogScopeSchool
	schoolActor.SchoolID = strPtr(school.ID)
	_ = s.logService.Log(schoolActor, "school.created", "school", strPtr(school.ID), domain.LogSeverityMedium, map[string]any{
		"school_code": school.Code,
		"school_name": school.Name,
	})

	return schoolUser, nil
}

func (s *schoolService) sanitizeInput(school *domain.School) {
	school.Name = strings.TrimSpace(school.Name)
	school.Address = strings.TrimSpace(school.Address)
	school.Email = strings.TrimSpace(school.Email)
	school.Phone = strings.TrimSpace(school.Phone)
	school.Code = strings.TrimSpace(school.Code)

	if school.Website != nil {
		trimmed := strings.TrimSpace(*school.Website)
		school.Website = &trimmed
	}
}

func (s *schoolService) GetSchools(search string, status string, page int, limit int, sortBy string, order string) ([]*domain.School, int64, error) {
	return s.repo.GetSchools(search, status, page, limit, sortBy, order)
}

func (s *schoolService) GetSchoolByCode(schoolCode string) (*domain.School, error) {
	return s.repo.GetSchoolByCode(schoolCode)
}

func (s *schoolService) GetSchoolByID(schoolID string) (*domain.School, error) {
	return s.repo.GetSchoolByID(schoolID)
}

func (s *schoolService) UpdateSchool(actor domain.ActorContext, school *domain.School) error {
	s.sanitizeInput(school)

	existing, err := s.repo.GetSchoolByID(school.ID)
	if err != nil {
		return err
	}

	// 1. Validasi Duplikasi Email (kecuali milik sekolah ini sendiri)
	emailExists, err := s.repo.CheckEmailExists(school.Email, school.ID)
	if err != nil {
		return err
	}
	if emailExists {
		return fmt.Errorf("email sekolah '%s' sudah terdaftar", school.Email)
	}

	// 2. Validasi Duplikasi Telepon (kecuali milik sekolah ini sendiri)
	phoneExists, err := s.repo.CheckPhoneExists(school.Phone, school.ID)
	if err != nil {
		return err
	}
	if phoneExists {
		return fmt.Errorf("nomor telepon sekolah '%s' sudah terdaftar", school.Phone)
	}

	byCode, err := s.repo.GetSchoolByCode(school.Code)

	// Kalau kodenya ketemu, cek apakah itu milik sekolah LAIN?
	if err == nil && byCode != nil {
		// Jika ID yang di DB beda dengan ID yang mau kita update, berarti DUPLIKAT
		if byCode.ID != school.ID {
			return fmt.Errorf("kode sekolah '%s' sudah dipakai oleh sekolah lain", school.Code)
		}
	}

	if err := s.repo.UpdateSchool(school); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "school.updated", "school", strPtr(school.ID), domain.LogSeverityMedium, map[string]any{
		"changed_fields": diffSchoolFields(existing, school),
	})
	return nil
}

func (s *schoolService) RestoreDeletedSchool(actor domain.ActorContext, schoolCode string) error {
	school, err := s.repo.GetSchoolByCode(schoolCode)
	if err != nil {
		return err
	}
	if err := s.repo.RestoreDeletedSchool(school.ID); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "school.restored", "school", strPtr(school.ID), domain.LogSeverityMedium, map[string]any{
		"school_code": school.Code,
		"school_name": school.Name,
	})
	return nil
}

func (s *schoolService) DeleteSchool(actor domain.ActorContext, schoolCode string) error {
	school, err := s.repo.GetSchoolByCode(schoolCode)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteSchool(school.ID); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "school.deleted", "school", strPtr(school.ID), domain.LogSeverityMedium, map[string]any{
		"school_code": school.Code,
		"school_name": school.Name,
	})
	return nil
}

func (s *schoolService) HardDeleteSchool(actor domain.ActorContext, schoolCode string) error {
	school, err := s.repo.GetSchoolByCode(schoolCode)
	if err != nil {
		return err
	}
	if err := s.repo.HardDeleteSchool(school.ID); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "school.hard_deleted", "school", strPtr(school.ID), domain.LogSeverityHigh, map[string]any{
		"school_code": school.Code,
		"school_name": school.Name,
	})
	return nil
}

func (s *schoolService) GetSchoolSummary() (*dto.SchoolSummaryDTO, error) {
	active, deleted, total, err := s.repo.GetSchoolSummary()
	if err != nil {
		return nil, err
	}
	return &dto.SchoolSummaryDTO{
		TotalActive:  active,
		TotalDeleted: deleted,
		TotalSchools: total,
	}, nil
}

func (s *schoolService) CheckCodeAvailability(schoolCode string) (bool, error) {
	_, err := s.repo.GetSchoolByCode(strings.TrimSpace(schoolCode))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil // Tersedia
		}
		return false, err
	}
	return false, nil // Sudah ada (tidak tersedia)
}

// functional methods
func (s *schoolService) ConvertCodeToID(schoolCode string) (string, error) {
	school, err := s.repo.GetSchoolByCode(schoolCode)
	if err != nil {
		return "", err
	}
	return school.ID, nil
}
