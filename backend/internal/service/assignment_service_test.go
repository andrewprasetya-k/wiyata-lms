package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"strings"
	"testing"
	"time"
)

// ─── AssignmentRepository stub ────────────────────────────────────────────────

type asgRepoStub struct {
	assignment        *domain.Assignment
	getAssignmentErr  error
	submissions       []*domain.Submission
	getSubmissionsErr error
	upsertErr         error
}

func (r *asgRepoStub) GetAssignmentByID(_ string) (*domain.Assignment, error) {
	return r.assignment, r.getAssignmentErr
}
func (r *asgRepoStub) GetSubmissionsByAssignment(_ string) ([]*domain.Submission, error) {
	return r.submissions, r.getSubmissionsErr
}
func (r *asgRepoStub) UpsertSubmission(_ *domain.Submission) error { return r.upsertErr }

// No-ops for remaining interface methods:
func (r *asgRepoStub) CreateCategory(_ *domain.AssignmentCategory) error { return nil }
func (r *asgRepoStub) GetCategoriesBySchool(_ string) ([]*domain.AssignmentCategory, error) {
	return nil, nil
}
func (r *asgRepoStub) AssignmentCategoryBelongsToSchool(_, _ string) (bool, error) {
	return true, nil
}
func (r *asgRepoStub) CreateAssignment(_ *domain.Assignment) error           { return nil }
func (r *asgRepoStub) GetAssignmentsBySubjectClass(_, _ string, _, _ int) ([]*domain.Assignment, int64, error) {
	return nil, 0, nil
}
func (r *asgRepoStub) GetAssignmentWithSubmissions(_ string) (*domain.Assignment, error) {
	return nil, nil
}
func (r *asgRepoStub) GetAssignmentsWithSubmissionsBySubjectClass(_, _ string) ([]*domain.Assignment, error) {
	return nil, nil
}
func (r *asgRepoStub) GetTeacherSubmissionInbox(_, _ string) ([]dto.TeacherSubmissionInboxItemDTO, error) {
	return nil, nil
}
func (r *asgRepoStub) GetTeacherAssignmentInbox(_, _ string) ([]dto.TeacherAssignmentInboxItemDTO, error) {
	return nil, nil
}
func (r *asgRepoStub) GetStudentAssignmentInbox(_, _ string) ([]dto.StudentAssignmentInboxItemDTO, error) {
	return nil, nil
}
func (r *asgRepoStub) CountStudentsInClass(_ string) (int, error)    { return 0, nil }
func (r *asgRepoStub) GetClassIDBySubjectClass(_ string) (string, error) { return "", nil }
func (r *asgRepoStub) UpdateAssignment(_ *domain.Assignment) error       { return nil }
func (r *asgRepoStub) DeleteAssignment(_ string) error                   { return nil }
func (r *asgRepoStub) GetSubmissionByID(_ string) (*domain.Submission, error) {
	return nil, nil
}
func (r *asgRepoStub) GetMySubmissionByAssignment(_, _, _ string) (*domain.Submission, error) {
	return nil, nil
}
func (r *asgRepoStub) UpdateSubmission(_ *domain.Submission) error       { return nil }
func (r *asgRepoStub) DeleteSubmission(_ string) error                   { return nil }
func (r *asgRepoStub) UpsertAssessment(_ *domain.Assessment) error       { return nil }
func (r *asgRepoStub) GetAssessmentBySubmission(_ string) (*domain.Assessment, error) {
	return nil, nil
}
func (r *asgRepoStub) UpdateAssessment(_ *domain.Assessment) error       { return nil }
func (r *asgRepoStub) DeleteAssessment(_ string) error                   { return nil }
func (r *asgRepoStub) SetWeight(_ *domain.AssessmentWeight) error        { return nil }
func (r *asgRepoStub) GetWeightsBySubject(_ string) ([]*domain.AssessmentWeight, error) {
	return nil, nil
}
func (r *asgRepoStub) DeleteBySubject(_ string) error                    { return nil }
func (r *asgRepoStub) GetTotalWeightBySubject(_ string) (float64, error) { return 0, nil }

// ─── AttachmentService stub ───────────────────────────────────────────────────

type asgAttServiceStub struct{}

func (r *asgAttServiceStub) Link(_ *domain.Attachment) error                                { return nil }
func (r *asgAttServiceStub) GetBySource(_, _ string) ([]*domain.Attachment, error)          { return nil, nil }
func (r *asgAttServiceStub) GetBySources(_ string, _ []string) (map[string][]*domain.Attachment, error) {
	return nil, nil
}
func (r *asgAttServiceStub) Unlink(_ string) error                                          { return nil }
func (r *asgAttServiceStub) UnlinkBySource(_, _ string) error                               { return nil }
func (r *asgAttServiceStub) ReplaceBySource(_, _, _ string, _ []string) error               { return nil }

// ─── MediaRepository stub ─────────────────────────────────────────────────────

type asgMediaRepoStub struct{}

func (r *asgMediaRepoStub) Create(_ *domain.Media) error                                       { return nil }
func (r *asgMediaRepoStub) GetByID(_ string) (*domain.Media, error)                            { return nil, nil }
func (r *asgMediaRepoStub) GetByIDs(_ []string) ([]*domain.Media, error)                       { return nil, nil }
func (r *asgMediaRepoStub) GetByOwner(_ domain.OwnerType, _ string) ([]*domain.Media, error)   { return nil, nil }
func (r *asgMediaRepoStub) Delete(_ string) error                                              { return nil }

// ─── NotificationService stub ─────────────────────────────────────────────────

type asgNotifServiceStub struct{}

func (r *asgNotifServiceStub) Create(_ *dto.CreateNotificationDTO) error                          { return nil }
func (r *asgNotifServiceStub) GetByUserID(_ string, _, _ int, _ bool) (*dto.NotificationListDTO, error) {
	return nil, nil
}
func (r *asgNotifServiceStub) GetUnreadCount(_ string) (*dto.UnreadCountDTO, error)              { return nil, nil }
func (r *asgNotifServiceStub) GetFeedUnreadCount(_, _ string, _ []string) (*dto.UnreadCountDTO, error) {
	return nil, nil
}
func (r *asgNotifServiceStub) MarkAsRead(_, _ string) error                                      { return nil }
func (r *asgNotifServiceStub) MarkAllAsRead(_ string) error                                      { return nil }
func (r *asgNotifServiceStub) MarkFeedNotificationsRead(_, _ string, _ []string) error           { return nil }
func (r *asgNotifServiceStub) Delete(_, _ string) error                                          { return nil }

// ─── EnrollmentRepository stub ────────────────────────────────────────────────

type asgEnrRepoStub struct{}

func (r *asgEnrRepoStub) Create(_ *domain.Enrollment) error                                          { return nil }
func (r *asgEnrRepoStub) GetByID(_ string) (*domain.Enrollment, error)                               { return nil, nil }
func (r *asgEnrRepoStub) GetByClassAndSchoolUser(_, _ string) (*domain.Enrollment, error)             { return nil, nil }
func (r *asgEnrRepoStub) GetByClass(_ string, _ string, _, _ int) ([]*domain.Enrollment, int64, error) {
	return nil, 0, nil
}
func (r *asgEnrRepoStub) GetByMember(_ string) ([]*domain.Enrollment, error)                         { return nil, nil }
func (r *asgEnrRepoStub) Update(_, _ string) error                                                   { return nil }
func (r *asgEnrRepoStub) Reactivate(_, _ string) error                                               { return nil }
func (r *asgEnrRepoStub) SoftDelete(_ string) error                                                  { return nil }
func (r *asgEnrRepoStub) CheckExists(_, _ string) (bool, error)                                      { return false, nil }
func (r *asgEnrRepoStub) BelongsToSchool(_, _ string) (bool, error)                                  { return false, nil }
func (r *asgEnrRepoStub) ActiveBelongsToSchool(_, _ string) (bool, error)                            { return false, nil }
func (r *asgEnrRepoStub) HasTeacherSubjectClassAssignment(_, _, _ string) (bool, error)               { return false, nil }
func (r *asgEnrRepoStub) GetStudentUserIDsByClass(_ string) ([]string, error)                         { return nil, nil }
func (r *asgEnrRepoStub) GetMemberUserIDsByClass(_ string) ([]string, error)                          { return nil, nil }
func (r *asgEnrRepoStub) UserEnrolledInClassAsRole(_, _, _, _ string) (bool, error)                   { return false, nil }
func (r *asgEnrRepoStub) BulkCloseBySchoolUser(_ string, _ time.Time) error                           { return nil }

// ─── Helper ───────────────────────────────────────────────────────────────────

func newAssignmentTestService(repo *asgRepoStub) AssignmentService {
	return NewAssignmentService(
		repo,
		&asgAttServiceStub{},
		&asgMediaRepoStub{},
		&asgNotifServiceStub{},
		&asgEnrRepoStub{},
	)
}

func pastTime() *time.Time {
	t := time.Now().Add(-24 * time.Hour)
	return &t
}

func futureTime() *time.Time {
	t := time.Now().Add(24 * time.Hour)
	return &t
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestAssignmentSubmitRejectsPastDeadlineWhenLateNotAllowed(t *testing.T) {
	repo := &asgRepoStub{
		assignment: &domain.Assignment{
			ID:                  "asg-1",
			AllowLateSubmission: false,
			Deadline:            pastTime(),
		},
	}
	svc := newAssignmentTestService(repo)

	sbm := &domain.Submission{AssignmentID: "asg-1", UserID: "user-1", SchoolID: "school-1"}
	err := svc.Submit(sbm, nil, "user-1", false)
	if err == nil || !strings.Contains(err.Error(), "submission past due") {
		t.Fatalf("expected past due error, got %v", err)
	}
}

func TestAssignmentSubmitAllowsPastDeadlineWhenLateAllowed(t *testing.T) {
	repo := &asgRepoStub{
		assignment: &domain.Assignment{
			ID:                  "asg-1",
			AllowLateSubmission: true,
			Deadline:            pastTime(),
		},
	}
	svc := newAssignmentTestService(repo)

	sbm := &domain.Submission{AssignmentID: "asg-1", UserID: "user-1", SchoolID: "school-1"}
	if err := svc.Submit(sbm, nil, "user-1", false); err != nil {
		t.Fatalf("expected success for late submission when allowed, got %v", err)
	}
}

func TestAssignmentSubmitAllowsSubmissionBeforeDeadline(t *testing.T) {
	repo := &asgRepoStub{
		assignment: &domain.Assignment{
			ID:                  "asg-1",
			AllowLateSubmission: false,
			Deadline:            futureTime(),
		},
	}
	svc := newAssignmentTestService(repo)

	sbm := &domain.Submission{AssignmentID: "asg-1", UserID: "user-1", SchoolID: "school-1"}
	if err := svc.Submit(sbm, nil, "user-1", false); err != nil {
		t.Fatalf("expected success for on-time submission, got %v", err)
	}
}

func TestAssignmentSubmitAllowsSubmissionWithNoDeadline(t *testing.T) {
	repo := &asgRepoStub{
		assignment: &domain.Assignment{
			ID:                  "asg-1",
			AllowLateSubmission: false,
			Deadline:            nil,
		},
	}
	svc := newAssignmentTestService(repo)

	sbm := &domain.Submission{AssignmentID: "asg-1", UserID: "user-1", SchoolID: "school-1"}
	if err := svc.Submit(sbm, nil, "user-1", false); err != nil {
		t.Fatalf("expected success for assignment with no deadline, got %v", err)
	}
}

func TestAssignmentDeleteRejectsWhenSubmissionsExist(t *testing.T) {
	repo := &asgRepoStub{
		submissions: []*domain.Submission{
			{ID: "sbm-1", AssignmentID: "asg-1"},
		},
	}
	svc := newAssignmentTestService(repo)

	err := svc.DeleteAssignment("asg-1")
	if err == nil || !strings.Contains(err.Error(), "cannot be deleted") {
		t.Fatalf("expected cannot-be-deleted error, got %v", err)
	}
}

func TestAssignmentDeleteSucceedsWhenNoSubmissions(t *testing.T) {
	repo := &asgRepoStub{submissions: []*domain.Submission{}}
	svc := newAssignmentTestService(repo)

	if err := svc.DeleteAssignment("asg-1"); err != nil {
		t.Fatalf("expected success for assignment with no submissions, got %v", err)
	}
}
