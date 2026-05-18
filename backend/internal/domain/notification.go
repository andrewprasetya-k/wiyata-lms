package domain

import (
	"time"
)

type Notification struct {
	ID        string    `gorm:"primaryKey;column:ntf_id;default:gen_random_uuid()" json:"notificationId"`
	UserID    string    `gorm:"column:ntf_usr_id;type:uuid;not null" json:"userId"`
	Type      string    `gorm:"column:ntf_type;not null" json:"type"`
	Title     string    `gorm:"column:ntf_title;not null" json:"title"`
	Message   string    `gorm:"column:ntf_message" json:"message"`
	Link      string    `gorm:"column:ntf_link" json:"link"`
	RelatedID string    `gorm:"column:ntf_related_id;type:uuid" json:"relatedId"`
	IsRead    bool      `gorm:"column:is_read;default:false" json:"isRead"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "edv.notifications"
}

const (
	NotifAssignmentCreated = "assignment_created"
	NotifAssignmentGraded  = "assignment_graded"
	NotifCommentAdded      = "comment_added"
	NotifMaterialAdded     = "material_added"
	NotifFeedPosted        = "feed_posted"
)
