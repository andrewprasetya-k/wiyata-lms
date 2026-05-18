package domain

import (
	"time"
)

type Role struct {
	ID        string    `gorm:"primaryKey;column:rol_id;default:gen_random_uuid()" json:"roleId"`
	Name      string    `gorm:"column:rol_name;unique" json:"roleName"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Role) TableName() string {
	return "edv.roles"
}
