package domain

import (
	"time"
)

type SourceType string

const (
	SourceMaterial   SourceType = "material"
	SourceAssignment SourceType = "assignment"
	SourceFeed       SourceType = "feed"
	SourceSubmission SourceType = "submission"
	SourceComment    SourceType = "comment"
)

type Attachment struct {
	ID         string     `gorm:"primaryKey;column:att_id;default:gen_random_uuid()" json:"attachmentId"`
	SchoolID   string     `gorm:"column:att_sch_id;type:uuid" json:"schoolId"`
	School     School     `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	SourceID   string     `gorm:"column:att_source_id;type:uuid;index:idx_attachments_source,priority:2" json:"sourceId"`
	SourceType SourceType `gorm:"column:att_source_type;type:source_type;index:idx_attachments_source,priority:1" json:"sourceType"`
	MediaID    string     `gorm:"column:att_med_id;type:uuid" json:"mediaId"`
	Media      Media      `gorm:"foreignKey:MediaID;references:ID" json:"media,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Attachment) TableName() string {
	return "edv.attachments"
}
