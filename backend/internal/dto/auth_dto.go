package dto

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterDTO struct {
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponseDTO.Token is the short-lived (15 min) access token as of the
// refresh-token migration — it was a 24h token before. The long-lived
// refresh token is never included here; it's set as an httpOnly cookie by
// the handler instead. Only ever returned once a login is fully complete —
// see LoginChallengeDTO for the "one more step required" shape.
type LoginResponseDTO struct {
	Token          string           `json:"token"`
	User           UserInfo         `json:"user"`
	Memberships    []MembershipInfo `json:"memberships"`
	GlobalRoles    []string         `json:"globalRoles"`
	DefaultContext *DefaultContext  `json:"defaultContext,omitempty"`
	// MFAGraceDaysRemaining is set only when the user hasn't enrolled in
	// MFA yet but is still within the grace period — a reminder for the
	// frontend to show, never a blocker (nil once MFA is enabled, or once
	// enrollment is enforced via LoginChallengeDTO instead).
	MFAGraceDaysRemaining *int `json:"mfaGraceDaysRemaining,omitempty"`
}

// LoginChallengeDTO is what POST /login (and POST /register's auto-login)
// return instead of LoginResponseDTO when the password was correct but a
// second step still stands between here and a completed login. Exactly one
// of MFARequired/MFASetupRequired is true. No access/refresh token is
// issued yet — PreAuthToken is a short-lived, single-use, unrelated token
// that only proves the password step already succeeded.
type LoginChallengeDTO struct {
	// MFARequired: the user has MFA enabled — submit a code to
	// POST /login/mfa-verify.
	MFARequired bool `json:"mfaRequired,omitempty"`
	// MFASetupRequired: the user's MFA grace period has elapsed with
	// nothing enrolled — they must complete enrollment before a real
	// session can start.
	MFASetupRequired bool   `json:"mfaSetupRequired,omitempty"`
	PreAuthToken     string `json:"preAuthToken"`
}

// RefreshTokenResponseDTO is POST /refresh-token's response body — only the
// new access token. The rotated refresh token is set as a cookie, never
// returned here.
type RefreshTokenResponseDTO struct {
	AccessToken string `json:"accessToken"`
}

// LogoutResponseDTO is POST /logout's response body.
type LogoutResponseDTO struct {
	Message string `json:"message"`
}

// WSTicketResponseDTO is GET /me/ws-ticket's response body — a short-lived,
// single-use ticket for the WebSocket/SSE handshake (?ticket=...), used
// instead of the raw access token now that it may live somewhere JS can't
// read it from.
type WSTicketResponseDTO struct {
	Ticket string `json:"ticket"`
}

// SessionDTO is one row of GET /me/sessions — the token_hash itself is
// never included. UserAgent is returned raw (not parsed server-side — no
// user-agent-parsing dependency exists in go.mod, and adding one for this
// alone wasn't worth it); the frontend parses it into a friendly
// device/browser summary.
type SessionDTO struct {
	ID         string `json:"id"`
	LoggedInAt string `json:"loggedInAt"`
	ExpiresAt  string `json:"expiresAt"`
	UserAgent  string `json:"userAgent"`
	IPAddress  string `json:"ipAddress"`
}

type AuthContextResponseDTO struct {
	Memberships     []MembershipInfo `json:"memberships"`
	GlobalRoles     []string         `json:"globalRoles"`
	DefaultContext  *DefaultContext  `json:"defaultContext,omitempty"`
	EmailVerified   bool             `json:"emailVerified"`
	EmailVerifiedAt *string          `json:"emailVerifiedAt,omitempty"`
}

type UserInfo struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type SchoolInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type MembershipInfo struct {
	SchoolUserID string     `json:"schoolUserId"`
	School       SchoolInfo `json:"school"`
	Roles        []string   `json:"roles"`
	IsDefault    bool       `json:"isDefault"`
}

type DefaultContext struct {
	SchoolID     string   `json:"schoolId"`
	SchoolUserID string   `json:"schoolUserId"`
	Roles        []string `json:"roles"`
}
