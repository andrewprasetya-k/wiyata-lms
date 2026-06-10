package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"errors"
	"fmt"
)

type EnrollmentService interface {
	Enroll(schoolID string, classID string, schoolUserIDs []string, role string) error
	GetByID(id string) (*domain.Enrollment, error)
	GetByIDInSchool(id string, schoolID string) (*domain.Enrollment, error)
	GetByClass(classID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error)
	GetByClassInSchool(classID string, schoolID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error)
	GetByMember(schoolUserID string) ([]*domain.Enrollment, error)
	GetByMemberInSchool(schoolUserID string, schoolID string) ([]*domain.Enrollment, error)
	Update(id string, schoolID string, role string) error
	Unenroll(id string, schoolID string) error
}

type enrollmentService struct {
	repo           repository.EnrollmentRepository
	classRepo      repository.ClassRepository
	schoolUserRepo repository.SchoolUserRepository
}

func NewEnrollmentService(repo repository.EnrollmentRepository, classRepo repository.ClassRepository, schoolUserRepo repository.SchoolUserRepository) EnrollmentService {
	return &enrollmentService{repo: repo, classRepo: classRepo, schoolUserRepo: schoolUserRepo}
}

func (s *enrollmentService) Enroll(schoolID string, classID string, schoolUserIDs []string, role string) error {
	if err := s.ensureClassInSchool(classID, schoolID); err != nil {
		return err
	}

	for _, scuID := range schoolUserIDs {
		if err := s.ensureSchoolUserInSchool(scuID, schoolID); err != nil {
			return err
		}

		// 1. Validasi: Apakah sudah terdaftar di kelas ini?
		already, err := s.repo.CheckExists(classID, scuID)
		if err != nil {
			return err
		}
		if already {
			continue // Jika sudah ada, lewati user ini
		}

		enr := domain.Enrollment{
			SchoolID:     schoolID,
			ClassID:      classID,
			SchoolUserID: scuID,
			Role:         role,
		}

		if err := s.repo.Create(&enr); err != nil {
			return err
		}
	}
	return nil
}

func (s *enrollmentService) GetByID(id string) (*domain.Enrollment, error) {
	return s.repo.GetByID(id)
}

func (s *enrollmentService) GetByIDInSchool(id string, schoolID string) (*domain.Enrollment, error) {
	if err := s.ensureEnrollmentInSchool(id, schoolID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *enrollmentService) GetByClass(classID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error) {
	return s.repo.GetByClass(classID, search, page, limit)
}

func (s *enrollmentService) GetByClassInSchool(classID string, schoolID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error) {
	if err := s.ensureClassInSchool(classID, schoolID); err != nil {
		return nil, 0, err
	}
	return s.repo.GetByClass(classID, search, page, limit)
}

func (s *enrollmentService) GetByMember(schoolUserID string) ([]*domain.Enrollment, error) {
	return s.repo.GetByMember(schoolUserID)
}

func (s *enrollmentService) GetByMemberInSchool(schoolUserID string, schoolID string) ([]*domain.Enrollment, error) {
	if err := s.ensureSchoolUserInSchool(schoolUserID, schoolID); err != nil {
		return nil, err
	}
	return s.repo.GetByMember(schoolUserID)
}

func (s *enrollmentService) Update(id string, schoolID string, role string) error {
	if err := s.ensureEnrollmentInSchool(id, schoolID); err != nil {
		return err
	}
	return s.repo.Update(id, role)
}

func (s *enrollmentService) Unenroll(id string, schoolID string) error {
	if err := s.ensureEnrollmentInSchool(id, schoolID); err != nil {
		return err
	}

	enrollment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if enrollment.Role == "teacher" {
		hasAssignment, err := s.repo.HasTeacherSubjectClassAssignment(enrollment.ClassID, enrollment.SchoolUserID, schoolID)
		if err != nil {
			return err
		}
		if hasAssignment {
			return fmt.Errorf("teacher subject class assignment exists")
		}
	}

	return s.repo.Delete(id)
}

func (s *enrollmentService) ensureClassInSchool(classID string, schoolID string) error {
	classSchoolID, err := s.classRepo.GetSchoolIDByClass(classID)
	if err != nil {
		return err
	}
	if classSchoolID != schoolID {
		return errors.New("forbidden: class does not belong to active school")
	}
	return nil
}

func (s *enrollmentService) ensureSchoolUserInSchool(schoolUserID string, schoolID string) error {
	ok, err := s.schoolUserRepo.BelongsToSchool(schoolUserID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden: school user does not belong to active school")
	}
	return nil
}

func (s *enrollmentService) ensureEnrollmentInSchool(enrollmentID string, schoolID string) error {
	ok, err := s.repo.BelongsToSchool(enrollmentID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden: enrollment does not belong to active school")
	}
	return nil
}
