package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              string         `gorm:"primaryKey;column:usr_id;default:gen_random_uuid()" json:"userId"`
	FullName        string         `gorm:"column:usr_nama_lengkap" json:"fullName"`
	Email           string         `gorm:"column:usr_email;not null" json:"email"`
	Password        string         `gorm:"column:usr_password" json:"-"` // Hidden from JSON
	IsActive        bool           `gorm:"column:is_active;default:true" json:"isActive"`
	EmailVerifiedAt *time.Time     `gorm:"column:usr_email_verified_at" json:"emailVerifiedAt,omitempty"`
	// MFAGraceStartedAt: when this user's 7-day MFA grace period began —
	// set automatically on their first login after the MFA feature existed
	// (never backdated to CreatedAt or a migration date), so every user
	// gets a full, fair 7-day window regardless of account age. NULL means
	// the clock hasn't started yet (including every brand-new account,
	// which always gets a fresh window starting at its very first login).
	MFAGraceStartedAt *time.Time     `gorm:"column:usr_mfa_grace_started_at" json:"mfaGraceStartedAt,omitempty"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (User) TableName() string {
	return "edv.users"
}
