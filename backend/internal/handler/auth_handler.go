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

// Refresh is POST /refresh-token — public (no AuthRequired: a caller with an
// expired or missing access token still needs to be able to call this).
// Reads the refresh token from the httpOnly cookie only, never the body.
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
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		// Deliberately the same generic message for both "invalid/expired"
		// and "reuse detected" — never tell whoever holds this token which
		// case occurred, same principle as login's generic failure message.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please log in again"})
		return
	}

	setRefreshTokenCookie(c, newRawToken)
	c.JSON(http.StatusOK, dto.RefreshTokenResponseDTO{AccessToken: accessToken})
}

// Logout is POST /logout — public (same reasoning as Refresh: a caller with
// no valid access token must still be able to end their session). Always
// succeeds from the caller's point of view.
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
