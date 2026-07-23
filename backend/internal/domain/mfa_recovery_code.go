package domain

import "time"

// MFARecoveryCode is one row per recovery code — single-use like an
// invitation/verification token, so it's hashed (SHA-256), never encrypted;
// there's no need to ever recover the plaintext, only to check whether a
// presented code matches an unconsumed hash.
type MFARecoveryCode struct {
	ID         string     `gorm:"primaryKey;column:mrc_id;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"column:mrc_usr_id;type:uuid" json:"userId"`
	CodeHash   string     `gorm:"column:mrc_code_hash" json:"-"`
	ConsumedAt *time.Time `gorm:"column:mrc_consumed_at" json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (MFARecoveryCode) TableName() string {
	return "edv.mfa_recovery_codes"
}
