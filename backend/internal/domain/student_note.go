package domain

import "time"

type StudentNote struct {
	ID         string    `gorm:"primaryKey;column:snt_id;default:gen_random_uuid()" json:"noteId"`
	SchoolID   string    `gorm:"column:snt_sch_id;type:uuid" json:"schoolId"`
	UserID     string    `gorm:"column:snt_usr_id;type:uuid" json:"userId"`
	MaterialID string    `gorm:"column:snt_mat_id;type:uuid" json:"materialId"`
	Content    string    `gorm:"column:snt_content" json:"content"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (StudentNote) TableName() string {
	return "edv.student_notes"
}

type StudentNoteWithMaterial struct {
	ID            string    `gorm:"column:snt_id"`
	MaterialID    string    `gorm:"column:snt_mat_id"`
	MaterialTitle string    `gorm:"column:material_title"`
	Content       string    `gorm:"column:snt_content"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}
