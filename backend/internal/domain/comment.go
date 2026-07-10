package domain

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID         string         `gorm:"primaryKey;column:cmn_id;default:gen_random_uuid()" json:"commentId"`
	SchoolID   string         `gorm:"column:cmn_sch_id;type:uuid" json:"schoolId"`
	SourceType SourceType     `gorm:"column:cmn_source_type;type:source_type;index:idx_comments_source,priority:1" json:"sourceType"`
	SourceID   string         `gorm:"column:cmn_source_id;type:uuid;index:idx_comments_source,priority:2" json:"sourceId"`
	UserID     string         `gorm:"column:cmn_usr_id;type:uuid" json:"userId"`
	User       User           `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Content    string         `gorm:"column:cmn_content" json:"content"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Comment) TableName() string {
	return "edv.comments"
}
