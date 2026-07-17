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
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (User) TableName() string {
	return "edv.users"
}
