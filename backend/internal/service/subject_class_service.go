package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
)

type SubjectClassService interface {
	Assign(scl *domain.SubjectClass) error
	GetByClass(classID string) ([]*domain.SubjectClass, error)
	GetTeachingByUserAndSchool(userID string, schoolID string) ([]repository.TeacherSubjectClassRow, error)
	GetByID(id string) (*domain.SubjectClass, error)
	TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error)
	Update(scl *domain.SubjectClass) error
	Unassign(id string) error
}

type subjectClassService struct {
	repo repository.SubjectClassRepository
}

func NewSubjectClassService(repo repository.SubjectClassRepository) SubjectClassService {
	return &subjectClassService{repo: repo}
}

func (s *subjectClassService) Assign(scl *domain.SubjectClass) error {
	// 1. Validasi: Apakah sudah ditugaskan (kombinasi yang sama)?
	already, err := s.repo.CheckExists(scl.ClassID, scl.SubjectID, scl.SchoolUserID)
	if err != nil {
		return err
	}
	if already {
		return fmt.Errorf("this subject is already assigned to the class with the same teacher")
	}

	return s.repo.Create(scl)
}

func (s *subjectClassService) Update(scl *domain.SubjectClass) error {
	// Validasi duplikasi (jika data yang diupdate ternyata sama dengan assignment lain)
	// butuh method CheckExists yang lebih detail jika ingin validasi update,
	// tapi untuk sekarang kita asumsikan update guru saja yang paling sering.
	return s.repo.Update(scl)
}

func (s *subjectClassService) GetByClass(classID string) ([]*domain.SubjectClass, error) {
	return s.repo.GetByClass(classID)
}

func (s *subjectClassService) GetTeachingByUserAndSchool(userID string, schoolID string) ([]repository.TeacherSubjectClassRow, error) {
	return s.repo.GetTeachingByUserAndSchool(userID, schoolID)
}

func (s *subjectClassService) GetByID(id string) (*domain.SubjectClass, error) {
	return s.repo.GetByID(id)
}

func (s *subjectClassService) TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error) {
	return s.repo.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
}

func (s *subjectClassService) Unassign(id string) error {
	return s.repo.Delete(id)
}
