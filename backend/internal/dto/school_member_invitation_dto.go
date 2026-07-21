package dto

type CreateSchoolMemberInvitationDTO struct {
	FullName  string `json:"fullName,omitempty"`
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role" binding:"required"`
	ClassCode string `json:"classCode,omitempty"`
}

type InvitationClassDTO struct {
	ClassID    string `json:"classId"`
	ClassCode  string `json:"classCode"`
	ClassTitle string `json:"classTitle"`
}

type SchoolMemberInvitationDTO struct {
	InvitationID string              `json:"invitationId"`
	FullName     string              `json:"fullName"`
	Email        string              `json:"email"`
	Role         string              `json:"role"`
	Class        *InvitationClassDTO `json:"class,omitempty"`
	Status       string              `json:"status"`
	ExpiresAt    string              `json:"expiresAt"`
	AcceptedAt   *string             `json:"acceptedAt,omitempty"`
	RevokedAt    *string             `json:"revokedAt,omitempty"`
	CreatedAt    string              `json:"createdAt"`
}

type CreateSchoolMemberInvitationResponseDTO struct {
	Message    string                    `json:"message"`
	Invitation SchoolMemberInvitationDTO `json:"invitation"`
	AcceptURL  string                    `json:"acceptUrl"`
	Token      string                    `json:"token"`
}

type SchoolMemberInvitationListResponseDTO struct {
	Data       []SchoolMemberInvitationDTO `json:"data"`
	TotalItems int64                       `json:"totalItems"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalPages int                         `json:"totalPages"`
}
