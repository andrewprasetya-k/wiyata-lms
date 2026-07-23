package domain

import "time"

// MFA pre-auth token purposes — see MFAPreAuthToken.
const (
	// MFAPreAuthPurposeVerify: the user has MFA enabled and must submit a
	// TOTP/recovery code before real tokens are issued.
	MFAPreAuthPurposeVerify = "mfa_verify"
	// MFAPreAuthPurposeEnrollRequired: the user's MFA grace period has
	// elapsed with no MFA enrolled — they must complete enrollment before
	// real tokens are issued. Kept distinct from MFAPreAuthPurposeVerify so
	// /login/mfa-verify can reject a token issued for the wrong purpose
	// (there's no code to verify yet for a user who hasn't enrolled).
	MFAPreAuthPurposeEnrollRequired = "mfa_enroll_required"
)

// MFAPreAuthToken proves "this device already presented the correct
// password for this user" without being a real access/refresh token —
// issued by AuthService.Login in place of real tokens whenever a second
// step (MFA verification, or forced enrollment) still stands between the
// password check and a completed login.
type MFAPreAuthToken struct {
	ID         string     `gorm:"primaryKey;column:mpt_id;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"column:mpt_usr_id;type:uuid" json:"userId"`
	TokenHash  string     `gorm:"column:mpt_token_hash" json:"-"`
	Purpose    string     `gorm:"column:mpt_purpose" json:"purpose"`
	ExpiresAt  time.Time  `gorm:"column:mpt_expires_at" json:"expiresAt"`
	ConsumedAt *time.Time `gorm:"column:mpt_consumed_at" json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (MFAPreAuthToken) TableName() string {
	return "edv.mfa_preauth_tokens"
}
