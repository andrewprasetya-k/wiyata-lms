package domain

import (
	"time"

	"gorm.io/gorm"
)

type AssignmentCategory struct {
	ID        string    `gorm:"primaryKey;column:asc_id;default:gen_random_uuid()" json:"categoryId"`
	SchoolID  string    `gorm:"column:asc_sch_id;type:uuid" json:"schoolId"`
	School    School    `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	Name      string    `gorm:"column:asc_name" json:"categoryName"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (AssignmentCategory) TableName() string {
	return "edv.assignment_categories"
}

type Assignment struct {
	ID                  string             `gorm:"primaryKey;column:asg_id;default:gen_random_uuid()" json:"assignmentId"`
	SchoolID            string             `gorm:"column:asg_sch_id;type:uuid" json:"schoolId"`
	SubjectClassID      string             `gorm:"column:asg_scl_id;type:uuid" json:"subjectClassId"`
	SubjectClass        SubjectClass       `gorm:"foreignKey:SubjectClassID;references:ID" json:"subjectClass,omitempty"`
	CategoryID          string             `gorm:"column:asg_asc_id;type:uuid" json:"categoryId"`
	Category            AssignmentCategory `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
	Title               string             `gorm:"column:asg_title" json:"assignmentTitle"`
	Description         string             `gorm:"column:asg_desc" json:"assignmentDescription"`
	Deadline            *time.Time         `gorm:"column:asg_deadline" json:"deadline"`
	AllowLateSubmission bool               `gorm:"column:asg_allowed_late;default:true" json:"allowLateSubmission"`
	CreatedBy           string             `gorm:"column:created_by;type:uuid" json:"createdBy"`
	Creator             User               `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
	CreatedAt           time.Time          `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt           gorm.DeletedAt     `gorm:"column:deleted_at;index" json:"-"`
	Attachments         []Attachment       `gorm:"-" json:"attachments,omitempty"`
	Submissions         []Submission       `gorm:"foreignKey:AssignmentID" json:"submissions,omitempty"`
}

func (Assignment) TableName() string {
	return "edv.assignments"
}

type Submission struct {
	ID           string         `gorm:"primaryKey;column:sbm_id;default:gen_random_uuid()" json:"submissionId"`
	SchoolID     string         `gorm:"column:sbm_sch_id;type:uuid" json:"schoolId"`
	AssignmentID string         `gorm:"column:sbm_asg_id;type:uuid" json:"assignmentId"`
	Assignment   Assignment     `gorm:"foreignKey:AssignmentID;references:ID" json:"assignment,omitempty"`
	UserID       string         `gorm:"column:sbm_usr_id;type:uuid" json:"userId"`
	User         User           `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SubmittedAt  time.Time      `gorm:"column:submitted_at;autoCreateTime" json:"submittedAt"`
	IsLate       bool           `gorm:"-" json:"isLate"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
	Attachments  []Attachment   `gorm:"-" json:"attachments,omitempty"`
	Assessment   *Assessment    `gorm:"foreignKey:SubmissionID" json:"assessment,omitempty"`
}

func (Submission) TableName() string {
	return "edv.submissions"
}

type Assessment struct {
	ID           string     `gorm:"primaryKey;column:asm_id;default:gen_random_uuid()" json:"assessmentId"`
	SubmissionID string     `gorm:"column:asm_sbm_id;type:uuid" json:"submissionId"`
	Submission   Submission `gorm:"foreignKey:SubmissionID;references:ID" json:"submission,omitempty"` // TAMBAH INI
	Score        float64    `gorm:"column:asm_score" json:"score"`
	Feedback     string     `gorm:"column:asm_feedback" json:"feedback"`
	AssessedBy   string     `gorm:"column:assessed_by;type:uuid" json:"assessedBy"`
	Assessor     User       `gorm:"foreignKey:AssessedBy;references:ID" json:"assessor,omitempty"`
	AssessedAt   time.Time  `gorm:"column:assessed_at;autoCreateTime" json:"assessedAt"`
}

func (Assessment) TableName() string {
	return "edv.assessments"
}

type AssessmentWeight struct {
	ID         string  `gorm:"primaryKey;column:asw_id;default:gen_random_uuid()" json:"weightId"`
	SubjectID  string  `gorm:"column:asw_sub_id;type:uuid" json:"subjectId"`
	CategoryID string  `gorm:"column:asw_asc_id;type:uuid" json:"categoryId"`
	Weight     float64 `gorm:"column:asw_weight" json:"weight"`

	Subject  Subject            `gorm:"foreignKey:SubjectID;references:ID" json:"subject,omitempty"`
	Category AssignmentCategory `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

func (AssessmentWeight) TableName() string {
	return "edv.assessments_weights"
}
