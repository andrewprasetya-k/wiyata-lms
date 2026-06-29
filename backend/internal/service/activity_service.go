package service

import (
	"backend/internal/dto"
	"backend/internal/repository"
	"fmt"
	"sort"
	"time"
)

const maxActivityRangeDays = 60

type ActivityService interface {
	GetAcademicActivity(userID string, schoolID string, roles []string, from *time.Time, to *time.Time) (*dto.AcademicActivityResponseDTO, error)
}

type activityService struct {
	repo repository.ActivityRepository
	now  func() time.Time
}

func NewActivityService(repo repository.ActivityRepository) ActivityService {
	return &activityService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *activityService) GetAcademicActivity(userID string, schoolID string, roles []string, from *time.Time, to *time.Time) (*dto.AcademicActivityResponseDTO, error) {
	fromTime, toTime, err := s.normalizeRange(from, to)
	if err != nil {
		return nil, err
	}

	rows := make([]repository.ActivityRow, 0)
	if hasRole(roles, "student") {
		studentRows, err := s.studentActivity(userID, schoolID, fromTime, toTime)
		if err != nil {
			return nil, err
		}
		rows = append(rows, studentRows...)
	}

	if hasRole(roles, "teacher") {
		teacherRows, err := s.teacherActivity(userID, schoolID, fromTime, toTime)
		if err != nil {
			return nil, err
		}
		rows = append(rows, teacherRows...)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].EventAt.Equal(rows[j].EventAt) {
			return rows[i].Title < rows[j].Title
		}
		return rows[i].EventAt.Before(rows[j].EventAt)
	})

	items := make([]dto.AcademicActivityItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapActivityRow(row))
	}

	return &dto.AcademicActivityResponseDTO{Items: items}, nil
}

func (s *activityService) normalizeRange(from *time.Time, to *time.Time) (time.Time, time.Time, error) {
	today := startOfDay(s.now())
	fromTime := today
	toDate := today.AddDate(0, 0, 7)

	if from != nil {
		fromTime = startOfDay(*from)
	}
	if to != nil {
		toDate = startOfDay(*to)
	}
	if toDate.Before(fromTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("activity date range is invalid")
	}
	if toDate.Sub(fromTime) >= maxActivityRangeDays*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("activity date range exceeds 60 days")
	}

	return fromTime, toDate.AddDate(0, 0, 1), nil
}

func (s *activityService) studentActivity(userID string, schoolID string, from time.Time, to time.Time) ([]repository.ActivityRow, error) {
	var rows []repository.ActivityRow
	queries := []func(string, string, time.Time, time.Time) ([]repository.ActivityRow, error){
		s.repo.GetStudentAssignmentDue,
		s.repo.GetStudentMaterialCreated,
		s.repo.GetStudentFeedPosted,
		s.repo.GetStudentAssignmentGraded,
	}
	for _, query := range queries {
		result, err := query(userID, schoolID, from, to)
		if err != nil {
			return nil, err
		}
		rows = append(rows, result...)
	}
	return rows, nil
}

func (s *activityService) teacherActivity(userID string, schoolID string, from time.Time, to time.Time) ([]repository.ActivityRow, error) {
	var rows []repository.ActivityRow
	queries := []func(string, string, time.Time, time.Time) ([]repository.ActivityRow, error){
		s.repo.GetTeacherSubmissionPendingReview,
		s.repo.GetTeacherSubmissionReceived,
		s.repo.GetTeacherAssignmentDue,
		s.repo.GetTeacherFeedComments,
	}
	for _, query := range queries {
		result, err := query(userID, schoolID, from, to)
		if err != nil {
			return nil, err
		}
		rows = append(rows, result...)
	}
	return rows, nil
}

func mapActivityRow(row repository.ActivityRow) dto.AcademicActivityItemDTO {
	item := dto.AcademicActivityItemDTO{
		ID:          fmt.Sprintf("%s:%s", row.ActivityType, row.SourceID),
		Type:        row.ActivityType,
		Title:       row.Title,
		Description: row.Description,
		Date:        row.EventAt.Format("2006-01-02"),
		Time:        row.EventAt.Format("15:04"),
		Priority:    row.Priority,
		Link:        row.Link,
		Metadata:    mapActivityMetadata(row),
	}

	if row.SubjectID != "" {
		item.Subject = &dto.ActivitySubjectDTO{
			ID:    row.SubjectID,
			Name:  row.SubjectName,
			Code:  row.SubjectCode,
			Color: row.SubjectColor,
		}
	}

	if row.ClassID != "" {
		item.Class = &dto.ActivityClassDTO{
			ID:   row.ClassID,
			Name: row.ClassName,
			Code: row.ClassCode,
		}
	}

	return item
}

func mapActivityMetadata(row repository.ActivityRow) map[string]interface{} {
	metadata := map[string]interface{}{}
	addMetadata(metadata, "assignmentId", row.AssignmentID)
	addMetadata(metadata, "subjectClassId", row.SubjectClassID)
	addMetadata(metadata, "materialId", row.MaterialID)
	addMetadata(metadata, "feedId", row.FeedID)
	addMetadata(metadata, "commentId", row.CommentID)
	addMetadata(metadata, "submissionId", row.SubmissionID)
	addMetadata(metadata, "studentId", row.StudentID)
	addMetadata(metadata, "studentName", row.StudentName)
	return metadata
}

func addMetadata(metadata map[string]interface{}, key string, value string) {
	if value != "" {
		metadata[key] = value
	}
}

func startOfDay(value time.Time) time.Time {
	local := value.Local()
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

func hasRole(roles []string, expected string) bool {
	for _, role := range roles {
		if role == expected {
			return true
		}
	}
	return false
}
