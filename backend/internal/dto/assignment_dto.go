package dto

import (
	"time"
)

// Category
type CreateAssignmentCategoryDTO struct {
	SchoolID string `json:"schoolId" binding:"required,uuid"`
	Name     string `json:"categoryName" binding:"required"`
}

type AssignmentCategoryResponseDTO struct {
	ID        string `json:"categoryId"`
	SchoolID  string `json:"schoolId"`
	Name      string `json:"categoryName"`
	CreatedAt string `json:"createdAt"`
}

type SchoolWithAssignmentCategoriesDTO struct {
	School     SchoolHeaderDTO                 `json:"school"`
	Categories []AssignmentCategoryResponseDTO `json:"categories"`
}

// Assignment
type CreateAssignmentDTO struct {
	SchoolID            string     `json:"schoolId" binding:"required,uuid"`
	SubjectClassID      string     `json:"subjectClassId" binding:"required,uuid"`
	CategoryID          string     `json:"categoryId" binding:"required,uuid"`
	Title               string     `json:"assignmentTitle" binding:"required"`
	Description         string     `json:"assignmentDescription"`
	Deadline            *time.Time `json:"deadline"`
	AllowLateSubmission bool       `json:"allowLateSubmission"`
	MediaIDs            []string   `json:"mediaIds"`
}

type UpdateAssignmentDTO struct {
	CategoryID          *string    `json:"categoryId" binding:"omitempty,uuid"`
	Title               *string    `json:"assignmentTitle"`
	Description         *string    `json:"assignmentDescription"`
	Deadline            *time.Time `json:"deadline"`
	AllowLateSubmission *bool      `json:"allowLateSubmission"`
	MediaIDs            []string   `json:"mediaIds"`
}

type AssignmentResponseDTO struct {
	ID                  string             `json:"assignmentId"`
	Title               string             `json:"assignmentTitle"`
	Description         string             `json:"assignmentDescription"`
	CategoryName        string             `json:"categoryName"`
	Deadline            *time.Time         `json:"deadline,omitempty"`
	AllowLateSubmission bool               `json:"allowLateSubmission"`
	CreatedAt           string             `json:"createdAt"`
	Attachments         []MediaResponseDTO `json:"attachments,omitempty"`
}

type AssignmentPerSubjectClassResponseDTO struct {
	SubjectClass SubjectClassHeaderDTO `json:"subjectClass"`
	Data         PaginatedResponse     `json:"data"`
}

type AssignmentHeaderDTO struct {
	ID           string     `json:"assignmentId"`
	Title        string     `json:"assignmentTitle"`
	SubjectName  string     `json:"subjectName"`
	CategoryName string     `json:"categoryName"`
	Deadline     *time.Time `json:"deadline,omitempty"`
}

type AssignmentWithSubmissionsDTO struct {
	Assignment  AssignmentHeaderDTO     `json:"assignment"`
	Submissions []SubmissionResponseDTO `json:"submissions"`
}

type AssignmentSubmissionGroupDTO struct {
	Assignment      AssignmentHeaderDTO     `json:"assignment"`
	SubmissionCount int                     `json:"submissionCount"`
	GradedCount     int                     `json:"gradedCount"`
	PendingCount    int                     `json:"pendingCount"`
	Submissions     []SubmissionResponseDTO `json:"submissions"`
}

type SubjectClassSubmissionSummaryDTO struct {
	AssignmentCount int `json:"assignmentCount"`
	SubmissionCount int `json:"submissionCount"`
	GradedCount     int `json:"gradedCount"`
	PendingCount    int `json:"pendingCount"`
	LateCount       int `json:"lateCount"`
}

type SubjectClassSubmissionsResponseDTO struct {
	SubjectClass SubjectClassHeaderDTO            `json:"subjectClass"`
	Assignments  []AssignmentSubmissionGroupDTO   `json:"assignments"`
	Summary      SubjectClassSubmissionSummaryDTO `json:"summary"`
}

type TeacherSubmissionInboxSummaryDTO struct {
	TotalSubmissions int `json:"totalSubmissions"`
	PendingCount     int `json:"pendingCount"`
	GradedCount      int `json:"gradedCount"`
	LateCount        int `json:"lateCount"`
}

type TeacherSubmissionInboxItemDTO struct {
	AssignmentID    string     `json:"assignmentId" gorm:"column:assignment_id"`
	SubjectClassID  string     `json:"subjectClassId" gorm:"column:subject_class_id"`
	AssignmentTitle string     `json:"assignmentTitle" gorm:"column:assignment_title"`
	SubjectName     string     `json:"subjectName" gorm:"column:subject_name"`
	SubjectCode     string     `json:"subjectCode" gorm:"column:subject_code"`
	ClassName       string     `json:"className" gorm:"column:class_name"`
	ClassCode       string     `json:"classCode" gorm:"column:class_code"`
	Deadline        *time.Time `json:"deadline" gorm:"column:deadline"`
	SubmissionCount int        `json:"submissionCount" gorm:"column:submission_count"`
	PendingCount    int        `json:"pendingCount" gorm:"column:pending_count"`
	GradedCount     int        `json:"gradedCount" gorm:"column:graded_count"`
	LateCount       int        `json:"lateCount" gorm:"column:late_count"`
}

type TeacherSubmissionInboxResponseDTO struct {
	Summary TeacherSubmissionInboxSummaryDTO `json:"summary"`
	Items   []TeacherSubmissionInboxItemDTO  `json:"items"`
}

// Submission
type CreateSubmissionDTO struct {
	SchoolID string   `json:"schoolId" binding:"required,uuid"`
	MediaIDs []string `json:"mediaIds"`
}

type SubmissionResponseDTO struct {
	ID          string                 `json:"submissionId"`
	UserName    string                 `json:"studentName"`
	SubmittedAt string                 `json:"submittedAt"`
	IsLate      bool                   `json:"isLate"`
	Attachments []MediaResponseDTO     `json:"attachments,omitempty"`
	Assessment  *AssessmentResponseDTO `json:"assessment,omitempty"`
}

// Assessment
type CreateAssessmentDTO struct {
	Score    float64 `json:"score" binding:"required"`
	Feedback string  `json:"feedback"`
}

type UpdateAssessmentDTO struct {
	Score    *float64 `json:"score"`
	Feedback *string  `json:"feedback"`
}

type AssessmentResponseDTO struct {
	Score      float64 `json:"score"`
	Feedback   string  `json:"feedback"`
	Assessor   string  `json:"assessorName"`
	AssessedAt string  `json:"assessedAt"`
}

type MySubmissionAssessmentDTO struct {
	ID           string  `json:"assessmentId"`
	Score        float64 `json:"score"`
	Feedback     string  `json:"feedback"`
	AssessedAt   string  `json:"assessedAt"`
	AssessorName string  `json:"assessorName"`
}

type MySubmissionDTO struct {
	ID           string                     `json:"submissionId"`
	AssignmentID string                     `json:"assignmentId"`
	SubmittedAt  string                     `json:"submittedAt"`
	Attachments  []MediaResponseDTO         `json:"attachments,omitempty"`
	Assessment   *MySubmissionAssessmentDTO `json:"assessment"`
}

type MySubmissionResponseDTO struct {
	Status     string           `json:"status"`
	Submission *MySubmissionDTO `json:"submission"`
}

// Weight
type SetAssessmentWeightDTO struct {
	SubjectID  string  `json:"subjectId" binding:"required,uuid"`
	CategoryID string  `json:"categoryId" binding:"required,uuid"`
	Weight     float64 `json:"weight" binding:"required"`
}

type WeightItemDTO struct {
	CategoryID string  `json:"categoryId" binding:"required,uuid"`
	Weight     float64 `json:"weight" binding:"required,min=0,max=100"`
}

type WeightResponseDTO struct {
	SubjectID   string            `json:"subjectId"`
	SubjectName string            `json:"subjectName"`
	SubjectCode string            `json:"subjectCode"`
	Weights     []WeightDetailDTO `json:"weights"`
	TotalWeight float64           `json:"totalWeight"`
}

type WeightDetailDTO struct {
	WeightID     string  `json:"weightId"`
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Weight       float64 `json:"weight"`
}

type GradeReportDTO struct {
	StudentID   string                 `json:"studentId"`
	StudentName string                 `json:"studentName"`
	SubjectID   string                 `json:"subjectId"`
	SubjectName string                 `json:"subjectName"`
	Breakdown   []CategoryBreakdownDTO `json:"breakdown"`
	FinalGrade  float64                `json:"finalGrade"`
	LetterGrade string                 `json:"letterGrade"`
}

type ConfigureWeightsDTO struct {
	SubjectID string                   `json:"subjectId" binding:"required,uuid"`
	Weights   []SetAssessmentWeightDTO `json:"weights" binding:"required,dive"`
}

type CategoryBreakdownDTO struct {
	CategoryID      string  `json:"categoryId"`
	CategoryName    string  `json:"categoryName"`
	AverageScore    float64 `json:"averageScore"`
	WeightedScore   float64 `json:"weightedScore"`
	Weight          float64 `json:"weight"`
	AssignmentCount int     `json:"assignmentCount"`
}

type ClassGradeReportDTO struct {
	Class    ClassHeaderDTO           `json:"class"`
	Subject  SubjectHeaderDTO         `json:"subject"`
	Students []StudentGradeSummaryDTO `json:"students"`
}

type StudentGradeSummaryDTO struct {
	StudentID    string  `json:"studentId"`
	StudentName  string  `json:"studentName"`
	StudentEmail string  `json:"studentEmail"`
	FinalGrade   float64 `json:"finalGrade"`
	LetterGrade  string  `json:"letterGrade"`
}

type MyGradebookResponseDTO struct {
	Class    MyGradebookClassDTO     `json:"class"`
	Subjects []MyGradebookSubjectDTO `json:"subjects"`
	Summary  MyGradebookSummaryDTO   `json:"summary"`
}

type MyGradebookClassDTO struct {
	ClassID   string `json:"classId"`
	ClassName string `json:"className"`
	ClassCode string `json:"classCode"`
}

type MyGradebookSubjectDTO struct {
	SubjectClassID string                     `json:"subjectClassId"`
	SubjectID      string                     `json:"subjectId"`
	SubjectName    string                     `json:"subjectName"`
	SubjectCode    string                     `json:"subjectCode"`
	FinalGrade     *float64                   `json:"finalGrade"`
	LetterGrade    *string                    `json:"letterGrade"`
	GradedCount    int                        `json:"gradedCount"`
	SubmittedCount int                        `json:"submittedCount"`
	PendingCount   int                        `json:"pendingCount"`
	Assignments    []MyGradebookAssignmentDTO `json:"assignments"`
}

type MyGradebookAssignmentDTO struct {
	AssignmentID    string     `json:"assignmentId"`
	AssignmentTitle string     `json:"assignmentTitle"`
	CategoryName    string     `json:"categoryName"`
	Deadline        *time.Time `json:"deadline,omitempty"`
	Status          string     `json:"status"`
	SubmittedAt     *string    `json:"submittedAt"`
	Score           *float64   `json:"score"`
	Feedback        *string    `json:"feedback"`
	AssessedAt      *string    `json:"assessedAt"`
	AssessorName    *string    `json:"assessorName"`
}

type MyGradebookSummaryDTO struct {
	SubjectCount             int `json:"subjectCount"`
	GradedAssignmentCount    int `json:"gradedAssignmentCount"`
	SubmittedAssignmentCount int `json:"submittedAssignmentCount"`
	PendingAssessmentCount   int `json:"pendingAssessmentCount"`
}

type StudentGradebookClassRow struct {
	ClassID   string `gorm:"column:class_id"`
	ClassName string `gorm:"column:class_name"`
	ClassCode string `gorm:"column:class_code"`
}

type StudentGradebookRow struct {
	SubjectClassID  string     `gorm:"column:subject_class_id"`
	SubjectID       string     `gorm:"column:subject_id"`
	SubjectName     string     `gorm:"column:subject_name"`
	SubjectCode     string     `gorm:"column:subject_code"`
	AssignmentID    *string    `gorm:"column:assignment_id"`
	AssignmentTitle *string    `gorm:"column:assignment_title"`
	CategoryID      *string    `gorm:"column:category_id"`
	CategoryName    *string    `gorm:"column:category_name"`
	Deadline        *time.Time `gorm:"column:deadline"`
	SubmissionID    *string    `gorm:"column:submission_id"`
	SubmittedAt     *time.Time `gorm:"column:submitted_at"`
	Score           *float64   `gorm:"column:score"`
	Feedback        *string    `gorm:"column:feedback"`
	AssessedAt      *time.Time `gorm:"column:assessed_at"`
	AssessorName    *string    `gorm:"column:assessor_name"`
}

type SubjectHeaderDTO struct {
	SubjectID   string `json:"subjectId"`
	SubjectName string `json:"subjectName"`
	SubjectCode string `json:"subjectCode"`
}
