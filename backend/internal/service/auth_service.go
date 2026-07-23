package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AttemptLimiter interface {
	Allow(key string) bool
	Reset(key string)
}

var (
	ErrInvalidCurrentPassword  = errors.New("current password is incorrect")
	ErrTooManyPasswordAttempts = errors.New("too many failed attempts, try again later")

	ErrRefreshTokenInvalid     = errors.New("refresh token is invalid or expired")
	ErrRefreshTokenReused      = errors.New("refresh token reuse detected")
	ErrRefreshTokenRateLimited = errors.New("too many refresh attempts, try again later")

	ErrMFAPreAuthInvalid  = errors.New("mfa session is invalid or expired, please log in again")
	ErrMFACodeInvalid     = errors.New("invalid verification code")
	ErrMFATooManyAttempts = errors.New("too many failed attempts, try again later")
)

const changePasswordLockKeyPrefix = "change_password:"
const refreshRateLimitKeyPrefix = "refresh_token:"
const mfaVerifyLockKeyPrefix = "mfa_verify:"

// accessTokenTTL/refreshTokenTTL: 15-minute access token, 7-day refresh
// token — approved design for the Phase 11.3 refresh-token migration.
const accessTokenTTL = 15 * time.Minute
const refreshTokenTTL = 7 * 24 * time.Hour

// mfaPreAuthTokenTTL: long enough for a user to open their authenticator
// app and type a code, short enough not to leave a "password verified,
// waiting on MFA" window open for long.
const mfaPreAuthTokenTTL = 10 * time.Minute

// mfaGracePeriod: 7 days, confirmed design — counted from
// User.MFAGraceStartedAt, which is set on first login after this feature
// existed (see checkMFAGracePeriod), never from CreatedAt or a migration
// date.
const mfaGracePeriod = 7 * 24 * time.Hour

// RefreshTokenMetadata carries request-transport details (never gin.Context
// itself, to keep this package free of a web-framework dependency) that get
// persisted alongside a refresh token for the future Session Management UI.
type RefreshTokenMetadata struct {
	UserAgent string
	IPAddress string
}

// LoginResult is what Login/Register/VerifyMFA return instead of a bare
// (*dto.LoginResponseDTO, string, error) tuple, now that a login can stop
// short of issuing real tokens. Exactly one of Response/Challenge is
// non-nil.
type LoginResult struct {
	// Response is set once a login is fully complete — real tokens issued.
	Response *dto.LoginResponseDTO
	// RawRefreshToken is only meaningful when Response is non-nil — the
	// handler sets this as the refresh_token cookie.
	RawRefreshToken string
	// Challenge is set when a second step (MFA code, or forced enrollment)
	// still stands between here and a completed login.
	Challenge *dto.LoginChallengeDTO
}

type AuthService interface {
	Login(email string, password string, meta RefreshTokenMetadata) (*LoginResult, error)
	Register(fullName string, email string, password string, meta RefreshTokenMetadata) (*LoginResult, error)
	GetContext(userID string) (*dto.AuthContextResponseDTO, error)
	ChangePassword(userID string, currentPassword string, newPassword string) error
	Refresh(rawRefreshToken string, meta RefreshTokenMetadata) (string, string, error)
	Logout(rawRefreshToken string) error
	ListSessions(userID string) ([]*domain.RefreshToken, error)
	RevokeSession(userID string, sessionID string) error
	VerifyMFA(rawPreAuthToken string, code string, recoveryCode string, meta RefreshTokenMetadata) (*LoginResult, error)
}

type authService struct {
	userRepo              repository.UserRepository
	schoolUserRepo        repository.SchoolUserRepository
	emailVerificationSvc  EmailVerificationService
	logService            LogService
	refreshTokenRepo      repository.RefreshTokenRepository
	mfaService            MFAService
	passwordAttemptLimit  AttemptLimiter
	refreshAttemptLimit   AttemptLimiter
	mfaVerifyAttemptLimit AttemptLimiter
}

func NewAuthService(
	userRepo repository.UserRepository,
	schoolUserRepo repository.SchoolUserRepository,
	emailVerificationSvc EmailVerificationService,
	logService LogService,
	refreshTokenRepo repository.RefreshTokenRepository,
	mfaService MFAService,
	passwordAttemptLimit AttemptLimiter,
	refreshAttemptLimit AttemptLimiter,
	mfaVerifyAttemptLimit AttemptLimiter,
) AuthService {
	return &authService{
		userRepo:              userRepo,
		schoolUserRepo:        schoolUserRepo,
		emailVerificationSvc:  emailVerificationSvc,
		logService:            logService,
		refreshTokenRepo:      refreshTokenRepo,
		mfaService:            mfaService,
		passwordAttemptLimit:  passwordAttemptLimit,
		refreshAttemptLimit:   refreshAttemptLimit,
		mfaVerifyAttemptLimit: mfaVerifyAttemptLimit,
	}
}

func (s *authService) logLoginFailed(email string, reason string) {
	_ = s.logService.Log(domain.ActorContext{Scope: domain.LogScopePlatform}, "auth.login.failed", "user", nil, domain.LogSeverityMedium, map[string]any{
		"email":  email,
		"reason": reason,
	})
}

func (s *authService) Login(email string, password string, meta RefreshTokenMetadata) (*LoginResult, error) {
	userEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Return generic error to prevent user enumeration
		s.logLoginFailed(email, "user_not_found")
		return nil, errors.New("invalid email or password")
	}

	err = verifyPassword(userEmail.Password, password)
	if err != nil {
		// Return same generic error for password mismatch
		s.logLoginFailed(email, "invalid_password")
		return nil, errors.New("invalid email or password")
	}

	// --- MFA gate: sits between "password verified" and "tokens issued" ---
	mfaEnabled, err := s.mfaService.IsEnabled(userEmail.ID)
	if err != nil {
		return nil, err
	}

	if mfaEnabled {
		rawPreAuth, err := s.mfaService.IssuePreAuthToken(userEmail.ID, domain.MFAPreAuthPurposeVerify, mfaPreAuthTokenTTL)
		if err != nil {
			return nil, err
		}
		return &LoginResult{Challenge: &dto.LoginChallengeDTO{
			MFARequired:  true,
			PreAuthToken: rawPreAuth,
		}}, nil
	}

	graceDaysRemaining, graceExpired, err := s.checkMFAGracePeriod(userEmail)
	if err != nil {
		return nil, err
	}

	if graceExpired {
		rawPreAuth, err := s.mfaService.IssuePreAuthToken(userEmail.ID, domain.MFAPreAuthPurposeEnrollRequired, mfaPreAuthTokenTTL)
		if err != nil {
			return nil, err
		}
		return &LoginResult{Challenge: &dto.LoginChallengeDTO{
			MFASetupRequired: true,
			PreAuthToken:     rawPreAuth,
		}}, nil
	}

	return s.completeLogin(userEmail, meta, graceDaysRemaining)
}

func (s *authService) checkMFAGracePeriod(user *domain.User) (daysRemaining *int, expired bool, err error) {
	now := time.Now()
	if user.MFAGraceStartedAt == nil {
		user.MFAGraceStartedAt = &now
		if err := s.userRepo.Update(user); err != nil {
			return nil, false, err
		}
		remaining := int(mfaGracePeriod.Hours() / 24)
		return &remaining, false, nil
	}

	deadline := user.MFAGraceStartedAt.Add(mfaGracePeriod)
	if now.After(deadline) {
		return nil, true, nil
	}
	remaining := int(deadline.Sub(now).Hours()/24) + 1
	return &remaining, false, nil
}

func (s *authService) Register(fullName string, email string, password string, meta RefreshTokenMetadata) (*LoginResult, error) {
	isEmailExists, err := s.userRepo.CheckEmailExists(email, "")
	if err != nil {
		return nil, err
	}
	if isEmailExists {
		return nil, errors.New("Email already registered")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName: fullName,
		Email:    email,
		Password: hashedPassword,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.registered", "user", strPtr(user.ID), domain.LogSeverityMedium, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})

	if s.emailVerificationSvc != nil {
		if err := s.emailVerificationSvc.IssueForNewUser(user); err != nil {
			fmt.Printf("[Email Verification Warning] failed to issue token for user_id=%s error=%s\n", user.ID, err.Error())
		}
	}

	// Auto-login after registration. A brand-new user has no UserMFA row
	// (mfaEnabled is always false) and MFAGraceStartedAt is always NULL at
	// this point (just created) — checkMFAGracePeriod always treats a NULL
	// start as "just started, not expired," so this can never land in the
	// mfaSetupRequired branch. See auth_service_test.go for a test that
	// exercises this directly rather than relying on the reasoning alone.
	return s.Login(email, password, meta)
}

func (s *authService) GetContext(userID string) (*dto.AuthContextResponseDTO, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return s.buildAuthContext(user)
}

// ChangePassword is the self-service counterpart to UserService.ChangePassword
// (which is a super-admin-on-behalf-of-another-user reset). userID always
// comes from the caller's own JWT claims, never a path/body-supplied ID.
func (s *authService) ChangePassword(userID string, currentPassword string, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if err := verifyPassword(user.Password, currentPassword); err != nil {
		lockKey := changePasswordLockKeyPrefix + userID
		reason := "invalid_current_password"
		failErr := ErrInvalidCurrentPassword
		if s.passwordAttemptLimit != nil && !s.passwordAttemptLimit.Allow(lockKey) {
			reason = "rate_limited"
			failErr = ErrTooManyPasswordAttempts
		}
		s.logChangePasswordFailed(userID, reason)
		return failErr
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if s.passwordAttemptLimit != nil {
		s.passwordAttemptLimit.Reset(changePasswordLockKeyPrefix + userID)
	}

	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.password.changed", "user", strPtr(userID), domain.LogSeverityHigh, map[string]any{
		"user_id": userID,
		"method":  "self_service",
	})
	return nil
}

func (s *authService) logChangePasswordFailed(userID string, reason string) {
	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.password.change.failed", "user", strPtr(userID), domain.LogSeverityMedium, map[string]any{
		"user_id": userID,
		"reason":  reason,
	})
}

// Refresh validates and rotates a refresh token, returning a fresh access
// token + fresh raw refresh token on success.
func (s *authService) Refresh(rawRefreshToken string, meta RefreshTokenMetadata) (string, string, error) {
	tokenHash, err := hashRefreshToken(rawRefreshToken)
	if err != nil {
		return "", "", ErrRefreshTokenInvalid
	}

	now := time.Now()
	familyIDForRateLimit := ""
	if validToken, validErr := s.refreshTokenRepo.FindValidByTokenHash(tokenHash, now); validErr == nil {
		familyIDForRateLimit = validToken.FamilyID
	} else if existing, findErr := s.refreshTokenRepo.FindByTokenHash(tokenHash); findErr == nil {
		familyIDForRateLimit = existing.FamilyID
	}
	if familyIDForRateLimit != "" && s.refreshAttemptLimit != nil && !s.refreshAttemptLimit.Allow(refreshRateLimitKeyPrefix+familyIDForRateLimit) {
		return "", "", ErrRefreshTokenRateLimited
	}

	newRawToken, newTokenHash, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	newRecord := &domain.RefreshToken{
		TokenHash: newTokenHash,
		ExpiresAt: now.Add(refreshTokenTTL),
	}
	if meta.UserAgent != "" {
		userAgent := meta.UserAgent
		newRecord.UserAgent = &userAgent
	}
	if meta.IPAddress != "" {
		ipAddress := meta.IPAddress
		newRecord.IPAddress = &ipAddress
	}

	rotated, err := s.refreshTokenRepo.Rotate(tokenHash, newRecord)
	if err != nil {
		var reused *repository.ReusedRefreshTokenError
		if errors.As(err, &reused) {
			_ = s.refreshTokenRepo.RevokeFamily(reused.FamilyID)
			_ = s.logService.Log(domain.ActorContext{UserID: reused.UserID, Scope: domain.LogScopePlatform}, "auth.token.reuse_detected", "user", strPtr(reused.UserID), domain.LogSeverityHigh, map[string]any{
				"user_id":   reused.UserID,
				"family_id": reused.FamilyID,
			})
			return "", "", ErrRefreshTokenReused
		}
		return "", "", ErrRefreshTokenInvalid
	}

	user, err := s.userRepo.GetByID(rotated.UserID)
	if err != nil {
		return "", "", err
	}

	accessToken, err := s.issueAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.token.refreshed", "user", strPtr(user.ID), domain.LogSeverityLow, map[string]any{
		"user_id": user.ID,
	})

	return accessToken, newRawToken, nil
}

// Logout always returns nil to the caller — an already-invalid or garbage
// token is not an error, it just means there's nothing left to revoke.
func (s *authService) Logout(rawRefreshToken string) error {
	tokenHash, err := hashRefreshToken(rawRefreshToken)
	if err != nil {
		return nil
	}

	record, _ := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	_ = s.refreshTokenRepo.RevokeByTokenHash(tokenHash)

	actor := domain.ActorContext{Scope: domain.LogScopePlatform}
	metadata := map[string]any{}
	var entityID *string
	if record != nil {
		actor.UserID = record.UserID
		entityID = strPtr(record.UserID)
		metadata["user_id"] = record.UserID
	}
	_ = s.logService.Log(actor, "auth.logout", "user", entityID, domain.LogSeverityLow, metadata)
	return nil
}

func (s *authService) ListSessions(userID string) ([]*domain.RefreshToken, error) {
	return s.refreshTokenRepo.FindActiveByUserID(userID, time.Now())
}

// RevokeSession fetches the row first and explicitly checks ownership —
// same defense-in-depth pattern as LogHandler.GetByIDInSchool — rather than
// relying solely on a WHERE clause to enforce it.
func (s *authService) RevokeSession(userID string, sessionID string) error {
	token, err := s.refreshTokenRepo.FindByID(sessionID)
	if err != nil {
		return ErrRefreshTokenInvalid
	}
	if token.UserID != userID || token.RevokedAt != nil {
		return ErrRefreshTokenInvalid
	}

	if err := s.refreshTokenRepo.RevokeByID(sessionID); err != nil {
		return err
	}

	userAgent := ""
	if token.UserAgent != nil {
		userAgent = *token.UserAgent
	}
	ipAddress := ""
	if token.IPAddress != nil {
		ipAddress = *token.IPAddress
	}

	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.session.revoked", "session", strPtr(sessionID), domain.LogSeverityMedium, map[string]any{
		"user_id":    userID,
		"session_id": sessionID,
		"user_agent": userAgent,
		"ip_address": ipAddress,
	})
	return nil
}

// VerifyMFA completes a login that was paused for MFA (a Challenge with
// MFARequired=true) — resolves the pre-auth token, validates the submitted
// code (TOTP or recovery code — exactly one should be given; if both are,
// Code takes precedence), and only then issues real tokens via
// completeLogin, exactly like a normal password-only Login.
func (s *authService) VerifyMFA(rawPreAuthToken string, code string, recoveryCode string, meta RefreshTokenMetadata) (*LoginResult, error) {
	token, err := s.mfaService.ResolvePreAuthToken(rawPreAuthToken)
	if err != nil {
		return nil, ErrMFAPreAuthInvalid
	}
	if token.Purpose != domain.MFAPreAuthPurposeVerify {
		return nil, ErrMFAPreAuthInvalid
	}

	user, err := s.userRepo.GetByID(token.UserID)
	if err != nil {
		return nil, err
	}

	usedRecovery := code == ""
	var verifyErr error
	if !usedRecovery {
		verifyErr = s.mfaService.VerifyCode(user.ID, code)
	} else {
		verifyErr = s.mfaService.VerifyRecoveryCode(user.ID, recoveryCode)
	}

	if verifyErr != nil {
		lockKey := mfaVerifyLockKeyPrefix + user.ID
		failErr := ErrMFACodeInvalid
		if s.mfaVerifyAttemptLimit != nil && !s.mfaVerifyAttemptLimit.Allow(lockKey) {
			failErr = ErrMFATooManyAttempts
		}
		_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.mfa.verify.failed", "user", strPtr(user.ID), domain.LogSeverityMedium, map[string]any{
			"user_id":       user.ID,
			"used_recovery": usedRecovery,
		})
		return nil, failErr
	}

	if s.mfaVerifyAttemptLimit != nil {
		s.mfaVerifyAttemptLimit.Reset(mfaVerifyLockKeyPrefix + user.ID)
	}

	if err := s.mfaService.ConsumePreAuthToken(token.ID); err != nil {
		return nil, err
	}

	if usedRecovery {
		_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.mfa.recovery_code.used", "user", strPtr(user.ID), domain.LogSeverityHigh, map[string]any{
			"user_id": user.ID,
		})
	} else {
		_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.mfa.verified", "user", strPtr(user.ID), domain.LogSeverityLow, map[string]any{
			"user_id": user.ID,
		})
	}

	// MFA is already enabled for this user — no grace-period reminder needed.
	return s.completeLogin(user, meta, nil)
}

// completeLogin issues real tokens and builds the success response —
// shared by the direct password-only Login path and VerifyMFA's completion
// path, so auth.login.success/member.login fire identically regardless of
// which path established the session.
func (s *authService) completeLogin(user *domain.User, meta RefreshTokenMetadata, graceDaysRemaining *int) (*LoginResult, error) {
	accessToken, err := s.issueAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// A fresh login always starts a brand-new session family — it has no
	// prior token to rotate from.
	rawRefreshToken, err := s.issueRefreshToken(user.ID, uuid.NewString(), meta)
	if err != nil {
		return nil, err
	}

	response, err := s.buildLoginResponse(accessToken, user)
	if err != nil {
		return nil, err
	}
	response.MFAGraceDaysRemaining = graceDaysRemaining

	_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.login.success", "user", strPtr(user.ID), domain.LogSeverityLow, map[string]any{
		"user_id":      user.ID,
		"login_method": "password",
	})

	if response.DefaultContext != nil {
		schoolID := response.DefaultContext.SchoolID
		schoolUserID := response.DefaultContext.SchoolUserID
		_ = s.logService.Log(domain.ActorContext{
			UserID:       user.ID,
			SchoolID:     &schoolID,
			SchoolUserID: &schoolUserID,
			Scope:        domain.LogScopeSchool,
		}, "member.login", "school_user", strPtr(schoolUserID), domain.LogSeverityLow, map[string]any{
			"login_method": "password",
			"user_id":      user.ID,
			"school_id":    schoolID,
		})
	}

	return &LoginResult{Response: response, RawRefreshToken: rawRefreshToken}, nil
}

// issueAccessToken mints a short-lived (accessTokenTTL) JWT — same claims
// shape used since before this migration (user_id, sub, email, exp).
func (s *authService) issueAccessToken(userID string, email string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("server configuration error")
	}

	payload := jwt.MapClaims{
		"user_id": userID,
		"sub":     userID,
		"email":   email,
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(secretKey))
}

func (s *authService) issueRefreshToken(userID string, familyID string, meta RefreshTokenMetadata) (string, error) {
	rawToken, tokenHash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	record := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}
	if meta.UserAgent != "" {
		userAgent := meta.UserAgent
		record.UserAgent = &userAgent
	}
	if meta.IPAddress != "" {
		ipAddress := meta.IPAddress
		record.IPAddress = &ipAddress
	}

	if err := s.refreshTokenRepo.Create(record); err != nil {
		return "", err
	}
	return rawToken, nil
}

func generateRefreshToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func hashRefreshToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", repository.ErrRefreshTokenInvalid
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func (s *authService) buildLoginResponse(token string, user *domain.User) (*dto.LoginResponseDTO, error) {
	response := &dto.LoginResponseDTO{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
		},
	}

	context, err := s.buildAuthContext(user)
	if err != nil {
		return nil, err
	}
	response.Memberships = context.Memberships
	response.GlobalRoles = context.GlobalRoles
	response.DefaultContext = context.DefaultContext

	return response, nil
}

func (s *authService) buildAuthContext(user *domain.User) (*dto.AuthContextResponseDTO, error) {
	response := &dto.AuthContextResponseDTO{
		Memberships:     []dto.MembershipInfo{},
		GlobalRoles:     []string{},
		EmailVerified:   user.EmailVerifiedAt != nil,
		EmailVerifiedAt: formatAPITimePtr(user.EmailVerifiedAt),
	}

	if s.schoolUserRepo == nil {
		return response, nil
	}

	schoolUsers, err := s.schoolUserRepo.GetByUser(user.ID)
	if err != nil {
		return nil, err
	}

	globalRoleSet := map[string]bool{}
	activeMembershipIndex := 0
	for _, schoolUser := range schoolUsers {
		if schoolUser.DeletedAt.Valid {
			continue
		}

		roles := make([]string, 0, len(schoolUser.Roles))
		for _, userRole := range schoolUser.Roles {
			if userRole.Role.Name == "" {
				continue
			}
			roles = append(roles, userRole.Role.Name)
			if userRole.Role.Name == "super_admin" && !globalRoleSet[userRole.Role.Name] {
				response.GlobalRoles = append(response.GlobalRoles, userRole.Role.Name)
				globalRoleSet[userRole.Role.Name] = true
			}
		}

		membership := dto.MembershipInfo{
			SchoolUserID: schoolUser.ID,
			School: dto.SchoolInfo{
				ID:   schoolUser.School.ID,
				Code: schoolUser.School.Code,
				Name: schoolUser.School.Name,
			},
			Roles:     roles,
			IsDefault: activeMembershipIndex == 0,
		}
		response.Memberships = append(response.Memberships, membership)
		activeMembershipIndex++

		if response.DefaultContext == nil {
			response.DefaultContext = &dto.DefaultContext{
				SchoolID:     schoolUser.SchoolID,
				SchoolUserID: schoolUser.ID,
				Roles:        roles,
			}
		}
	}

	return response, nil
}
