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
	mfaService      service.MFAService
}

func NewAuthHandler(authService service.AuthService, wsTicketService service.WSTicketService, mfaService service.MFAService) *AuthHandler {
	return &AuthHandler{authService: authService, wsTicketService: wsTicketService, mfaService: mfaService}
}

// respondLoginResult writes either the challenge JSON (no cookie — no real
// session started yet) or the full success response with the refresh
// cookie set, depending on which branch of LoginResult is populated.
func respondLoginResult(c *gin.Context, statusCode int, result *service.LoginResult) {
	if result.Challenge != nil {
		c.JSON(http.StatusOK, result.Challenge)
		return
	}
	setRefreshTokenCookie(c, result.RawRefreshToken)
	c.JSON(statusCode, result.Response)
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

	result, err := h.authService.Login(input.Email, input.Password, refreshTokenMetadataFromRequest(c))
	if err != nil {
		// Always return 401 Unauthorized with generic message
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	respondLoginResult(c, http.StatusOK, result)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	result, err := h.authService.Register(input.FullName, input.Email, input.Password, refreshTokenMetadataFromRequest(c))
	if err != nil {
		if err.Error() == "Email already registered" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			HandleError(c, err)
		}
		return
	}

	respondLoginResult(c, http.StatusCreated, result)
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

// ListSessions is GET /me/sessions — AuthRequired.
func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessions, err := h.authService.ListSessions(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := make([]dto.SessionDTO, 0, len(sessions))
	for _, session := range sessions {
		userAgent := ""
		if session.UserAgent != nil {
			userAgent = *session.UserAgent
		}
		ipAddress := ""
		if session.IPAddress != nil {
			ipAddress = *session.IPAddress
		}
		response = append(response, dto.SessionDTO{
			ID:         session.ID,
			LoggedInAt: formatAPITime(session.CreatedAt),
			ExpiresAt:  formatAPITime(session.ExpiresAt),
			UserAgent:  userAgent,
			IPAddress:  ipAddress,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionID := c.Param("id")
	if err := h.authService.RevokeSession(userID, sessionID); err != nil {
		if errors.Is(err, service.ErrRefreshTokenInvalid) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked"})
}

// EnrollMFA is POST /me/mfa/enroll — AuthRequired. Starts (or restarts, if
// abandoned before confirmation) TOTP enrollment for the caller.
func (h *AuthHandler) EnrollMFA(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	secret, otpauthURL, err := h.mfaService.Enroll(userID)
	if err != nil {
		if errors.Is(err, service.ErrMFAAlreadyEnabled) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MFAEnrollResponseDTO{Secret: secret, OTPAuthURL: otpauthURL})
}

// ConfirmMFA is POST /me/mfa/confirm — AuthRequired. Validates the first
// code from the authenticator app configured via EnrollMFA, enables MFA,
// and returns recovery codes exactly once.
func (h *AuthHandler) ConfirmMFA(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.MFAConfirmDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	recoveryCodes, err := h.mfaService.ConfirmEnrollment(userID, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			HandleError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, dto.MFAConfirmResponseDTO{RecoveryCodes: recoveryCodes})
}

// VerifyMFALogin is POST /login/mfa-verify — public. Completes a login that
// was paused for MFA (a Login/Register response with mfaRequired=true).
func (h *AuthHandler) VerifyMFALogin(c *gin.Context) {
	var input dto.MFAVerifyLoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	result, err := h.authService.VerifyMFA(input.PreAuthToken, input.Code, input.RecoveryCode, refreshTokenMetadataFromRequest(c))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAPreAuthInvalid):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrMFATooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrMFACodeInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			HandleError(c, err)
		}
		return
	}

	respondLoginResult(c, http.StatusOK, result)
}
