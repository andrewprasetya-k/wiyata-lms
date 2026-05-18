package domain

import (
	"time"

	"gorm.io/gorm"
)

type School struct {
	ID        string         `gorm:"primaryKey;column:sch_id;default:gen_random_uuid()" json:"schoolId"`
	Name      string         `gorm:"column:sch_name" json:"schoolName"`
	Code      string         `gorm:"column:sch_code;unique" json:"schoolCode"`
	LogoID    *string        `gorm:"column:sch_logo;type:uuid" json:"logoId,omitempty"`
	Address   string         `gorm:"column:sch_address" json:"schoolAddress"`
	Email     string         `gorm:"column:sch_email" json:"schoolEmail"`
	Phone     string         `gorm:"column:sch_phone" json:"schoolPhone"`
	Website   *string        `gorm:"column:sch_website" json:"schoolWebsite,omitempty"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (School) TableName() string {
	return "edv.schools"
}
