package domain

import "time"

type SchoolRegistrationRequestStatus string

const (
	SchoolRegistrationPending  SchoolRegistrationRequestStatus = "pending"
	SchoolRegistrationApproved SchoolRegistrationRequestStatus = "approved"
	SchoolRegistrationRejected SchoolRegistrationRequestStatus = "rejected"
)

type SchoolRegistrationRequest struct {
	ID string `gorm:"primaryKey;column:srr_id;default:gen_random_uuid()" json:"requestId"`
	// TODO(db-migration): srr_usr_id does not exist in the database yet.
	// A migration adding "srr_usr_id uuid" (FK -> edv.users.usr_id, nullable
	// for backward compatibility with pre-existing rows) must be applied
	// before this field will actually persist. Until then, Create() will
	// fail at the database level because the column is missing.
	RequesterUserID string                          `gorm:"column:srr_usr_id" json:"requesterUserId,omitempty"`
	SchoolName      string                          `gorm:"column:srr_school_name" json:"schoolName"`
	NPSN            *string                         `gorm:"column:srr_npsn" json:"npsn,omitempty"`
	PICName         string                          `gorm:"column:srr_pic_name" json:"picName"`
	PICEmail        string                          `gorm:"column:srr_pic_email" json:"picEmail"`
	PICPhone        *string                         `gorm:"column:srr_pic_phone" json:"picPhone,omitempty"`
	PICRole         *string                         `gorm:"column:srr_pic_role" json:"picRole,omitempty"`
	Message         *string                         `gorm:"column:srr_message" json:"message,omitempty"`
	Status          SchoolRegistrationRequestStatus `gorm:"column:srr_status" json:"status"`
	ReviewedBy      *string                         `gorm:"column:srr_reviewed_by" json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time                      `gorm:"column:srr_reviewed_at" json:"reviewedAt,omitempty"`
	ReviewNote      *string                         `gorm:"column:srr_review_note" json:"reviewNote,omitempty"`
	CreatedAt       time.Time                       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time                       `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (SchoolRegistrationRequest) TableName() string {
	return "edv.school_registration_requests"
}
