package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SchoolService interface {
	CreateSchool(school *domain.School, creatorUserID string) (*domain.SchoolUser, error)
	GetSchools(search string, status string, page int, limit int, sortBy string, order string) ([]*domain.School, int64, error)
	GetSchoolByCode(schoolCode string) (*domain.School, error)
	GetSchoolByID(schoolID string) (*domain.School, error)
	RestoreDeletedSchool(schoolCode string) error
	UpdateSchool(school *domain.School) error
	DeleteSchool(schoolCode string) error
	HardDeleteSchool(schoolCode string) error
	GetSchoolSummary() (*dto.SchoolSummaryDTO, error)
	CheckCodeAvailability(schoolCode string) (bool, error)

	//functional methods
	ConvertCodeToID(schoolCode string) (string, error)
}

type schoolService struct {
	repo repository.SchoolRepository
}

// constructor
func NewSchoolService(repo repository.SchoolRepository) SchoolService {
	return &schoolService{repo: repo}
}

func (s *schoolService) CreateSchool(school *domain.School, creatorUserID string) (*domain.SchoolUser, error) {
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
		school.Code = s.generateRandomCode()
	} else {
		_, err := s.repo.GetSchoolByCode(school.Code)
		if err == nil {
			return nil, fmt.Errorf("school code '%s' already exists", school.Code)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return s.repo.CreateSchoolWithAdmin(school, creatorUserID)
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

func (s *schoolService) generateRandomCode() string {
	word := []rune("ABCDEFGHJKMNPQRSTUVWXYZ23456789")
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	for range 10 { // Coba maksimal 10 kali
		code := make([]rune, 6)
		for j := range code {
			code[j] = word[seededRand.Intn(len(word))]
		}

		// Cek keunikan
		_, err := s.repo.GetSchoolByCode(string(code))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return string(code)
		}
	}
	return "" // Atau handle error jika gagal dapet kode unik
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

func (s *schoolService) UpdateSchool(school *domain.School) error {
	s.sanitizeInput(school)

	// 1. Validasi Duplikasi Email (kecuali milik sekolah ini sendiri)
	exists, err := s.repo.CheckEmailExists(school.Email, school.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email sekolah '%s' sudah terdaftar", school.Email)
	}

	// 2. Validasi Duplikasi Telepon (kecuali milik sekolah ini sendiri)
	exists, err = s.repo.CheckPhoneExists(school.Phone, school.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("nomor telepon sekolah '%s' sudah terdaftar", school.Phone)
	}

	existing, err := s.repo.GetSchoolByCode(school.Code)

	// Kalau kodenya ketemu, cek apakah itu milik sekolah LAIN?
	if err == nil && existing != nil {
		// Jika ID yang di DB beda dengan ID yang mau kita update, berarti DUPLIKAT
		if existing.ID != school.ID {
			return fmt.Errorf("kode sekolah '%s' sudah dipakai oleh sekolah lain", school.Code)
		}
	}

	return s.repo.UpdateSchool(school)
}

func (s *schoolService) RestoreDeletedSchool(schoolCode string) error {
	schoolID, err := s.ConvertCodeToID(schoolCode)
	if err != nil {
		return err
	}
	return s.repo.RestoreDeletedSchool(schoolID)
}

func (s *schoolService) DeleteSchool(schoolCode string) error {
	schoolID, err := s.ConvertCodeToID(schoolCode)
	if err != nil {
		return err
	}
	return s.repo.DeleteSchool(schoolID)
}

func (s *schoolService) HardDeleteSchool(schoolCode string) error {
	schoolID, err := s.ConvertCodeToID(schoolCode)
	if err != nil {
		return err
	}
	return s.repo.HardDeleteSchool(schoolID)
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
