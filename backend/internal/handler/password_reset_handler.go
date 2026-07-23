package handler

import (
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	service service.PasswordResetService
}

func NewPasswordResetHandler(service service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: service}
}

// Request is POST /forgot-password — always returns the same generic
// success shape regardless of whether the email is registered, to avoid
// leaking which emails exist in the system. A real service-layer error
// (DB fault, token generation failure) still surfaces as an error response,
// since that doesn't distinguish email existence either way.
func (h *PasswordResetHandler) Request(c *gin.Context) {
	var input dto.ForgotPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := h.service.Request(input.Email); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ForgotPasswordResponseDTO{
		Message: "Jika email tersebut terdaftar, tautan reset password telah dikirim.",
	})
}

// GetMetadata is GET /reset-password/:token — validates without consuming,
// mirroring GET /invitations/:token's shape.
func (h *PasswordResetHandler) GetMetadata(c *gin.Context) {
	response, err := h.service.GetMetadata(c.Param("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link reset password tidak valid atau sudah kedaluwarsa."})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Reset is POST /reset-password/:token.
func (h *PasswordResetHandler) Reset(c *gin.Context) {
	var input dto.ResetPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := dto.ValidatePasswordComplexity(input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Reset(c.Param("token"), input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link reset password tidak valid atau sudah kedaluwarsa."})
		return
	}

	c.JSON(http.StatusOK, dto.ResetPasswordResponseDTO{
		Message: "Password berhasil direset. Silakan masuk dengan password baru.",
	})
}
