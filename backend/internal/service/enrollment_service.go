package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type EnrollmentService interface {
	Enroll(actor domain.ActorContext, schoolID string, classID string, schoolUserIDs []string, role string) error
	GetByID(id string) (*domain.Enrollment, error)
	GetByIDInSchool(id string, schoolID string) (*domain.Enrollment, error)
	GetByClass(classID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error)
	GetByClassInSchool(classID string, schoolID string, search string, page int, limit int) ([]*domain.Enrollment, int64, error)
	GetByMember(schoolUserID string) ([]*domain.Enrollment, error)
	GetByMemberInSchool(schoolUserID string, schoolID string) ([]*domain.Enrollment, error)
	Update(actor domain.ActorContext, id string, schoolID string, role string) error
	Unenroll(actor domain.ActorContext, id string, schoolID string) error
}

type enrollmentService struct {
	repo           repository.EnrollmentRepository
	classRepo      repository.ClassRepository
	schoolUserRepo repository.SchoolUserRepository
	logService     LogService
}

func NewEnrollmentService(repo repository.EnrollmentRepository, classRepo repository.ClassRepository, schoolUserRepo repository.SchoolUserRepository, logService LogService) EnrollmentService {
	return &enrollmentService{repo: repo, classRepo: classRepo, schoolUserRepo: schoolUserRepo, logService: logService}
}

func (s *enrollmentService) Enroll(actor domain.ActorContext, schoolID string, classID string, schoolUserIDs []string, role string) error {
	if err := s.ensureClassInSchool(classID, schoolID); err != nil {
		return err
	}
	class, err := s.classRepo.GetByID(classID)
	if err != nil {
		return err
	}

	for _, scuID := range schoolUserIDs {
		if err := s.ensureSchoolUserInSchool(scuID, schoolID); err != nil {
			return err
		}

		existing, err := s.repo.GetByClassAndSchoolUser(classID, scuID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if existing.LeftAt == nil {
				continue // already actively enrolled
			}
			beforeRole := existing.Role
			if err := s.repo.Reactivate(existing.ID, role); err != nil {
				return err
			}
			_ = s.logService.Log(actor, "enrollment.created", "enrollment", strPtr(existing.ID), domain.LogSeverityMedium, map[string]any{
				"class_code":  class.Code,
				"before_role": beforeRole,
				"after_role":  role,
			})
			continue
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
		_ = s.logService.Log(actor, "enrollment.created", "enrollment", strPtr(enr.ID), domain.LogSeverityMedium, map[string]any{
			"class_code": class.Code,
			"after_role": role,
		})
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

func (s *enrollmentService) Update(actor domain.ActorContext, id string, schoolID string, role string) error {
	if err := s.ensureActiveEnrollmentInSchool(id, schoolID); err != nil {
		return err
	}
	enrollment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	beforeRole := enrollment.Role

	if err := s.repo.Update(id, role); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "enrollment.updated", "enrollment", strPtr(id), domain.LogSeverityMedium, map[string]any{
		"class_code":  enrollment.Class.Code,
		"before_role": beforeRole,
		"after_role":  role,
	})
	return nil
}

func (s *enrollmentService) Unenroll(actor domain.ActorContext, id string, schoolID string) error {
	if err := s.ensureEnrollmentInSchool(id, schoolID); err != nil {
		return err
	}

	enrollment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if enrollment.LeftAt != nil {
		return nil
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

	if err := s.repo.SoftDelete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "enrollment.removed", "enrollment", strPtr(id), domain.LogSeverityMedium, map[string]any{
		"class_code":  enrollment.Class.Code,
		"before_role": enrollment.Role,
	})
	return nil
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

func (s *enrollmentService) ensureActiveEnrollmentInSchool(enrollmentID string, schoolID string) error {
	ok, err := s.repo.ActiveBelongsToSchool(enrollmentID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden: enrollment does not belong to active school")
	}
	return nil
}
