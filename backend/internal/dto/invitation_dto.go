package dto

type InvitationSchoolDTO struct {
	SchoolID   string `json:"schoolId"`
	SchoolCode string `json:"schoolCode"`
	SchoolName string `json:"schoolName"`
}

type InvitationMetadataDTO struct {
	InvitationID string              `json:"invitationId"`
	Email        string              `json:"email"`
	Role         string              `json:"role"`
	School       InvitationSchoolDTO `json:"school"`
	ExpiresAt    string              `json:"expiresAt"`
	Status       string              `json:"status"`
	ExistingUser bool                `json:"existingUser"`
}

type AcceptInvitationDTO struct {
	Name            string `json:"name" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type InvitationAcceptedUserDTO struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type AcceptInvitationResponseDTO struct {
	Message string                    `json:"message"`
	User    InvitationAcceptedUserDTO `json:"user"`
	School  InvitationSchoolDTO       `json:"school"`
	Role    string                    `json:"role"`
}
