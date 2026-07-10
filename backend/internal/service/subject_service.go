package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"regexp"
	"strings"
)

var subjectColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

type SubjectService interface {
	Create(subject *domain.Subject) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Subject, int64, error)
	GetBySchool(schoolCode string) ([]*domain.Subject, error)
	GetByID(id string) (*domain.Subject, error)
	GetByCode(schoolCode string, subjectCode string) (*domain.Subject, error)
	Update(subject *domain.Subject) error
	Delete(id string) error
}

type subjectService struct {
	repo          repository.SubjectRepository
	schoolService SchoolService
}

func NewSubjectService(repo repository.SubjectRepository, schoolService SchoolService) SubjectService {
	return &subjectService{
		repo:          repo,
		schoolService: schoolService,
	}
}

func (s *subjectService) Create(subject *domain.Subject) error {
	subject.Name = strings.TrimSpace(subject.Name)
	subject.Code = strings.ToUpper(strings.TrimSpace(subject.Code))
	if err := normalizeSubjectColor(&subject.Color); err != nil {
		return err
	}

	// 1. Validasi Duplikasi Kode di Sekolah yang sama
	exists, err := s.repo.CheckDuplicateCode(subject.SchoolID, subject.Code, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode mata pelajaran '%s' sudah terdaftar di sekolah ini", subject.Code)
	}

	return s.repo.Create(subject)
}

func (s *subjectService) FindAll(schoolID string, search string, page int, limit int) ([]*domain.Subject, int64, error) {
	return s.repo.FindAll(schoolID, search, page, limit)
}

func (s *subjectService) GetBySchool(schoolCode string) ([]*domain.Subject, error) {
	schoolID, err := s.schoolService.ConvertCodeToID(schoolCode)
	if err != nil {
		return nil, err
	}
	return s.repo.GetBySchool(schoolID)
}

func (s *subjectService) GetByID(id string) (*domain.Subject, error) {
	return s.repo.GetByID(id)
}

func (s *subjectService) GetByCode(schoolCode string, subjectCode string) (*domain.Subject, error) {
	schoolID, err := s.schoolService.ConvertCodeToID(schoolCode)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByCode(schoolID, strings.ToUpper(subjectCode))
}

func (s *subjectService) Update(subject *domain.Subject) error {
	subject.Name = strings.TrimSpace(subject.Name)
	subject.Code = strings.ToUpper(strings.TrimSpace(subject.Code))
	if err := normalizeSubjectColor(&subject.Color); err != nil {
		return err
	}

	// 1. Validasi Duplikasi Kode
	exists, err := s.repo.CheckDuplicateCode(subject.SchoolID, subject.Code, subject.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode mata pelajaran '%s' sudah terdaftar di sekolah ini", subject.Code)
	}

	return s.repo.Update(subject)
}

func (s *subjectService) Delete(id string) error {
	subjectClassCount, err := s.repo.CountSubjectClassesBySubject(id)
	if err != nil {
		return err
	}
	if subjectClassCount > 0 {
		return fmt.Errorf("tidak dapat menghapus mata pelajaran karena masih memiliki %d penugasan kelas", subjectClassCount)
	}

	return s.repo.Delete(id)
}

func normalizeSubjectColor(color *string) error {
	trimmed := strings.TrimSpace(*color)
	if trimmed == "" {
		*color = ""
		return nil
	}
	if !subjectColorPattern.MatchString(trimmed) {
		return fmt.Errorf("invalid subject color format")
	}
	*color = strings.ToUpper(trimmed)
	return nil
}
