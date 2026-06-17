package domain

import (
	"time"
)

type Enrollment struct {
	ID           string     `gorm:"primaryKey;column:enr_id;default:gen_random_uuid()" json:"enrollmentId"`
	SchoolID     string     `gorm:"column:enr_sch_id;type:uuid" json:"schoolId"`
	School       School     `gorm:"foreignKey:SchoolID;references:ID" json:"school,omitempty"`
	SchoolUserID string     `gorm:"column:enr_scu_id;type:uuid" json:"schoolUserId"`
	SchoolUser   SchoolUser `gorm:"foreignKey:SchoolUserID;references:ID" json:"schoolUser,omitempty"`
	ClassID      string     `gorm:"column:enr_cls_id;type:uuid" json:"classId"`
	Class        Class      `gorm:"foreignKey:ClassID;references:ID" json:"class,omitempty"`
	Role         string     `gorm:"column:enr_role;type:class_role" json:"role"` // teacher or student
	JoinedAt     time.Time  `gorm:"column:joined_at;autoCreateTime" json:"joinedAt"`
	LeftAt       *time.Time `gorm:"column:left_at" json:"leftAt,omitempty"`
}

func (Enrollment) TableName() string {
	return "edv.enrollments"
}
