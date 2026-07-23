package domain

import "time"

// RefreshToken models an ongoing session, not a one-shot action (unlike
// EmailVerification/PasswordResetToken) — a row is expected to be read
// repeatedly and rotated many times over its family's lifetime, not
// consumed once. RevokedAt marks the session (or, on reuse-detection, the
// whole family) as ended; it does not mean "used up."
type RefreshToken struct {
	ID        string     `gorm:"primaryKey;column:rft_id;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"column:rft_usr_id;type:uuid" json:"userId"`
	TokenHash string     `gorm:"column:rft_token_hash" json:"-"`
	FamilyID  string     `gorm:"column:rft_family_id;type:uuid" json:"familyId"`
	ExpiresAt time.Time  `gorm:"column:rft_expires_at" json:"expiresAt"`
	RevokedAt *time.Time `gorm:"column:rft_revoked_at" json:"revokedAt,omitempty"`
	UserAgent *string    `gorm:"column:rft_user_agent" json:"userAgent,omitempty"`
	IPAddress *string    `gorm:"column:rft_ip_address" json:"ipAddress,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (RefreshToken) TableName() string {
	return "edv.refresh_tokens"
}
