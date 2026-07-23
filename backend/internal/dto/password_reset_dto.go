package dto

// ForgotPasswordDTO is the POST /forgot-password request body.
type ForgotPasswordDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordResponseDTO struct {
	Message string `json:"message"`
}

// PasswordResetMetadataDTO is the GET /reset-password/:token response — a
// non-consuming validity check so the frontend can show a valid form vs.
// an error state before the user submits anything.
type PasswordResetMetadataDTO struct {
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt"`
}

// ResetPasswordDTO is the POST /reset-password/:token request body — the
// token itself comes from the URL path, not the body. Complexity beyond
// min=8 is checked via ValidatePasswordComplexity (password_policy.go),
// reused as-is rather than duplicated here.
type ResetPasswordDTO struct {
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type ResetPasswordResponseDTO struct {
	Message string `json:"message"`
}
