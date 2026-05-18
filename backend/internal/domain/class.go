package domain

import (
	"time"

	"gorm.io/gorm"
)

type Class struct {
	ID          string         `gorm:"primaryKey;column:cls_id;default:gen_random_uuid()" json:"classId"`
	SchoolID    string         `gorm:"column:cls_sch_id;type:uuid" json:"schoolId"`
	School      School         `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	TermID      string         `gorm:"column:cls_trm_id;type:uuid" json:"termId"`
	Term        Term           `gorm:"foreignKey:TermID;references:ID" json:"term,omitempty"`
	Code        string         `gorm:"column:cls_code" json:"classCode"`
	Title       string         `gorm:"column:cls_title" json:"classTitle"`
	Description string         `gorm:"column:cls_desc" json:"classDescription"`
	CreatedBy   string         `gorm:"column:created_by;type:uuid" json:"createdBy"`
	Creator     User           `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
	IsActive    bool           `gorm:"column:is_active;default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Class) TableName() string {
	return "edv.classes"
}
