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

type SubjectHeaderDTO struct {
	SubjectID   string `json:"subjectId"`
	SubjectName string `json:"subjectName"`
	SubjectCode string `json:"subjectCode"`
}
