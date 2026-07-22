package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
)

type SubjectClassService interface {
	Assign(scl *domain.SubjectClass) error
	AssignInSchool(actor domain.ActorContext, scl *domain.SubjectClass, schoolID string) error
	GetByClass(classID string) ([]*domain.SubjectClass, error)
	GetByClassInSchool(classID string, schoolID string) ([]*domain.SubjectClass, error)
	GetTeachingByUserAndSchool(userID string, schoolID string) ([]repository.TeacherSubjectClassRow, error)
	GetByID(id string) (*domain.SubjectClass, error)
	GetByIDInSchool(id string, schoolID string) (*domain.SubjectClass, error)
	TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error)
	TeacherOwnsClassSubject(userID string, schoolID string, classID string, subjectID string) (bool, error)
	UserCanAccessSubjectClass(userID string, schoolID string, subjectClassID string, roles []string) (bool, error)
	Update(scl *domain.SubjectClass) error
	UpdateInSchool(actor domain.ActorContext, scl *domain.SubjectClass, schoolID string) error
	Unassign(id string) error
	UnassignInSchool(actor domain.ActorContext, id string, schoolID string) error
}

type subjectClassService struct {
	repo       repository.SubjectClassRepository
	logService LogService
}

func NewSubjectClassService(repo repository.SubjectClassRepository, logService LogService) SubjectClassService {
	return &subjectClassService{repo: repo, logService: logService}
}

// subjectClassAuditMetadata builds the {class_name, teacher_name, subject_name}
// metadata shape (Phase 10.2) from a fully preloaded SubjectClass — used by
// AssignInSchool/UpdateInSchool/UnassignInSchool alike.
func subjectClassAuditMetadata(scl *domain.SubjectClass) map[string]any {
	return map[string]any{
		"class_name":   scl.Class.Title,
		"teacher_name": scl.Teacher.User.FullName,
		"subject_name": scl.Subject.Name,
	}
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

func (s *subjectClassService) AssignInSchool(actor domain.ActorContext, scl *domain.SubjectClass, schoolID string) error {
	if err := s.validateAssignmentScope(scl.ClassID, scl.SubjectID, scl.SchoolUserID, schoolID); err != nil {
		return err
	}

	if err := s.ensureNoClassSubjectDuplicate(scl.ClassID, scl.SubjectID, ""); err != nil {
		return err
	}

	if err := s.repo.Create(scl); err != nil {
		return err
	}

	if full, err := s.repo.GetByID(scl.ID); err == nil {
		_ = s.logService.Log(actor, "subject_class.assigned", "subject_class", strPtr(scl.ID), domain.LogSeverityMedium, subjectClassAuditMetadata(full))
	}
	return nil
}

func (s *subjectClassService) Update(scl *domain.SubjectClass) error {
	// Validasi duplikasi (jika data yang diupdate ternyata sama dengan assignment lain)
	// butuh method CheckExists yang lebih detail jika ingin validasi update,
	// tapi untuk sekarang kita asumsikan update guru saja yang paling sering.
	return s.repo.Update(scl)
}

func (s *subjectClassService) UpdateInSchool(actor domain.ActorContext, scl *domain.SubjectClass, schoolID string) error {
	if err := s.ensureSubjectClassInSchool(scl.ID, schoolID); err != nil {
		return err
	}
	if err := s.validateAssignmentScope(scl.ClassID, scl.SubjectID, scl.SchoolUserID, schoolID); err != nil {
		return err
	}
	if err := s.ensureNoClassSubjectDuplicate(scl.ClassID, scl.SubjectID, scl.ID); err != nil {
		return err
	}
	if err := s.repo.Update(scl); err != nil {
		return err
	}

	if full, err := s.repo.GetByID(scl.ID); err == nil {
		_ = s.logService.Log(actor, "subject_class.reassigned", "subject_class", strPtr(scl.ID), domain.LogSeverityMedium, subjectClassAuditMetadata(full))
	}
	return nil
}

func (s *subjectClassService) GetByClass(classID string) ([]*domain.SubjectClass, error) {
	return s.repo.GetByClass(classID)
}

func (s *subjectClassService) GetByClassInSchool(classID string, schoolID string) ([]*domain.SubjectClass, error) {
	if err := s.ensureClassInSchool(classID, schoolID); err != nil {
		return nil, err
	}
	return s.repo.GetByClass(classID)
}

func (s *subjectClassService) GetTeachingByUserAndSchool(userID string, schoolID string) ([]repository.TeacherSubjectClassRow, error) {
	return s.repo.GetTeachingByUserAndSchool(userID, schoolID)
}

func (s *subjectClassService) GetByID(id string) (*domain.SubjectClass, error) {
	return s.repo.GetByID(id)
}

func (s *subjectClassService) GetByIDInSchool(id string, schoolID string) (*domain.SubjectClass, error) {
	if err := s.ensureSubjectClassInSchool(id, schoolID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *subjectClassService) TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error) {
	return s.repo.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
}

func (s *subjectClassService) TeacherOwnsClassSubject(userID string, schoolID string, classID string, subjectID string) (bool, error) {
	return s.repo.TeacherOwnsClassSubject(userID, schoolID, classID, subjectID)
}

func (s *subjectClassService) UserCanAccessSubjectClass(userID string, schoolID string, subjectClassID string, roles []string) (bool, error) {
	for _, role := range roles {
		switch role {
		case "admin":
			return s.repo.SubjectClassBelongsToSchool(subjectClassID, schoolID)
		case "teacher":
			ok, err := s.repo.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
			if err != nil || ok {
				return ok, err
			}
		case "student":
			ok, err := s.repo.UserEnrolledInSubjectClassAsRole(userID, schoolID, subjectClassID, "student")
			if err != nil || ok {
				return ok, err
			}
		}
	}
	return false, nil
}

func (s *subjectClassService) Unassign(id string) error {
	return s.repo.Delete(id)
}

func (s *subjectClassService) UnassignInSchool(actor domain.ActorContext, id string, schoolID string) error {
	if err := s.ensureSubjectClassInSchool(id, schoolID); err != nil {
		return err
	}
	hasContent, err := s.repo.HasSubjectClassContent(id, schoolID)
	if err != nil {
		return err
	}
	if hasContent {
		return fmt.Errorf("subject class has content")
	}

	full, fullErr := s.repo.GetByID(id)

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if fullErr == nil {
		_ = s.logService.Log(actor, "subject_class.unassigned", "subject_class", strPtr(id), domain.LogSeverityMedium, subjectClassAuditMetadata(full))
	}
	return nil
}

func (s *subjectClassService) validateAssignmentScope(classID string, subjectID string, teacherSchoolUserID string, schoolID string) error {
	if err := s.ensureClassInSchool(classID, schoolID); err != nil {
		return err
	}
	if err := s.ensureSubjectInSchool(subjectID, schoolID); err != nil {
		return err
	}
	if err := s.ensureTeacherEligible(teacherSchoolUserID, classID, schoolID); err != nil {
		return err
	}
	return nil
}

func (s *subjectClassService) ensureClassInSchool(classID string, schoolID string) error {
	ok, err := s.repo.ClassBelongsToSchool(classID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("forbidden: class does not belong to active school")
	}
	return nil
}

func (s *subjectClassService) ensureSubjectInSchool(subjectID string, schoolID string) error {
	ok, err := s.repo.SubjectBelongsToSchool(subjectID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("forbidden: subject does not belong to active school")
	}
	return nil
}

func (s *subjectClassService) ensureSubjectClassInSchool(subjectClassID string, schoolID string) error {
	ok, err := s.repo.SubjectClassBelongsToSchool(subjectClassID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("forbidden: subject class does not belong to active school")
	}
	return nil
}

func (s *subjectClassService) ensureTeacherEligible(schoolUserID string, classID string, schoolID string) error {
	inSchool, err := s.repo.SchoolUserBelongsToSchool(schoolUserID, schoolID)
	if err != nil {
		return err
	}
	if !inSchool {
		return fmt.Errorf("forbidden: teacher does not belong to active school")
	}

	hasTeacherRole, err := s.repo.SchoolUserHasRole(schoolUserID, "teacher")
	if err != nil {
		return err
	}
	if !hasTeacherRole {
		return fmt.Errorf("forbidden: school user does not have teacher role")
	}

	enrolledAsTeacher, err := s.repo.SchoolUserEnrolledInClassAsRole(schoolUserID, classID, schoolID, "teacher")
	if err != nil {
		return err
	}
	if !enrolledAsTeacher {
		return fmt.Errorf("forbidden: teacher is not enrolled in this class as teacher")
	}

	return nil
}

func (s *subjectClassService) ensureNoClassSubjectDuplicate(classID string, subjectID string, excludeID string) error {
	exists, err := s.repo.CheckClassSubjectExists(classID, subjectID, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("subject class already assigned for this class and subject")
	}
	return nil
}
