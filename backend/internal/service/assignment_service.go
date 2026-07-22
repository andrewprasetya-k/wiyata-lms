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
	CreateCategory(actor domain.ActorContext, cat *domain.AssignmentCategory) error
	GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error)

	// Assignment
	CreateAssignment(actor domain.ActorContext, asg *domain.Assignment, mediaIDs []string, actorUserID string, isAdmin bool) error
	GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error)
	GetAssignmentByID(id string) (*domain.Assignment, error)
	GetAssignmentWithSubmissions(id string) (*domain.Assignment, error)
	GetSubjectClassSubmissions(subjectClassID string, schoolID string) ([]*domain.Assignment, error)
	GetTeacherSubmissionInbox(userID string, schoolID string) (*dto.TeacherSubmissionInboxResponseDTO, error)
	GetTeacherAssignmentInbox(userID string, schoolID string) (*dto.TeacherAssignmentInboxResponseDTO, error)
	GetStudentAssignmentInbox(userID string, schoolID string) (*dto.StudentAssignmentInboxResponseDTO, error)
	GetAssignmentStatus(assignmentID string, schoolID string) (map[string]interface{}, error)
	UpdateAssignment(actor domain.ActorContext, id string, asg *domain.Assignment, mediaIDs []string, actorUserID string, isAdmin bool, validateCategory bool) error
	DeleteAssignment(actor domain.ActorContext, id string) error

	// Submission
	Submit(actor domain.ActorContext, sbm *domain.Submission, mediaIDs []string, actorUserID string, isAdmin bool) error
	GetSubmissions(asgID string) ([]*domain.Submission, error)
	GetSubmissionByID(id string) (*domain.Submission, error)
	GetMySubmissionByAssignment(assignmentID string, userID string, schoolID string) (*domain.Submission, error)
	UpdateSubmission(actor domain.ActorContext, id string, mediaIDs []string, actorUserID string, isAdmin bool) error
	DeleteSubmission(actor domain.ActorContext, id string) error

	// Assessment
	Assess(actor domain.ActorContext, asm *domain.Assessment) error
	UpdateAssessment(actor domain.ActorContext, submissionID string, asm *domain.Assessment) error
	DeleteAssessment(actor domain.ActorContext, submissionID string) error
}

type assignmentService struct {
	repo         repository.AssignmentRepository
	attService   AttachmentService
	mediaRepo    repository.MediaRepository
	notifService NotificationService
	enrRepo      repository.EnrollmentRepository
	db           *gorm.DB
	logService   LogService
}

func NewAssignmentService(repo repository.AssignmentRepository, attService AttachmentService, mediaRepo repository.MediaRepository, notifService NotificationService, enrRepo repository.EnrollmentRepository, db *gorm.DB, logService LogService) AssignmentService {
	return &assignmentService{
		repo:         repo,
		attService:   attService,
		mediaRepo:    mediaRepo,
		notifService: notifService,
		enrRepo:      enrRepo,
		db:           db,
		logService:   logService,
	}
}

func (s *assignmentService) CreateCategory(actor domain.ActorContext, cat *domain.AssignmentCategory) error {
	if err := s.repo.CreateCategory(cat); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.category.created", "assignment_category", strPtr(cat.ID), domain.LogSeverityMedium, map[string]any{
		"category_name": cat.Name,
	})

	return nil
}

func (s *assignmentService) GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error) {
	return s.repo.GetCategoriesBySchool(schoolID)
}

func (s *assignmentService) CreateAssignment(actor domain.ActorContext, asg *domain.Assignment, mediaIDs []string, actorUserID string, isAdmin bool) error {
	if err := s.validateAssignmentCategory(asg.CategoryID, asg.SchoolID); err != nil {
		return err
	}
	attachmentMediaIDs, err := prepareAttachableMediaIDs(s.mediaRepo, mediaIDs, asg.SchoolID, actorUserID, isAdmin)
	if err != nil {
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).CreateAssignment(asg); err != nil {
			return err
		}
		return replaceSourceAttachments(s.attService.WithTx(tx), asg.SchoolID, domain.SourceAssignment, asg.ID, attachmentMediaIDs)
	}); err != nil {
		return err
	}

	// Best-effort: notify students in the class
	runAsync(func() {
		if classID, err := s.repo.GetClassIDBySubjectClass(asg.SubjectClassID); err == nil && classID != "" {
			if userIDs, err := s.enrRepo.GetStudentUserIDsByClass(classID); err == nil {
				for _, uid := range userIDs {
					_ = s.notifService.Create(&dto.CreateNotificationDTO{
						UserID:    uid,
						Type:      domain.NotifAssignmentCreated,
						Title:     "Tugas baru",
						Message:   asg.Title,
						Link:      fmt.Sprintf("/student/subjects/%s/assignments/%s", asg.SubjectClassID, asg.ID),
						RelatedID: asg.ID,
					})
				}
			}
		}
	})

	_ = s.logService.Log(actor, "assignment.created", "assignment", strPtr(asg.ID), domain.LogSeverityMedium, map[string]any{
		"assignment_title": asg.Title,
		"class_id":         asg.SubjectClassID,
		"due_date":         asg.Deadline,
	})

	return nil
}

func (s *assignmentService) GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error) {
	results, total, err := s.repo.GetAssignmentsBySubjectClass(subjectClassID, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	sourceIDs := make([]string, 0, len(results))
	for _, asg := range results {
		sourceIDs = append(sourceIDs, asg.ID)
	}
	attachmentsBySource, err := s.attService.GetBySources(string(domain.SourceAssignment), sourceIDs)
	if err != nil {
		return nil, 0, err
	}

	for _, asg := range results {
		atts := attachmentsBySource[asg.ID]
		asg.Attachments = make([]domain.Attachment, 0, len(atts))
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

	atts, err := s.attService.GetBySource(string(domain.SourceAssignment), id)
	if err != nil {
		return nil, err
	}
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

	sourceIDs := make([]string, 0, len(asg.Submissions))
	for i := range asg.Submissions {
		sourceIDs = append(sourceIDs, asg.Submissions[i].ID)
	}
	attachmentsBySource, err := s.attService.GetBySources(string(domain.SourceSubmission), sourceIDs)
	if err != nil {
		return nil, err
	}

	for i := range asg.Submissions {
		atts := attachmentsBySource[asg.Submissions[i].ID]
		asg.Submissions[i].Attachments = make([]domain.Attachment, 0, len(atts))
		for _, a := range atts {
			asg.Submissions[i].Attachments = append(asg.Submissions[i].Attachments, *a)
		}
	}

	return asg, nil
}

func (s *assignmentService) GetSubjectClassSubmissions(subjectClassID string, schoolID string) ([]*domain.Assignment, error) {
	assignments, err := s.repo.GetAssignmentsWithSubmissionsBySubjectClass(subjectClassID, schoolID)
	if err != nil {
		return nil, err
	}

	sourceIDs := make([]string, 0)
	for _, asg := range assignments {
		for i := range asg.Submissions {
			sourceIDs = append(sourceIDs, asg.Submissions[i].ID)
		}
	}
	attachmentsBySource, err := s.attService.GetBySources(string(domain.SourceSubmission), sourceIDs)
	if err != nil {
		return nil, err
	}

	for _, asg := range assignments {
		for i := range asg.Submissions {
			atts := attachmentsBySource[asg.Submissions[i].ID]
			asg.Submissions[i].Attachments = make([]domain.Attachment, 0, len(atts))
			for _, a := range atts {
				asg.Submissions[i].Attachments = append(asg.Submissions[i].Attachments, *a)
			}
		}
	}

	return assignments, nil
}

func (s *assignmentService) GetTeacherSubmissionInbox(userID string, schoolID string) (*dto.TeacherSubmissionInboxResponseDTO, error) {
	items, err := s.repo.GetTeacherSubmissionInbox(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []dto.TeacherSubmissionInboxItemDTO{}
	}

	response := &dto.TeacherSubmissionInboxResponseDTO{
		Items: items,
	}
	for _, item := range items {
		response.Summary.TotalSubmissions += item.SubmissionCount
		response.Summary.PendingCount += item.PendingCount
		response.Summary.GradedCount += item.GradedCount
		response.Summary.LateCount += item.LateCount
	}

	return response, nil
}

func (s *assignmentService) GetTeacherAssignmentInbox(userID string, schoolID string) (*dto.TeacherAssignmentInboxResponseDTO, error) {
	items, err := s.repo.GetTeacherAssignmentInbox(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []dto.TeacherAssignmentInboxItemDTO{}
	}

	response := &dto.TeacherAssignmentInboxResponseDTO{
		Items: items,
	}
	response.Summary.TotalAssignments = len(items)
	now := time.Now()
	for _, item := range items {
		if item.Deadline != nil && item.Deadline.Before(now) {
			response.Summary.OverdueAssignments++
		} else {
			response.Summary.ActiveAssignments++
		}
		response.Summary.PendingReviewCount += item.PendingCount
		response.Summary.TotalSubmissions += item.SubmissionCount
	}

	return response, nil
}

func (s *assignmentService) GetStudentAssignmentInbox(userID string, schoolID string) (*dto.StudentAssignmentInboxResponseDTO, error) {
	items, err := s.repo.GetStudentAssignmentInbox(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []dto.StudentAssignmentInboxItemDTO{}
	}

	response := &dto.StudentAssignmentInboxResponseDTO{
		Items: items,
	}
	response.Summary.TotalAssignments = len(items)
	for _, item := range items {
		if item.IsSubmitted {
			response.Summary.SubmittedCount++
		} else {
			response.Summary.NotSubmittedCount++
		}
		if item.IsGraded {
			response.Summary.GradedCount++
		}
		if item.IsOverdue {
			response.Summary.OverdueCount++
		}
	}

	return response, nil
}

func (s *assignmentService) GetAssignmentStatus(assignmentID string, schoolID string) (map[string]interface{}, error) {
	asg, err := s.repo.GetAssignmentWithSubmissions(assignmentID)
	if err != nil {
		return nil, err
	}
	if asg.SchoolID != schoolID {
		return nil, fmt.Errorf("forbidden: assignment does not belong to active school")
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

func (s *assignmentService) UpdateAssignment(actor domain.ActorContext, id string, asg *domain.Assignment, mediaIDs []string, actorUserID string, isAdmin bool, validateCategory bool) error {
	asg.ID = id
	if validateCategory {
		if err := s.validateAssignmentCategory(asg.CategoryID, asg.SchoolID); err != nil {
			return err
		}
	}
	var attachmentMediaIDs []string
	if mediaIDs != nil {
		var err error
		attachmentMediaIDs, err = prepareAttachableMediaIDs(s.mediaRepo, mediaIDs, asg.SchoolID, actorUserID, isAdmin)
		if err != nil {
			return err
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).UpdateAssignment(asg); err != nil {
			return err
		}
		if mediaIDs != nil {
			if err := replaceSourceAttachments(s.attService.WithTx(tx), asg.SchoolID, domain.SourceAssignment, id, attachmentMediaIDs); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.updated", "assignment", strPtr(asg.ID), domain.LogSeverityMedium, map[string]any{
		"assignment_title": asg.Title,
		"class_id":         asg.SubjectClassID,
		"due_date":         asg.Deadline,
	})

	return nil
}

func (s *assignmentService) DeleteAssignment(actor domain.ActorContext, id string) error {
	asg, err := s.repo.GetAssignmentByID(id)
	if err != nil {
		return err
	}

	submissions, err := s.repo.GetSubmissionsByAssignment(id)
	if err != nil {
		return err
	}
	if len(submissions) > 0 {
		return fmt.Errorf("assignment cannot be deleted because it already has student submissions")
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.attService.WithTx(tx).UnlinkBySource(string(domain.SourceAssignment), id); err != nil {
			return err
		}
		return s.repo.WithTx(tx).DeleteAssignment(id)
	}); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.deleted", "assignment", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"assignment_title": asg.Title,
		"class_id":         asg.SubjectClassID,
		"due_date":         asg.Deadline,
	})

	return nil
}

func (s *assignmentService) Submit(actor domain.ActorContext, sbm *domain.Submission, mediaIDs []string, actorUserID string, isAdmin bool) error {
	sbm.SubmittedAt = time.Now()

	// Check deadline before submitting
	assignment, err := s.repo.GetAssignmentByID(sbm.AssignmentID)
	if err != nil {
		return err
	}

	if !assignment.AllowLateSubmission && assignment.Deadline != nil && assignment.Deadline.Before(sbm.SubmittedAt) {
		return fmt.Errorf("submission past due")
	}

	attachmentMediaIDs, err := prepareAttachableMediaIDs(s.mediaRepo, mediaIDs, sbm.SchoolID, actorUserID, isAdmin)
	if err != nil {
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).UpsertSubmission(sbm); err != nil {
			return err
		}
		return replaceSourceAttachments(s.attService.WithTx(tx), sbm.SchoolID, domain.SourceSubmission, sbm.ID, attachmentMediaIDs)
	}); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.submitted", "submission", strPtr(sbm.ID), domain.LogSeverityMedium, map[string]any{
		"assignment_id": sbm.AssignmentID,
		"student_id":    sbm.UserID,
	})

	return nil
}

func (s *assignmentService) GetSubmissions(asgID string) ([]*domain.Submission, error) {
	results, err := s.repo.GetSubmissionsByAssignment(asgID)
	if err != nil {
		return nil, err
	}

	sourceIDs := make([]string, 0, len(results))
	for _, sbm := range results {
		sourceIDs = append(sourceIDs, sbm.ID)
	}
	attachmentsBySource, _ := s.attService.GetBySources(string(domain.SourceSubmission), sourceIDs)

	for _, sbm := range results {
		for _, a := range attachmentsBySource[sbm.ID] {
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

func (s *assignmentService) Assess(actor domain.ActorContext, asm *domain.Assessment) error {
	if err := s.repo.UpsertAssessment(asm); err != nil {
		return err
	}

	sbm, err := s.repo.GetSubmissionByID(asm.SubmissionID)
	if err == nil {
		_ = s.notifService.Create(&dto.CreateNotificationDTO{
			UserID:    sbm.UserID,
			Type:      domain.NotifAssignmentGraded,
			Title:     "Tugas sudah dinilai",
			Message:   fmt.Sprintf("Nilai Anda sudah tersedia: %.2f", asm.Score),
			Link:      "/student/grades",
			RelatedID: asm.SubmissionID,
		})

		_ = s.logService.Log(actor, "assignment.assessed", "assessment", strPtr(asm.ID), domain.LogSeverityMedium, map[string]any{
			"assignment_id": sbm.AssignmentID,
			"submission_id": asm.SubmissionID,
			"student_id":    sbm.UserID,
			"score_after":   asm.Score,
		})
	}

	return nil
}

func (s *assignmentService) validateSubmissionMutable(sbm *domain.Submission, assignment *domain.Assignment) error {
	if sbm.Assessment != nil {
		return fmt.Errorf("submission already graded")
	}
	if !assignment.AllowLateSubmission && assignment.Deadline != nil && assignment.Deadline.Before(time.Now()) {
		return fmt.Errorf("assignment closed")
	}
	return nil
}

func (s *assignmentService) UpdateSubmission(actor domain.ActorContext, id string, mediaIDs []string, actorUserID string, isAdmin bool) error {
	sbm, err := s.repo.GetSubmissionByID(id)
	if err != nil {
		return err
	}
	assignment, err := s.repo.GetAssignmentByID(sbm.AssignmentID)
	if err != nil {
		return err
	}
	if err := s.validateSubmissionMutable(sbm, assignment); err != nil {
		return err
	}

	var attachmentMediaIDs []string
	if mediaIDs != nil {
		var err error
		attachmentMediaIDs, err = prepareAttachableMediaIDs(s.mediaRepo, mediaIDs, sbm.SchoolID, actorUserID, isAdmin)
		if err != nil {
			return err
		}
	}

	sbm.SubmittedAt = time.Now()

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).UpdateSubmission(sbm); err != nil {
			return err
		}
		if mediaIDs != nil {
			if err := replaceSourceAttachments(s.attService.WithTx(tx), sbm.SchoolID, domain.SourceSubmission, id, attachmentMediaIDs); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.submission.updated", "submission", strPtr(sbm.ID), domain.LogSeverityMedium, map[string]any{
		"assignment_id": sbm.AssignmentID,
		"student_id":    sbm.UserID,
	})

	return nil
}

func (s *assignmentService) DeleteSubmission(actor domain.ActorContext, id string) error {
	sbm, err := s.repo.GetSubmissionByID(id)
	if err != nil {
		return err
	}
	assignment, err := s.repo.GetAssignmentByID(sbm.AssignmentID)
	if err != nil {
		return err
	}
	if err := s.validateSubmissionMutable(sbm, assignment); err != nil {
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.attService.WithTx(tx).UnlinkBySource(string(domain.SourceSubmission), id); err != nil {
			return err
		}
		return s.repo.WithTx(tx).DeleteSubmission(id)
	}); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "assignment.submission.deleted", "submission", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"assignment_id": sbm.AssignmentID,
		"student_id":    sbm.UserID,
	})

	return nil
}

func (s *assignmentService) UpdateAssessment(actor domain.ActorContext, submissionID string, asm *domain.Assessment) error {
	asm.SubmissionID = submissionID

	before, beforeErr := s.repo.GetAssessmentBySubmission(submissionID)

	if err := s.repo.UpdateAssessment(asm); err != nil {
		return err
	}

	if beforeErr == nil {
		if after, err := s.repo.GetAssessmentBySubmission(submissionID); err == nil {
			if sbm, err := s.repo.GetSubmissionByID(submissionID); err == nil {
				_ = s.logService.Log(actor, "assignment.assessment.updated", "assessment", strPtr(after.ID), domain.LogSeverityMedium, map[string]any{
					"assignment_id":    sbm.AssignmentID,
					"submission_id":    submissionID,
					"student_id":       sbm.UserID,
					"score_before":     before.Score,
					"score_after":      after.Score,
					"feedback_changed": after.Feedback != before.Feedback,
				})
			}
		}
	}

	return nil
}

func (s *assignmentService) DeleteAssessment(actor domain.ActorContext, submissionID string) error {
	before, beforeErr := s.repo.GetAssessmentBySubmission(submissionID)
	var sbm *domain.Submission
	if beforeErr == nil {
		sbm, _ = s.repo.GetSubmissionByID(submissionID)
	}

	if err := s.repo.DeleteAssessment(submissionID); err != nil {
		return err
	}

	if beforeErr == nil && sbm != nil {
		_ = s.logService.Log(actor, "assignment.assessment.deleted", "assessment", strPtr(before.ID), domain.LogSeverityMedium, map[string]any{
			"assignment_id": sbm.AssignmentID,
			"submission_id": submissionID,
			"student_id":    sbm.UserID,
			"score_before":  before.Score,
		})
	}

	return nil
}

func (s *assignmentService) validateAssignmentCategory(categoryID string, schoolID string) error {
	allowed, err := s.repo.AssignmentCategoryBelongsToSchool(categoryID, schoolID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("invalid assignment category")
	}
	return nil
}
