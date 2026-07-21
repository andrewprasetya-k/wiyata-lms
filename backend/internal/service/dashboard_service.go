package service

import (
	"backend/internal/dto"
	"backend/internal/repository"
	"math"
	"time"
)

type DashboardService interface {
	GetStudentDashboard(userID string) (*dto.StudentDashboardDTO, error)
	GetTeacherDashboard(schoolUserID string) (*dto.TeacherDashboardDTO, error)
	GetAdminDashboard(schoolID string) (*dto.AdminDashboardDTO, error)
	GetSuperAdminDashboard() (*dto.SuperAdminDashboardDTO, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func safePercentage(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func dashboardStringValue(value any) string {
	if result, ok := value.(string); ok {
		return result
	}
	return ""
}

func dashboardTimestampValue(value any) string {
	timestamp, ok := dashboardTimeValue(value)
	if !ok {
		return ""
	}
	return formatAPITime(timestamp)
}

func dashboardTimeValue(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return time.Time{}, false
		}
		return typed, true
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"02-01-2006 15:04:05",
			"02-01-2006 15:04",
		} {
			parsed, err := time.Parse(layout, typed)
			if err == nil && !parsed.IsZero() {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}

func (s *dashboardService) GetStudentDashboard(userID string) (*dto.StudentDashboardDTO, error) {
	pending, err := s.repo.GetPendingAssignmentsCount(userID)
	if err != nil {
		return nil, err
	}

	deadlines, err := s.repo.GetUpcomingDeadlines(userID, 5)
	if err != nil {
		return nil, err
	}

	avgScore, err := s.repo.GetAverageScore(userID)
	if err != nil {
		return nil, err
	}

	completed, total, err := s.repo.GetMaterialProgress(userID)
	if err != nil {
		return nil, err
	}

	var upcomingDeadlines []dto.AssignmentDeadlineDTO
	for _, d := range deadlines {
		upcomingDeadlines = append(upcomingDeadlines, dto.AssignmentDeadlineDTO{
			AssignmentID:    d["assignment_id"].(string),
			AssignmentTitle: d["assignment_title"].(string),
			SubjectName:     d["subject_name"].(string),
			Deadline:        dashboardTimestampValue(d["deadline"]),
			IsSubmitted:     d["is_submitted"].(bool),
		})
	}

	return &dto.StudentDashboardDTO{
		PendingAssignments: pending,
		UpcomingDeadlines:  upcomingDeadlines,
		AverageScore:       avgScore,
		CompletedMaterials: completed,
		TotalMaterials:     total,
	}, nil
}

func (s *dashboardService) GetTeacherDashboard(schoolUserID string) (*dto.TeacherDashboardDTO, error) {
	pending, err := s.repo.GetPendingReviewsCount(schoolUserID)
	if err != nil {
		return nil, err
	}

	performance, err := s.repo.GetClassPerformance(schoolUserID)
	if err != nil {
		return nil, err
	}

	var classPerformance []dto.ClassPerformanceDTO
	for _, p := range performance {
		classPerformance = append(classPerformance, dto.ClassPerformanceDTO{
			ClassID:        p["class_id"].(string),
			ClassName:      p["class_name"].(string),
			SubjectName:    p["subject_name"].(string),
			SubjectColor:   dashboardStringValue(p["subject_color"]),
			AverageScore:   p["average_score"].(float64),
			SubmissionRate: safePercentage(p["submission_rate"].(float64)),
			TotalStudents:  int(p["total_students"].(int64)),
		})
	}

	return &dto.TeacherDashboardDTO{
		PendingReviews:   pending,
		ClassPerformance: classPerformance,
	}, nil
}

func (s *dashboardService) GetAdminDashboard(schoolID string) (*dto.AdminDashboardDTO, error) {
	stats, err := s.repo.GetSchoolStatistics(schoolID)
	if err != nil {
		return nil, err
	}

	trends, err := s.repo.GetEnrollmentTrends(schoolID)
	if err != nil {
		return nil, err
	}

	activities, err := s.repo.GetRecentActivities(schoolID, 10)
	if err != nil {
		return nil, err
	}

	var enrollmentTrends []dto.EnrollmentTrendDTO
	for _, t := range trends {
		enrollmentTrends = append(enrollmentTrends, dto.EnrollmentTrendDTO{
			ClassName:     t["class_name"].(string),
			TotalEnrolled: int(t["total_enrolled"].(int64)),
			Teachers:      int(t["teachers"].(int64)),
			Students:      int(t["students"].(int64)),
		})
	}

	var recentActivities []dto.ActivityLogDTO
	for _, a := range activities {
		recentActivities = append(recentActivities, dto.ActivityLogDTO{
			UserName:  a["user_name"].(string),
			Action:    a["action"].(string),
			Timestamp: dashboardTimestampValue(a["timestamp"]),
		})
	}

	classesWithoutTeacher, classesWithoutTeacherTotal, err := s.repo.GetClassesWithoutTeacher(schoolID, 5)
	if err != nil {
		return nil, err
	}

	var classesWithoutTeacherDTOs []dto.ClassWithoutTeacherDTO
	for _, c := range classesWithoutTeacher {
		classesWithoutTeacherDTOs = append(classesWithoutTeacherDTOs, dto.ClassWithoutTeacherDTO{
			ClassID:   c["class_id"].(string),
			ClassName: c["class_name"].(string),
		})
	}

	contentLessSubjectClasses, contentLessSubjectClassesTotal, err := s.repo.GetContentLessSubjectClasses(schoolID, 5)
	if err != nil {
		return nil, err
	}

	var contentLessSubjectClassDTOs []dto.ContentLessSubjectClassDTO
	for _, c := range contentLessSubjectClasses {
		contentLessSubjectClassDTOs = append(contentLessSubjectClassDTOs, dto.ContentLessSubjectClassDTO{
			SubjectClassID: c["subject_class_id"].(string),
			ClassName:      c["class_name"].(string),
			SubjectName:    c["subject_name"].(string),
		})
	}

	subjectsWithoutWeight, subjectsWithoutWeightTotal, err := s.repo.GetSubjectsWithoutAssessmentWeight(schoolID, 5)
	if err != nil {
		return nil, err
	}

	var subjectsWithoutWeightDTOs []dto.SubjectWithoutAssessmentWeightDTO
	for _, sub := range subjectsWithoutWeight {
		subjectsWithoutWeightDTOs = append(subjectsWithoutWeightDTOs, dto.SubjectWithoutAssessmentWeightDTO{
			SubjectID:   sub["subject_id"].(string),
			SubjectName: sub["subject_name"].(string),
		})
	}

	return &dto.AdminDashboardDTO{
		TotalStudents:                        stats["totalStudents"],
		TotalTeachers:                        stats["totalTeachers"],
		TotalClasses:                         stats["totalClasses"],
		ActiveClasses:                        stats["activeClasses"],
		EnrollmentTrends:                     enrollmentTrends,
		RecentActivities:                     recentActivities,
		ClassesWithoutTeacher:                classesWithoutTeacherDTOs,
		ClassesWithoutTeacherTotal:           classesWithoutTeacherTotal,
		ContentLessSubjectClasses:            contentLessSubjectClassDTOs,
		ContentLessSubjectClassesTotal:       contentLessSubjectClassesTotal,
		SubjectsWithoutAssessmentWeight:      subjectsWithoutWeightDTOs,
		SubjectsWithoutAssessmentWeightTotal: subjectsWithoutWeightTotal,
	}, nil
}

func (s *dashboardService) GetSuperAdminDashboard() (*dto.SuperAdminDashboardDTO, error) {
	withoutAdmin, withoutAdminTotal, err := s.repo.GetSchoolsWithoutAdmin(5)
	if err != nil {
		return nil, err
	}

	withoutSetup, withoutSetupTotal, err := s.repo.GetSchoolsWithoutSetup(5)
	if err != nil {
		return nil, err
	}

	return &dto.SuperAdminDashboardDTO{
		SchoolsWithoutAdmin:      mapSchoolsNeedingAttention(withoutAdmin),
		SchoolsWithoutAdminTotal: withoutAdminTotal,
		SchoolsWithoutSetup:      mapSchoolsNeedingAttention(withoutSetup),
		SchoolsWithoutSetupTotal: withoutSetupTotal,
	}, nil
}

func mapSchoolsNeedingAttention(rows []map[string]interface{}) []dto.SchoolNeedsAttentionDTO {
	var result []dto.SchoolNeedsAttentionDTO
	for _, r := range rows {
		result = append(result, dto.SchoolNeedsAttentionDTO{
			SchoolID:   r["school_id"].(string),
			SchoolName: r["school_name"].(string),
			SchoolCode: r["school_code"].(string),
			CreatedAt:  dashboardTimestampValue(r["created_at"]),
		})
	}
	return result
}
