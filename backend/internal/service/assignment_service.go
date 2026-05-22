package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type AssignmentService interface {
	// Category
	CreateCategory(cat *domain.AssignmentCategory) error
	GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error)

	// Assignment
	CreateAssignment(asg *domain.Assignment, mediaIDs []string) error
	GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error)
	GetAssignmentByID(id string) (*domain.Assignment, error)
	GetAssignmentWithSubmissions(id string) (*domain.Assignment, error)
	GetAssignmentStatus(assignmentID string) (map[string]interface{}, error)
	UpdateAssignment(id string, asg *domain.Assignment, mediaIDs []string) error
	DeleteAssignment(id string) error

	// Submission
	Submit(sbm *domain.Submission, mediaIDs []string) error
	GetSubmissions(asgID string) ([]*domain.Submission, error)
	GetSubmissionByID(id string) (*domain.Submission, error)
	GetMySubmissionByAssignment(assignmentID string, userID string, schoolID string) (*domain.Submission, error)
	UpdateSubmission(id string, mediaIDs []string) error
	DeleteSubmission(id string) error

	// Assessment
	Assess(asm *domain.Assessment) error
	UpdateAssessment(submissionID string, asm *domain.Assessment) error
	DeleteAssessment(submissionID string) error
}

type assignmentService struct {
	repo         repository.AssignmentRepository
	attService   AttachmentService
	notifService NotificationService
	enrRepo      repository.EnrollmentRepository
}

func NewAssignmentService(repo repository.AssignmentRepository, attService AttachmentService, notifService NotificationService, enrRepo repository.EnrollmentRepository) AssignmentService {
	return &assignmentService{
		repo:         repo,
		attService:   attService,
		notifService: notifService,
		enrRepo:      enrRepo,
	}
}

func (s *assignmentService) CreateCategory(cat *domain.AssignmentCategory) error {
	return s.repo.CreateCategory(cat)
}

func (s *assignmentService) GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error) {
	return s.repo.GetCategoriesBySchool(schoolID)
}

func (s *assignmentService) CreateAssignment(asg *domain.Assignment, mediaIDs []string) error {
	err := s.repo.CreateAssignment(asg)
	if err != nil {
		return err
	}

	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   asg.SchoolID,
			SourceID:   asg.ID,
			SourceType: domain.SourceAssignment,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}

	// Best-effort: notify students in the class
	if classID, err := s.repo.GetClassIDBySubjectClass(asg.SubjectClassID); err == nil && classID != "" {
		if userIDs, err := s.enrRepo.GetStudentUserIDsByClass(classID); err == nil {
			for _, uid := range userIDs {
				_ = s.notifService.Create(&dto.CreateNotificationDTO{
					UserID:    uid,
					Type:      domain.NotifAssignmentCreated,
					Title:     "New Assignment",
					Message:   "A new assignment has been posted: " + asg.Title,
					RelatedID: asg.ID,
				})
			}
		}
	}

	return nil
}

func (s *assignmentService) GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error) {
	results, total, err := s.repo.GetAssignmentsBySubjectClass(subjectClassID, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, asg := range results {
		atts, _ := s.attService.GetBySource(string(domain.SourceAssignment), asg.ID)
		for _, a := range atts {
			asg.Attachments = append(asg.Attachments, *a)
		}
	}
	return results, total, nil
}

func (s *assignmentService) GetAssignmentByID(id string) (*domain.Assignment, error) {
	asg, err := s.repo.GetAssignmentByID(id)
	if err != nil {
		return nil, err
	}

	atts, _ := s.attService.GetBySource(string(domain.SourceAssignment), id)
	for _, a := range atts {
		asg.Attachments = append(asg.Attachments, *a)
	}
	return asg, nil
}

func (s *assignmentService) GetAssignmentWithSubmissions(id string) (*domain.Assignment, error) {
	asg, err := s.repo.GetAssignmentWithSubmissions(id)
	if err != nil {
		return nil, err
	}

	// Load attachments for each submission
	for i := range asg.Submissions {
		atts, _ := s.attService.GetBySource(string(domain.SourceSubmission), asg.Submissions[i].ID)
		for _, a := range atts {
			asg.Submissions[i].Attachments = append(asg.Submissions[i].Attachments, *a)
		}
	}

	return asg, nil
}

func (s *assignmentService) GetAssignmentStatus(assignmentID string) (map[string]interface{}, error) {
	// Get assignment with submissions
	asg, err := s.repo.GetAssignmentWithSubmissions(assignmentID)
	if err != nil {
		return nil, err
	}

	// Get total enrolled students in the class
	totalStudents, err := s.repo.CountStudentsInClass(asg.SubjectClass.ClassID)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	submitted := len(asg.Submissions)
	notSubmitted := totalStudents - submitted

	graded := 0
	lateSubmissions := 0
	for _, sub := range asg.Submissions {
		if sub.Assessment != nil {
			graded++
		}
		if asg.Deadline != nil && sub.SubmittedAt.After(*asg.Deadline) {
			lateSubmissions++
		}
	}
	ungraded := submitted - graded

	submissionRate := 0.0
	if totalStudents > 0 {
		submissionRate = float64(submitted) / float64(totalStudents) * 100
	}

	return map[string]interface{}{
		"totalStudents":   totalStudents,
		"submitted":       submitted,
		"notSubmitted":    notSubmitted,
		"graded":          graded,
		"ungraded":        ungraded,
		"lateSubmissions": lateSubmissions,
		"submissionRate":  submissionRate,
	}, nil
}

func (s *assignmentService) UpdateAssignment(id string, asg *domain.Assignment, mediaIDs []string) error {
	asg.ID = id
	err := s.repo.UpdateAssignment(asg)
	if err != nil {
		return err
	}

	// Update attachments
	s.attService.UnlinkBySource(string(domain.SourceAssignment), id)
	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   asg.SchoolID,
			SourceID:   id,
			SourceType: domain.SourceAssignment,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}
	return nil
}

func (s *assignmentService) DeleteAssignment(id string) error {
	s.attService.UnlinkBySource(string(domain.SourceAssignment), id)
	return s.repo.DeleteAssignment(id)
}

func (s *assignmentService) Submit(sbm *domain.Submission, mediaIDs []string) error {
	sbm.SubmittedAt = time.Now()

	// Check deadline before submitting
	assignment, err := s.repo.GetAssignmentByID(sbm.AssignmentID)
	if err != nil {
		return err
	}

	if !assignment.AllowLateSubmission && assignment.Deadline != nil && assignment.Deadline.Before(sbm.SubmittedAt) {
		return fmt.Errorf("submission past due")
	}

	err = s.repo.UpsertSubmission(sbm)
	if err != nil {
		return err
	}

	// Unlink existing attachments for this submission if updating
	s.attService.UnlinkBySource(string(domain.SourceSubmission), sbm.ID)

	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   sbm.SchoolID,
			SourceID:   sbm.ID,
			SourceType: domain.SourceSubmission,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}
	return nil
}

func (s *assignmentService) GetSubmissions(asgID string) ([]*domain.Submission, error) {
	results, err := s.repo.GetSubmissionsByAssignment(asgID)
	if err != nil {
		return nil, err
	}

	for _, sbm := range results {
		atts, _ := s.attService.GetBySource(string(domain.SourceSubmission), sbm.ID)
		for _, a := range atts {
			sbm.Attachments = append(sbm.Attachments, *a)
		}
	}
	return results, nil
}

func (s *assignmentService) GetSubmissionByID(id string) (*domain.Submission, error) {
	sbm, err := s.repo.GetSubmissionByID(id)
	if err != nil {
		return nil, err
	}

	atts, _ := s.attService.GetBySource(string(domain.SourceSubmission), id)
	for _, a := range atts {
		sbm.Attachments = append(sbm.Attachments, *a)
	}
	return sbm, nil
}

func (s *assignmentService) GetMySubmissionByAssignment(assignmentID string, userID string, schoolID string) (*domain.Submission, error) {
	sbm, err := s.repo.GetMySubmissionByAssignment(assignmentID, userID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	atts, _ := s.attService.GetBySource(string(domain.SourceSubmission), sbm.ID)
	for _, a := range atts {
		sbm.Attachments = append(sbm.Attachments, *a)
	}
	return sbm, nil
}

func (s *assignmentService) Assess(asm *domain.Assessment) error {
	if err := s.repo.UpsertAssessment(asm); err != nil {
		return err
	}

	sbm, err := s.repo.GetSubmissionByID(asm.SubmissionID)
	if err == nil {
		_ = s.notifService.Create(&dto.CreateNotificationDTO{
			UserID:    sbm.UserID,
			Type:      domain.NotifAssignmentGraded,
			Title:     "Assignment Graded",
			Message:   fmt.Sprintf("Your submission has been graded. Score: %.2f", asm.Score),
			RelatedID: asm.SubmissionID,
		})
	}

	return nil
}

func (s *assignmentService) UpdateSubmission(id string, mediaIDs []string) error {
	sbm, err := s.repo.GetSubmissionByID(id)
	if err != nil {
		return err
	}

	sbm.SubmittedAt = time.Now()
	err = s.repo.UpdateSubmission(sbm)
	if err != nil {
		return err
	}

	s.attService.UnlinkBySource(string(domain.SourceSubmission), id)
	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   sbm.SchoolID,
			SourceID:   id,
			SourceType: domain.SourceSubmission,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}
	return nil
}

func (s *assignmentService) DeleteSubmission(id string) error {
	s.attService.UnlinkBySource(string(domain.SourceSubmission), id)
	return s.repo.DeleteSubmission(id)
}

func (s *assignmentService) UpdateAssessment(submissionID string, asm *domain.Assessment) error {
	asm.SubmissionID = submissionID
	return s.repo.UpdateAssessment(asm)
}

func (s *assignmentService) DeleteAssessment(submissionID string) error {
	return s.repo.DeleteAssessment(submissionID)
}
