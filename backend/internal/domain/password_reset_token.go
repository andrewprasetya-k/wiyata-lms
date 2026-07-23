package domain

import "time"

type PasswordResetToken struct {
	ID         string     `gorm:"primaryKey;column:prt_id;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"column:prt_usr_id;type:uuid" json:"userId"`
	TokenHash  string     `gorm:"column:prt_token_hash" json:"-"`
	ExpiresAt  time.Time  `gorm:"column:prt_expires_at" json:"expiresAt"`
	ConsumedAt *time.Time `gorm:"column:prt_consumed_at" json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (PasswordResetToken) TableName() string {
	return "edv.password_reset_tokens"
}
