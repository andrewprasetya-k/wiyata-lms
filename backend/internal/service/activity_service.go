package service

import (
	"backend/internal/dto"
	"backend/internal/repository"
	"fmt"
	"sort"
	"time"
)

const maxActivityRangeDays = 60
const activityDisplayTimezone = "Asia/Jakarta"

type ActivityService interface {
	GetAcademicActivity(userID string, schoolID string, roles []string, from *string, to *string) (*dto.AcademicActivityResponseDTO, error)
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

func (s *activityService) GetAcademicActivity(userID string, schoolID string, roles []string, from *string, to *string) (*dto.AcademicActivityResponseDTO, error) {
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

func (s *activityService) normalizeRange(from *string, to *string) (time.Time, time.Time, error) {
	location := activityLocation()
	today := startOfDayInLocation(s.now(), location)
	fromTime := today
	toDate := today.AddDate(0, 0, 7)

	if from != nil {
		parsed, err := parseActivityDateInLocation(*from, location)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		fromTime = parsed
	}
	if to != nil {
		parsed, err := parseActivityDateInLocation(*to, location)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		toDate = parsed
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
	queries := []struct {
		name string
		run  func(string, string, time.Time, time.Time) ([]repository.ActivityRow, error)
	}{
		{name: "student assignment_due", run: s.repo.GetStudentAssignmentDue},
		{name: "student material_created", run: s.repo.GetStudentMaterialCreated},
		{name: "student feed_posted", run: s.repo.GetStudentFeedPosted},
		{name: "student assignment_graded", run: s.repo.GetStudentAssignmentGraded},
	}
	for _, query := range queries {
		result, err := query.run(userID, schoolID, from, to)
		if err != nil {
			return nil, fmt.Errorf("academic activity query failed: %s: %w", query.name, err)
		}
		rows = append(rows, result...)
	}

	overdueRows, err := s.repo.GetStudentAssignmentOverdue(userID, schoolID, from.AddDate(0, 0, -30))
	if err != nil {
		return nil, fmt.Errorf("academic activity query failed: student assignment_overdue: %w", err)
	}
	rows = append(rows, overdueRows...)

	return rows, nil
}

func (s *activityService) teacherActivity(userID string, schoolID string, from time.Time, to time.Time) ([]repository.ActivityRow, error) {
	var rows []repository.ActivityRow
	queries := []struct {
		name string
		run  func(string, string, time.Time, time.Time) ([]repository.ActivityRow, error)
	}{
		{name: "teacher submission_pending_review", run: s.repo.GetTeacherSubmissionPendingReview},
		{name: "teacher submission_received", run: s.repo.GetTeacherSubmissionReceived},
		{name: "teacher assignment_due", run: s.repo.GetTeacherAssignmentDue},
		{name: "teacher feed_comment", run: s.repo.GetTeacherFeedComments},
	}
	for _, query := range queries {
		result, err := query.run(userID, schoolID, from, to)
		if err != nil {
			return nil, fmt.Errorf("academic activity query failed: %s: %w", query.name, err)
		}
		rows = append(rows, result...)
	}
	return rows, nil
}

func mapActivityRow(row repository.ActivityRow) dto.AcademicActivityItemDTO {
	eventAt := row.EventAt.In(activityLocation())
	item := dto.AcademicActivityItemDTO{
		ID:          fmt.Sprintf("%s:%s", row.ActivityType, row.SourceID),
		Type:        row.ActivityType,
		Title:       row.Title,
		Description: row.Description,
		Date:        eventAt.Format("2006-01-02"),
		Time:        eventAt.Format("15:04"),
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

func activityLocation() *time.Location {
	location, err := time.LoadLocation(activityDisplayTimezone)
	if err != nil {
		return time.Local
	}
	return location
}

func parseActivityDateInLocation(value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", value, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("activity date range is invalid")
	}
	return startOfDayInLocation(parsed, location), nil
}

func startOfDayInLocation(value time.Time, location *time.Location) time.Time {
	local := value.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}

func hasRole(roles []string, expected string) bool {
	for _, role := range roles {
		if role == expected {
			return true
		}
	}
	return false
}
