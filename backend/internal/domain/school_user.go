package domain

import (
	"time"

	"gorm.io/gorm"
)

type SchoolUser struct {
	ID        string         `gorm:"primaryKey;column:scu_id;default:gen_random_uuid()" json:"schoolUserId"`
	UserID    string         `gorm:"column:scu_usr_id;type:uuid" json:"userId"`
	User      User           `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SchoolID  string         `gorm:"column:scu_sch_id;type:uuid" json:"schoolId"`
	School    School         `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	Roles     []UserRole     `gorm:"foreignKey:SchoolUserID;references:ID" json:"roles,omitempty"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (SchoolUser) TableName() string {
	return "edv.school_users"
}
