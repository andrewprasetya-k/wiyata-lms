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
	Create(actor domain.ActorContext, subject *domain.Subject) error
	FindAll(schoolID string, search string, page int, limit int) ([]*domain.Subject, int64, error)
	GetBySchool(schoolCode string) ([]*domain.Subject, error)
	GetByID(id string) (*domain.Subject, error)
	GetByCode(schoolCode string, subjectCode string) (*domain.Subject, error)
	Update(actor domain.ActorContext, subject *domain.Subject) error
	Delete(actor domain.ActorContext, id string) error
}

type subjectService struct {
	repo          repository.SubjectRepository
	schoolService SchoolService
	logService    LogService
}

func NewSubjectService(repo repository.SubjectRepository, schoolService SchoolService, logService LogService) SubjectService {
	return &subjectService{
		repo:          repo,
		schoolService: schoolService,
		logService:    logService,
	}
}

func (s *subjectService) Create(actor domain.ActorContext, subject *domain.Subject) error {
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

	if err := s.repo.Create(subject); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "subject.created", "subject", strPtr(subject.ID), domain.LogSeverityMedium, map[string]any{
		"subject_name": subject.Name,
		"subject_code": subject.Code,
	})

	return nil
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

func (s *subjectService) Update(actor domain.ActorContext, subject *domain.Subject) error {
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

	if err := s.repo.Update(subject); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "subject.updated", "subject", strPtr(subject.ID), domain.LogSeverityMedium, map[string]any{
		"subject_name": subject.Name,
		"subject_code": subject.Code,
	})

	return nil
}

func (s *subjectService) Delete(actor domain.ActorContext, id string) error {
	subject, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	subjectClassCount, err := s.repo.CountSubjectClassesBySubject(id)
	if err != nil {
		return err
	}
	if subjectClassCount > 0 {
		return fmt.Errorf("tidak dapat menghapus mata pelajaran karena masih memiliki %d penugasan kelas", subjectClassCount)
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "subject.deleted", "subject", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"subject_name": subject.Name,
		"subject_code": subject.Code,
	})

	return nil
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
