package dto

// Student Dashboard
type StudentDashboardDTO struct {
	PendingAssignments int                     `json:"pendingAssignments"`
	UpcomingDeadlines  []AssignmentDeadlineDTO `json:"upcomingDeadlines"`
	AverageScore       float64                 `json:"averageScore"`
	CompletedMaterials int                     `json:"completedMaterials"`
	TotalMaterials     int                     `json:"totalMaterials"`
}

type AssignmentDeadlineDTO struct {
	AssignmentID    string `json:"assignmentId"`
	AssignmentTitle string `json:"assignmentTitle"`
	SubjectName     string `json:"subjectName"`
	Deadline        string `json:"deadline"`
	IsSubmitted     bool   `json:"isSubmitted"`
}

// Teacher Dashboard
type TeacherDashboardDTO struct {
	PendingReviews   int                   `json:"pendingReviews"`
	ClassPerformance []ClassPerformanceDTO `json:"classPerformance"`
}

type ClassPerformanceDTO struct {
	ClassID        string  `json:"classId"`
	ClassName      string  `json:"className"`
	SubjectName    string  `json:"subjectName"`
	SubjectColor   string  `json:"subjectColor,omitempty"`
	AverageScore   float64 `json:"averageScore"`
	SubmissionRate float64 `json:"submissionRate"`
	TotalStudents  int     `json:"totalStudents"`
}

// Admin Dashboard
type AdminDashboardDTO struct {
	TotalStudents    int                  `json:"totalStudents"`
	TotalTeachers    int                  `json:"totalTeachers"`
	TotalClasses     int                  `json:"totalClasses"`
	ActiveClasses    int                  `json:"activeClasses"`
	EnrollmentTrends []EnrollmentTrendDTO `json:"enrollmentTrends"`
	RecentActivities []ActivityLogDTO     `json:"recentActivities"`
}

type EnrollmentTrendDTO struct {
	ClassName     string `json:"className"`
	TotalEnrolled int    `json:"totalEnrolled"`
	Teachers      int    `json:"teachers"`
	Students      int    `json:"students"`
}

type ActivityLogDTO struct {
	UserName  string `json:"userName"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}
