package dto

type VerifyEmailDTO struct {
	Token string `json:"token" binding:"required"`
}

type VerifyEmailResponseDTO struct {
	Message         string `json:"message"`
	EmailVerifiedAt string `json:"emailVerifiedAt"`
}

type ResendVerificationResponseDTO struct {
	Message string `json:"message"`
}
