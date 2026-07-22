package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
)

type SchoolUserService interface {
	Enroll(actor domain.ActorContext, scu *domain.SchoolUser) error
	GetMembersBySchool(schoolCode string, search string, page int, limit int) ([]*domain.SchoolUser, int64, error)
	GetSchoolsByUser(userID string) ([]*domain.SchoolUser, error)
	Unenroll(actor domain.ActorContext, id string) error
	BelongsToSchool(schoolUserID string, schoolID string) (bool, error)
}

type schoolUserService struct {
	repo          repository.SchoolUserRepository
	schoolService SchoolService
	logService    LogService
}

func NewSchoolUserService(repo repository.SchoolUserRepository, schoolService SchoolService, logService LogService) SchoolUserService {
	return &schoolUserService{
		repo:          repo,
		schoolService: schoolService,
		logService:    logService,
	}
}

func (s *schoolUserService) Enroll(actor domain.ActorContext, scu *domain.SchoolUser) error {
	// 1. Validasi: Apakah sudah terdaftar di sekolah ini?
	already, err := s.repo.IsEnrolled(scu.UserID, scu.SchoolID)
	if err != nil {
		return err
	}
	if already {
		return fmt.Errorf("user sudah terdaftar sebagai anggota di sekolah ini")
	}

	if err := s.repo.Create(scu); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "member.enrolled", "school_user", strPtr(scu.ID), domain.LogSeverityMedium, map[string]any{
		"user_id":   scu.UserID,
		"school_id": scu.SchoolID,
	})

	return nil
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

func (s *schoolUserService) Unenroll(actor domain.ActorContext, userId string) error {
	if err := s.repo.Delete(userId); err != nil {
		return err
	}

	schoolID := ""
	if actor.SchoolID != nil {
		schoolID = *actor.SchoolID
	}
	_ = s.logService.Log(actor, "member.unenrolled", "school_user", strPtr(userId), domain.LogSeverityMedium, map[string]any{
		"user_id":   userId,
		"school_id": schoolID,
	})

	return nil
}

func (s *schoolUserService) BelongsToSchool(schoolUserID string, schoolID string) (bool, error) {
	return s.repo.BelongsToSchool(schoolUserID, schoolID)
}
