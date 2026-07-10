package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
)

type SchoolUserService interface {
	Enroll(scu *domain.SchoolUser) error
	GetMembersBySchool(schoolCode string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error)
	GetSchoolsByUser(userID string) ([]*domain.SchoolUser, error)
	Unenroll(id string) error
	BelongsToSchool(schoolUserID string, schoolID string) (bool, error)
}

type schoolUserService struct {
	repo          repository.SchoolUserRepository
	schoolService SchoolService
}

func NewSchoolUserService(repo repository.SchoolUserRepository, schoolService SchoolService) SchoolUserService {
	return &schoolUserService{
		repo:          repo,
		schoolService: schoolService,
	}
}

func (s *schoolUserService) Enroll(scu *domain.SchoolUser) error {
	// 1. Validasi: Apakah sudah terdaftar di sekolah ini?
	already, err := s.repo.IsEnrolled(scu.UserID, scu.SchoolID)
	if err != nil {
		return err
	}
	if already {
		return fmt.Errorf("user sudah terdaftar sebagai anggota di sekolah ini")
	}

	return s.repo.Create(scu)
}

func (s *schoolUserService) GetMembersBySchool(schoolCode string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error) {
	schoolID, err := s.schoolService.ConvertCodeToID(schoolCode)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.GetBySchool(schoolID, search, page, limit)
}

func (s *schoolUserService) GetSchoolsByUser(userID string) ([]*domain.SchoolUser, error) {
	return s.repo.GetByUser(userID)
}

func (s *schoolUserService) Unenroll(userId string) error {
	return s.repo.Delete(userId)
}

func (s *schoolUserService) BelongsToSchool(schoolUserID string, schoolID string) (bool, error) {
	return s.repo.BelongsToSchool(schoolUserID, schoolID)
}
