package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmailVerificationHandler struct {
	service service.EmailVerificationService
}

func NewEmailVerificationHandler(service service.EmailVerificationService) *EmailVerificationHandler {
	return &EmailVerificationHandler{service: service}
}

func (h *EmailVerificationHandler) Verify(c *gin.Context) {
	var input dto.VerifyEmailDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification link is invalid or expired"})
		return
	}

	response, err := h.service.Verify(input.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification link is invalid or expired"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *EmailVerificationHandler) Resend(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.service.Resend(userID)
	if err != nil {
		if err == service.ErrEmailAlreadyVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already verified"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
