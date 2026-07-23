package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

var (
	ErrMFAInvalidCode    = errors.New("invalid verification code")
	ErrMFAAlreadyEnabled = errors.New("mfa is already enabled for this user")
)

const (
	mfaIssuer            = "Wiyata"
	mfaRecoveryCodeCount  = 10
	recoveryCodeAlphabet  = "ABCDEFGHJKMNPQRSTUVWXYZ23456789" // excludes O/0, I/1, L — reduces transcription errors
)

type MFAService interface {
	// Enroll starts (or restarts, if the previous attempt was abandoned
	// before confirmation) enrollment for userID: generates a fresh TOTP
	// secret, persists it encrypted with enabled_at still NULL, and returns
	// both the raw secret (for manual entry) and the full otpauth:// URI
	// (for the frontend to render as a QR code — this package doesn't
	// generate images).
	Enroll(userID string) (secret string, otpauthURL string, err error)
	// ConfirmEnrollment validates the first code against the pending
	// secret; on success it sets enabled_at and returns a freshly generated
	// set of recovery codes in the clear — the only time they're ever
	// returned, only their hashes are persisted.
	ConfirmEnrollment(userID string, code string) (recoveryCodes []string, err error)
	IsEnabled(userID string) (bool, error)
	// VerifyCode validates a TOTP code for a user who already has MFA
	// enabled — the login-time check.
	VerifyCode(userID string, code string) error
	// VerifyRecoveryCode validates AND consumes (single-use) one recovery
	// code.
	VerifyRecoveryCode(userID string, code string) error

	// IssuePreAuthToken/ResolvePreAuthToken/ConsumePreAuthToken back the
	// login-flow "pre-auth token" — see domain.MFAPreAuthToken.
	IssuePreAuthToken(userID string, purpose string, ttl time.Duration) (rawToken string, err error)
	ResolvePreAuthToken(rawToken string) (*domain.MFAPreAuthToken, error)
	ConsumePreAuthToken(id string) error
}

type mfaService struct {
	repo       repository.MFARepository
	userRepo   repository.UserRepository
	logService LogService
}

func NewMFAService(repo repository.MFARepository, userRepo repository.UserRepository, logService LogService) MFAService {
	return &mfaService{repo: repo, userRepo: userRepo, logService: logService}
}

func (s *mfaService) Enroll(userID string) (string, string, error) {
	existing, err := s.repo.GetByUserID(userID)
	if err != nil && !errors.Is(err, repository.ErrMFANotEnrolled) {
		return "", "", err
	}
	if existing != nil && existing.EnabledAt != nil {
		return "", "", ErrMFAAlreadyEnabled
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      mfaIssuer,
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", err
	}

	encryptedSecret, err := encryptMFASecret(key.Secret())
	if err != nil {
		return "", "", err
	}

	if err := s.repo.UpsertSecret(userID, encryptedSecret); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *mfaService) ConfirmEnrollment(userID string, code string) ([]string, error) {
	row, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if row.EnabledAt != nil {
		return nil, ErrMFAAlreadyEnabled
	}

	secret, err := decryptMFASecret(row.SecretEncrypted)
	if err != nil {
		return nil, err
	}
	if !totp.Validate(strings.TrimSpace(code), secret) {
		return nil, ErrMFAInvalidCode
	}

	if err := s.repo.SetEnabled(userID, time.Now()); err != nil {
		return nil, err
	}

	rawCodes, hashes, err := generateRecoveryCodes(mfaRecoveryCodeCount)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceRecoveryCodes(userID, hashes); err != nil {
		return nil, err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.mfa.enrolled", "user", &userID, domain.LogSeverityMedium, map[string]any{
		"user_id": userID,
	})

	return rawCodes, nil
}

func (s *mfaService) IsEnabled(userID string) (bool, error) {
	row, err := s.repo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrMFANotEnrolled) {
			return false, nil
		}
		return false, err
	}
	return row.EnabledAt != nil, nil
}

func (s *mfaService) VerifyCode(userID string, code string) error {
	row, err := s.repo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if row.EnabledAt == nil {
		return repository.ErrMFANotEnrolled
	}

	secret, err := decryptMFASecret(row.SecretEncrypted)
	if err != nil {
		return err
	}
	if !totp.Validate(strings.TrimSpace(code), secret) {
		return ErrMFAInvalidCode
	}
	return nil
}

func (s *mfaService) VerifyRecoveryCode(userID string, code string) error {
	return s.repo.ConsumeRecoveryCodeByHash(userID, hashRecoveryCode(code))
}

func (s *mfaService) IssuePreAuthToken(userID string, purpose string, ttl time.Duration) (string, error) {
	rawToken, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return "", err
	}
	record := &domain.MFAPreAuthToken{
		UserID:    userID,
		TokenHash: tokenHash,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := s.repo.CreatePreAuthToken(record); err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *mfaService) ResolvePreAuthToken(rawToken string) (*domain.MFAPreAuthToken, error) {
	hash, err := hashOpaqueToken(rawToken)
	if err != nil {
		return nil, err
	}
	return s.repo.FindValidPreAuthTokenByHash(hash, time.Now())
}

func (s *mfaService) ConsumePreAuthToken(id string) error {
	return s.repo.ConsumePreAuthTokenByID(id)
}

// generateOpaqueToken/hashOpaqueToken follow the exact same shape already
// used by password reset tokens, refresh tokens, and WS tickets in this
// codebase (32-byte crypto/rand, base64 raw-url-encoded, SHA-256 hex hash).
// Worth centralizing into one shared helper at some point — this is now
// the fourth near-identical copy — but not a rework this change should
// take on.
func generateOpaqueToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func hashOpaqueToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", repository.ErrMFAPreAuthInvalid
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func generateRecoveryCodes(count int) ([]string, []string, error) {
	raw := make([]string, 0, count)
	hashes := make([]string, 0, count)
	for range count {
		code, err := generateRecoveryCode()
		if err != nil {
			return nil, nil, err
		}
		raw = append(raw, code)
		hashes = append(hashes, hashRecoveryCode(code))
	}
	return raw, hashes, nil
}

func generateRecoveryCode() (string, error) {
	const length = 10
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	chars := make([]byte, length)
	for i, b := range buf {
		chars[i] = recoveryCodeAlphabet[int(b)%len(recoveryCodeAlphabet)]
	}
	return string(chars[:5]) + "-" + string(chars[5:]), nil
}

func hashRecoveryCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
