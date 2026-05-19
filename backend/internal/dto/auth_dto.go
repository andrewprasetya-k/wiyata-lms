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

type LoginResponseDTO struct {
	Token          string           `json:"token"`
	User           UserInfo         `json:"user"`
	Memberships    []MembershipInfo `json:"memberships"`
	GlobalRoles    []string         `json:"globalRoles"`
	DefaultContext *DefaultContext  `json:"defaultContext,omitempty"`
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
