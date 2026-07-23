package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// passwordResetTokenTTL is deliberately short — a reset link is meant to be
// used within minutes of requesting it (right after opening the email), not
// treated as a standing credential. 20 minutes, chosen from the confirmed
// 15-30 minute range: long enough that a slow mail provider or a user who
// steps away briefly doesn't get an unnecessarily expired link, short enough
// to keep the exposure window for a leaked/intercepted email tight.
const passwordResetTokenTTL = 20 * time.Minute

type PasswordResetService interface {
	Request(email string) error
	GetMetadata(token string) (*dto.PasswordResetMetadataDTO, error)
	Reset(token string, newPassword string) error
}

type passwordResetService struct {
	repo         repository.PasswordResetRepository
	userRepo     repository.UserRepository
	emailService EmailService
	logService   LogService
}

func NewPasswordResetService(repo repository.PasswordResetRepository, userRepo repository.UserRepository, emailService EmailService, logService LogService) PasswordResetService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &passwordResetService{repo: repo, userRepo: userRepo, emailService: emailService, logService: logService}
}

// Request always returns nil for "email not registered" (deliberately
// indistinguishable from "email registered, link sent" to the caller — see
// PasswordResetHandler.Request) and only returns a real error for genuine
// system faults (DB write failure, token generation failure).
func (s *passwordResetService) Request(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// auth.password.reset.requested never carries UserID, by design, even
	// once the user is resolved below — it records that a request for this
	// email happened, not who (if anyone) actually received it.
	logRequested := func() {
		_ = s.logService.Log(domain.ActorContext{Scope: domain.LogScopePlatform}, "auth.password.reset.requested", "user", nil, domain.LogSeverityLow, map[string]any{
			"email": email,
		})
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		logRequested()
		return nil
	}

	rawToken, tokenHash, err := generatePasswordResetToken()
	if err != nil {
		return err
	}

	if err := s.repo.InvalidateAllForUser(user.ID); err != nil {
		return err
	}

	reset := &domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(passwordResetTokenTTL),
	}
	if err := s.repo.Create(reset); err != nil {
		return err
	}

	resetURL := buildPasswordResetURL(rawToken)
	if err := s.emailService.SendPasswordReset(user.Email, user.FullName, resetURL); err != nil {
		fmt.Printf("[Password Reset Warning] failed to send reset email user_id=%s error=%s\n", user.ID, err.Error())
	}

	logRequested()
	return nil
}

func (s *passwordResetService) GetMetadata(token string) (*dto.PasswordResetMetadataDTO, error) {
	tokenHash, err := hashPasswordResetToken(token)
	if err != nil {
		return nil, err
	}

	reset, err := s.repo.GetValidByTokenHash(tokenHash, time.Now())
	if err != nil {
		return nil, err
	}

	return &dto.PasswordResetMetadataDTO{
		Status:    "valid",
		ExpiresAt: formatAPITime(reset.ExpiresAt),
	}, nil
}

func (s *passwordResetService) Reset(token string, newPassword string) error {
	tokenHash, err := hashPasswordResetToken(token)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	reset, err := s.repo.ConsumeAndSetPassword(tokenHash, hashedPassword, time.Now())
	if err != nil {
		return err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: reset.UserID, Scope: domain.LogScopePlatform}, "auth.password.reset.completed", "user", strPtr(reset.UserID), domain.LogSeverityHigh, map[string]any{
		"user_id": reset.UserID,
	})
	return nil
}

func generatePasswordResetToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func hashPasswordResetToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", repository.ErrPasswordResetInvalid
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func buildPasswordResetURL(rawToken string) string {
	path := "/reset-password/" + rawToken
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if publicURL == "" {
		return path
	}
	return publicURL + path
}
