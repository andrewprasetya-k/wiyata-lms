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
// the handler instead.
type LoginResponseDTO struct {
	Token          string           `json:"token"`
	User           UserInfo         `json:"user"`
	Memberships    []MembershipInfo `json:"memberships"`
	GlobalRoles    []string         `json:"globalRoles"`
	DefaultContext *DefaultContext  `json:"defaultContext,omitempty"`
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
