package domain

import (
	"time"
)

type Log struct {
	ID string `gorm:"primaryKey;column:log_id;default:gen_random_uuid()" json:"logId"`
	// Nullable (Phase 10.4 correction): platform-scoped actions (e.g. RBAC
	// role management, super admin creation) have no active school, so this
	// must be able to represent "none" rather than an empty-string uuid,
	// which Postgres rejects at insert time.
	SchoolID  *string   `gorm:"column:log_sch_id;type:uuid" json:"schoolId,omitempty"`
	UserID    string    `gorm:"column:log_usr_id;type:uuid" json:"userId"`
	User      User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Action    string    `gorm:"column:log_action" json:"action"`
	Metadata  string    `gorm:"column:log_metadata;type:jsonb" json:"metadata"` // Stored as string for simplicity in basic impl
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Added by Phase 10.4 (audit log infrastructure). All nullable: rows
	// written before this migration, and any future write that doesn't need
	// a given field, leave it null rather than requiring a value.
	ActorSchoolUserID *string `gorm:"column:actor_school_user_id;type:uuid" json:"actorSchoolUserId,omitempty"`
	EntityType        *string `gorm:"column:entity_type" json:"entityType,omitempty"`
	EntityID          *string `gorm:"column:entity_id;type:uuid" json:"entityId,omitempty"`
	Scope             *string `gorm:"column:scope" json:"scope,omitempty"`
	Severity          *string `gorm:"column:severity" json:"severity,omitempty"`
	IPAddress         *string `gorm:"column:ip_address" json:"ipAddress,omitempty"`
	UserAgent         *string `gorm:"column:user_agent" json:"userAgent,omitempty"`
	CorrelationID     *string `gorm:"column:correlation_id;type:uuid" json:"correlationId,omitempty"`
}

func (Log) TableName() string {
	return "edv.logs"
}

// Log scope values. Validated in application code, not a DB constraint
// (same convention as other free-text status columns in this schema).
const (
	LogScopeSchool   = "school"
	LogScopePlatform = "platform"
)

// Log severity values, per the Phase 10.2 severity policy.
const (
	LogSeverityLow    = "LOW"
	LogSeverityMedium = "MEDIUM"
	LogSeverityHigh   = "HIGH"
)
