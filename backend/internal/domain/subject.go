package domain

import (
	"time"
)

type Subject struct {
	ID        string    `gorm:"primaryKey;column:sub_id;default:gen_random_uuid()" json:"subjectId"`
	SchoolID  string    `gorm:"column:sub_sch_id;type:uuid" json:"schoolId"`
	School    School    `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	Name      string    `gorm:"column:sub_name" json:"subjectName"`
	Code      string    `gorm:"column:sub_code" json:"subjectCode"`
	Color     string    `gorm:"column:sub_color;not null" json:"color,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (Subject) TableName() string {
	return "edv.subjects"
}
