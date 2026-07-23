package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, err := h.authService.Login(input.Email, input.Password)
	if err != nil {
		// Always return 401 Unauthorized with generic message
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, err := h.authService.Register(input.FullName, input.Email, input.Password)
	if err != nil {
		if err.Error() == "Email already registered" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			HandleError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.ChangeOwnPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := dto.ValidatePasswordComplexity(input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ChangePassword(userID, input.CurrentPassword, input.NewPassword)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	case errors.Is(err, service.ErrTooManyPasswordAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCurrentPassword):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		HandleError(c, err)
	}
}

func (h *AuthHandler) GetContext(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.authService.GetContext(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
