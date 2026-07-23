package domain

import "time"

// RefreshToken models an ongoing session, not a one-shot action (unlike
// EmailVerification/PasswordResetToken) — a row is expected to be read
// repeatedly and rotated many times over its family's lifetime, not
// consumed once. RevokedAt marks the session (or, on reuse-detection, the
// whole family) as ended; it does not mean "used up." RevokedReason records
// *why* — Rotate() uses it to tell a deliberate end-of-session (user-revoked
// session, logout) apart from an actual token-reuse signal, since both look
// identical as just "found a row with RevokedAt set" otherwise.
type RefreshToken struct {
	ID            string     `gorm:"primaryKey;column:rft_id;default:gen_random_uuid()" json:"id"`
	UserID        string     `gorm:"column:rft_usr_id;type:uuid" json:"userId"`
	TokenHash     string     `gorm:"column:rft_token_hash" json:"-"`
	FamilyID      string     `gorm:"column:rft_family_id;type:uuid" json:"familyId"`
	ExpiresAt     time.Time  `gorm:"column:rft_expires_at" json:"expiresAt"`
	RevokedAt     *time.Time `gorm:"column:rft_revoked_at" json:"revokedAt,omitempty"`
	RevokedReason *string    `gorm:"column:rft_revoked_reason" json:"revokedReason,omitempty"`
	UserAgent     *string    `gorm:"column:rft_user_agent" json:"userAgent,omitempty"`
	IPAddress     *string    `gorm:"column:rft_ip_address" json:"ipAddress,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (RefreshToken) TableName() string {
	return "edv.refresh_tokens"
}

// Refresh token revocation reasons — see RefreshToken.RevokedReason.
const (
	// RefreshTokenRevokedReasonRotated marks the old token in a normal
	// Rotate() call — expected, not a reuse signal on its own (it only
	// becomes one if that same old token is presented again afterward).
	RefreshTokenRevokedReasonRotated = "rotated"
	// RefreshTokenRevokedReasonUserRevoked marks a session the user ended
	// deliberately via DELETE /me/sessions/:id.
	RefreshTokenRevokedReasonUserRevoked = "user_revoked"
	// RefreshTokenRevokedReasonLogout marks a session ended via POST /logout.
	RefreshTokenRevokedReasonLogout = "logout"
	// RefreshTokenRevokedReasonReuseDetected marks every token in a family
	// that RevokeFamily ended in response to a confirmed reuse/replay.
	RefreshTokenRevokedReasonReuseDetected = "reuse_detected"
)
