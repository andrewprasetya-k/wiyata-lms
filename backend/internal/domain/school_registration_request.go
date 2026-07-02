package domain

import "time"

type SchoolRegistrationRequestStatus string

const (
	SchoolRegistrationPending  SchoolRegistrationRequestStatus = "pending"
	SchoolRegistrationApproved SchoolRegistrationRequestStatus = "approved"
	SchoolRegistrationRejected SchoolRegistrationRequestStatus = "rejected"
)

type SchoolRegistrationRequest struct {
	ID         string                          `gorm:"primaryKey;column:srr_id;default:gen_random_uuid()" json:"requestId"`
	SchoolName string                          `gorm:"column:srr_school_name" json:"schoolName"`
	NPSN       *string                         `gorm:"column:srr_npsn" json:"npsn,omitempty"`
	PICName    string                          `gorm:"column:srr_pic_name" json:"picName"`
	PICEmail   string                          `gorm:"column:srr_pic_email" json:"picEmail"`
	PICPhone   *string                         `gorm:"column:srr_pic_phone" json:"picPhone,omitempty"`
	PICRole    *string                         `gorm:"column:srr_pic_role" json:"picRole,omitempty"`
	Message    *string                         `gorm:"column:srr_message" json:"message,omitempty"`
	Status     SchoolRegistrationRequestStatus `gorm:"column:srr_status" json:"status"`
	ReviewedBy *string                         `gorm:"column:srr_reviewed_by" json:"reviewedBy,omitempty"`
	ReviewedAt *time.Time                      `gorm:"column:srr_reviewed_at" json:"reviewedAt,omitempty"`
	ReviewNote *string                         `gorm:"column:srr_review_note" json:"reviewNote,omitempty"`
	CreatedAt  time.Time                       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time                       `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (SchoolRegistrationRequest) TableName() string {
	return "edv.school_registration_requests"
}
