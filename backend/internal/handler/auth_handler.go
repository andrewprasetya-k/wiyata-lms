package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService     service.AuthService
	wsTicketService service.WSTicketService
}

func NewAuthHandler(authService service.AuthService, wsTicketService service.WSTicketService) *AuthHandler {
	return &AuthHandler{authService: authService, wsTicketService: wsTicketService}
}

const refreshTokenCookieName = "refresh_token"
const refreshTokenCookiePath = "/api"
const refreshTokenCookieMaxAgeSeconds = 7 * 24 * 60 * 60 // 7 days, matches refreshTokenTTL

func setRefreshTokenCookie(c *gin.Context, rawToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshTokenCookieName, rawToken, refreshTokenCookieMaxAgeSeconds, refreshTokenCookiePath, "", true, true)
}

func clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshTokenCookieName, "", -1, refreshTokenCookiePath, "", true, true)
}

func refreshTokenMetadataFromRequest(c *gin.Context) service.RefreshTokenMetadata {
	return service.RefreshTokenMetadata{
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, rawRefreshToken, err := h.authService.Login(input.Email, input.Password, refreshTokenMetadataFromRequest(c))
	if err != nil {
		// Always return 401 Unauthorized with generic message
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	setRefreshTokenCookie(c, rawRefreshToken)
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, rawRefreshToken, err := h.authService.Register(input.FullName, input.Email, input.Password, refreshTokenMetadataFromRequest(c))
	if err != nil {
		if err.Error() == "Email already registered" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			HandleError(c, err)
		}
		return
	}

	setRefreshTokenCookie(c, rawRefreshToken)
	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil || rawToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accessToken, newRawToken, err := h.authService.Refresh(rawToken, refreshTokenMetadataFromRequest(c))
	if err != nil {
		clearRefreshTokenCookie(c)
		if errors.Is(err, service.ErrRefreshTokenRateLimited) {
			fmt.Printf("[Refresh Token] 429 rate limited, ip=%s\n", c.ClientIP())
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("[Refresh Token] 401 invalid/expired/reused, ip=%s, reason=%v\n", c.ClientIP(), err)
	
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please log in again"})
		return
	}

	setRefreshTokenCookie(c, newRawToken)
	c.JSON(http.StatusOK, dto.RefreshTokenResponseDTO{AccessToken: accessToken})
}

// Logout is POST /logout — public
func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken, _ := c.Cookie(refreshTokenCookieName)
	_ = h.authService.Logout(rawToken)
	clearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, dto.LogoutResponseDTO{Message: "Logged out"})
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

// IssueWSTicket is GET /me/ws-ticket — AuthRequired
func (h *AuthHandler) IssueWSTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ticket, err := h.wsTicketService.Issue(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.WSTicketResponseDTO{Ticket: ticket})
}
