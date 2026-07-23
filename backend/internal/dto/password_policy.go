package dto

import (
	"errors"
	"unicode"
)

// ChangeOwnPasswordDTO is the self-service "change my own password" request
// body — distinct from ChangePasswordDTO (user_dto.go), which is the
// super-admin-on-behalf-of-another-user reset and is untouched by this.
type ChangeOwnPasswordDTO struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

// Character-class requirements min=8 (the DTO tag above) can't express —
// Gin's binding tags have no built-in "must contain an uppercase letter"
// rule, so this is checked explicitly, once, right after binding succeeds.
var (
	ErrPasswordNeedsUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNeedsLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNeedsNumber    = errors.New("password must contain at least one number")
)

// ValidatePasswordComplexity checks the character-class rules the min=8
// binding tag can't express. Length is intentionally not re-checked here —
// that's already enforced by the DTO's own binding tag before this runs.
func ValidatePasswordComplexity(password string) error {
	var hasUpper, hasLower, hasNumber bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		}
	}
	if !hasUpper {
		return ErrPasswordNeedsUppercase
	}
	if !hasLower {
		return ErrPasswordNeedsLowercase
	}
	if !hasNumber {
		return ErrPasswordNeedsNumber
	}
	return nil
}
