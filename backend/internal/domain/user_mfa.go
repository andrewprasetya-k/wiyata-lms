package domain

import "time"

// UserMFA is one row per user (at most) — EnabledAt nullable doubles as the
// "setup started but not confirmed yet" vs "active" flag, so no separate
// boolean column is needed. SecretEncrypted is AES-GCM ciphertext, not a
// hash — unlike every other token table in this schema, a TOTP secret must
// be decryptable again to validate a code, so hashing it (one-way) would
// make verification impossible.
type UserMFA struct {
	ID              string     `gorm:"primaryKey;column:umf_id;default:gen_random_uuid()" json:"id"`
	UserID          string     `gorm:"column:umf_usr_id;type:uuid" json:"userId"`
	SecretEncrypted string     `gorm:"column:umf_secret_encrypted" json:"-"`
	EnabledAt       *time.Time `gorm:"column:umf_enabled_at" json:"enabledAt,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (UserMFA) TableName() string {
	return "edv.user_mfa"
}
