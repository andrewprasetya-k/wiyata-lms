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
)

var ErrEmailAlreadyVerified = errors.New("email is already verified")

type EmailVerificationService interface {
	IssueForNewUser(user *domain.User) error
	Verify(token string) (*dto.VerifyEmailResponseDTO, error)
	Resend(userID string) (*dto.ResendVerificationResponseDTO, error)
}

type emailVerificationService struct {
	repo         repository.EmailVerificationRepository
	userRepo     repository.UserRepository
	emailService EmailService
	logService   LogService
}

func NewEmailVerificationService(repo repository.EmailVerificationRepository, userRepo repository.UserRepository, emailService EmailService, logService LogService) EmailVerificationService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &emailVerificationService{repo: repo, userRepo: userRepo, emailService: emailService, logService: logService}
}

func (s *emailVerificationService) IssueForNewUser(user *domain.User) error {
	if user == nil || user.ID == "" {
		return fmt.Errorf("email verification requires a persisted user")
	}
	return s.issueAndSend(user)
}

func (s *emailVerificationService) Resend(userID string) (*dto.ResendVerificationResponseDTO, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.EmailVerifiedAt != nil {
		return nil, ErrEmailAlreadyVerified
	}

	if err := s.repo.InvalidateAllForUser(user.ID); err != nil {
		return nil, err
	}
	if err := s.issueAndSend(user); err != nil {
		return nil, err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.verification.resent", "user", strPtr(user.ID), domain.LogSeverityLow, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})

	return &dto.ResendVerificationResponseDTO{Message: "Verification email sent"}, nil
}

func (s *emailVerificationService) Verify(token string) (*dto.VerifyEmailResponseDTO, error) {
	tokenHash, err := hashVerificationToken(token)
	if err != nil {
		return nil, err
	}

	verification, err := s.repo.ConsumeByTokenHash(tokenHash, time.Now())
	if err != nil {
		return nil, err
	}

	if user, err := s.userRepo.GetByID(verification.UserID); err == nil {
		_ = s.logService.Log(domain.ActorContext{UserID: verification.UserID, Scope: domain.LogScopePlatform}, "auth.email.verified", "user", strPtr(verification.UserID), domain.LogSeverityLow, map[string]any{
			"user_id": verification.UserID,
			"email":   user.Email,
		})
	}

	return &dto.VerifyEmailResponseDTO{
		Message:         "Email verified",
		EmailVerifiedAt: formatAPITime(*verification.ConsumedAt),
	}, nil
}

func (s *emailVerificationService) issueAndSend(user *domain.User) error {
	rawToken, tokenHash, err := generateEmailVerificationToken()
	if err != nil {
		return err
	}

	verification := &domain.EmailVerification{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.repo.Create(verification); err != nil {
		return err
	}

	verifyURL := buildEmailVerificationURL(rawToken)
	if err := s.emailService.SendEmailVerification(user.Email, user.FullName, verifyURL); err != nil {
		fmt.Printf("[Email Warning] failed to send email verification user_id=%s error=%s\n", user.ID, err.Error())
	}
	return nil
}

func generateEmailVerificationToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func hashVerificationToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", repository.ErrEmailVerificationInvalid
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func buildEmailVerificationURL(rawToken string) string {
	path := "/verify-email?token=" + rawToken
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if publicURL == "" {
		return path
	}
	return publicURL + path
}
