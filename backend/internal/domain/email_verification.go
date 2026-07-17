package domain

import "time"

type EmailVerification struct {
	ID         string     `gorm:"primaryKey;column:evf_id;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"column:evf_usr_id;type:uuid" json:"userId"`
	TokenHash  string     `gorm:"column:evf_token_hash" json:"-"`
	ExpiresAt  time.Time  `gorm:"column:evf_expires_at" json:"expiresAt"`
	ConsumedAt *time.Time `gorm:"column:evf_consumed_at" json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (EmailVerification) TableName() string {
	return "edv.email_verifications"
}
