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
		deadline, _ := time.Parse(time.RFC3339, d["deadline"].(string))
		upcomingDeadlines = append(upcomingDeadlines, dto.AssignmentDeadlineDTO{
			AssignmentID:    d["assignment_id"].(string),
			AssignmentTitle: d["assignment_title"].(string),
			SubjectName:     d["subject_name"].(string),
			Deadline:        deadline.Format("02-01-2006 15:04"),
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

	totalStudents, err := s.repo.GetTotalStudentsByTeacher(schoolUserID)
	if err != nil {
		return nil, err
	}

	submissionRate, err := s.repo.GetSubmissionRateByTeacher(schoolUserID)
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
			AverageScore:   p["average_score"].(float64),
			SubmissionRate: safePercentage(p["submission_rate"].(float64)),
			TotalStudents:  int(p["total_students"].(int64)),
		})
	}

	return &dto.TeacherDashboardDTO{
		PendingReviews:   pending,
		TotalStudents:    totalStudents,
		SubmissionRate:   safePercentage(submissionRate),
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
		timestamp, _ := time.Parse(time.RFC3339, a["timestamp"].(string))
		recentActivities = append(recentActivities, dto.ActivityLogDTO{
			UserName:  a["user_name"].(string),
			Action:    a["action"].(string),
			Timestamp: timestamp.Format("02-01-2006 15:04:05"),
		})
	}

	return &dto.AdminDashboardDTO{
		TotalStudents:    stats["totalStudents"],
		TotalTeachers:    stats["totalTeachers"],
		TotalClasses:     stats["totalClasses"],
		ActiveClasses:    stats["activeClasses"],
		EnrollmentTrends: enrollmentTrends,
		RecentActivities: recentActivities,
	}, nil
}
