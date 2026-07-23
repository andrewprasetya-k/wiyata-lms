package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// ─── stubs ─────────────────────────────────────────────────────────────────

type authTestUserRepoStub struct {
	created *domain.User
	byID    map[string]*domain.User
}

func newAuthTestUserRepoStub() *authTestUserRepoStub {
	return &authTestUserRepoStub{byID: map[string]*domain.User{}}
}

func (s *authTestUserRepoStub) Create(user *domain.User) error {
	user.ID = "user-1"
	s.created = user
	s.byID[user.ID] = user
	return nil
}
func (s *authTestUserRepoStub) FindAll(_ string, _ int, _ int) ([]*domain.User, int64, error) {
	return nil, 0, nil
}
func (s *authTestUserRepoStub) GetByID(id string) (*domain.User, error) {
	if u, ok := s.byID[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (s *authTestUserRepoStub) GetByEmail(email string) (*domain.User, error) {
	if s.created != nil && s.created.Email == email {
		return s.created, nil
	}
	return nil, errors.New("not found")
}
func (s *authTestUserRepoStub) Update(user *domain.User) error {
	s.byID[user.ID] = user
	if s.created != nil && s.created.ID == user.ID {
		s.created = user
	}
	return nil
}
func (s *authTestUserRepoStub) Delete(_ string) error { return nil }
func (s *authTestUserRepoStub) CheckEmailExists(_ string, _ string) (bool, error) {
	return s.created != nil, nil
}

type authTestSchoolUserRepoStub struct{}

func (s *authTestSchoolUserRepoStub) Create(_ *domain.SchoolUser) error { return nil }
func (s *authTestSchoolUserRepoStub) GetBySchool(_ string, _ string, _ int, _ int) ([]*domain.SchoolUser, int64, error) {
	return nil, 0, nil
}
func (s *authTestSchoolUserRepoStub) GetBySchoolWithDeleted(_ string, _ string, _ string, _ bool, _ int, _ int) ([]*domain.SchoolUser, int64, error) {
	return nil, 0, nil
}
func (s *authTestSchoolUserRepoStub) GetByUser(_ string) ([]*domain.SchoolUser, error) {
	return nil, nil
}
func (s *authTestSchoolUserRepoStub) Delete(_ string) error                             { return nil }
func (s *authTestSchoolUserRepoStub) SoftDeleteByIDInSchool(_ string, _ string) error   { return nil }
func (s *authTestSchoolUserRepoStub) RestoreByIDInSchool(_ string, _ string) error      { return nil }
func (s *authTestSchoolUserRepoStub) IsEnrolled(_ string, _ string) (bool, error)       { return false, nil }
func (s *authTestSchoolUserRepoStub) BelongsToSchool(_ string, _ string) (bool, error)  { return false, nil }
func (s *authTestSchoolUserRepoStub) FindByUserAndSchoolIncludingDeleted(_ string, _ string) (*domain.SchoolUser, error) {
	return nil, nil
}
func (s *authTestSchoolUserRepoStub) WithTx(_ *gorm.DB) repository.SchoolUserRepository {
	return s
}

type authTestLogServiceStub struct{}

func (s *authTestLogServiceStub) Record(_ *domain.Log) error { return nil }
func (s *authTestLogServiceStub) GetBySchool(_ string, _ int, _ int) ([]*domain.Log, int64, error) {
	return nil, 0, nil
}
func (s *authTestLogServiceStub) GetByUser(_ string, _ int, _ int) ([]*domain.Log, int64, error) {
	return nil, 0, nil
}
func (s *authTestLogServiceStub) GetByCorrelationID(_ string) ([]*domain.Log, error) { return nil, nil }
func (s *authTestLogServiceStub) Log(_ domain.ActorContext, _ string, _ string, _ *string, _ string, _ any) error {
	return nil
}
func (s *authTestLogServiceStub) LogBatch(_ *gorm.DB, _ domain.ActorContext, _ string, _ string, _ *string, _ string, _ any, _ []LogBatchChild) error {
	return nil
}

type authTestRefreshTokenRepoStub struct{}

func (s *authTestRefreshTokenRepoStub) Create(_ *domain.RefreshToken) error { return nil }
func (s *authTestRefreshTokenRepoStub) FindValidByTokenHash(_ string, _ time.Time) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (s *authTestRefreshTokenRepoStub) FindByTokenHash(_ string) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (s *authTestRefreshTokenRepoStub) Rotate(_ string, _ *domain.RefreshToken) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (s *authTestRefreshTokenRepoStub) RevokeFamily(_ string) error      { return nil }
func (s *authTestRefreshTokenRepoStub) RevokeByTokenHash(_ string) error { return nil }
func (s *authTestRefreshTokenRepoStub) FindByID(_ string) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (s *authTestRefreshTokenRepoStub) FindActiveByUserID(_ string, _ time.Time) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (s *authTestRefreshTokenRepoStub) RevokeByID(_ string) error { return nil }

type authTestMFAServiceStub struct {
	enabled       bool
	validCode     string
	preAuthTokens map[string]*domain.MFAPreAuthToken
	consumed      map[string]bool
	enrolledUsers map[string]bool
}

func newAuthTestMFAServiceStub() *authTestMFAServiceStub {
	return &authTestMFAServiceStub{
		validCode:     "123456",
		preAuthTokens: map[string]*domain.MFAPreAuthToken{},
		consumed:      map[string]bool{},
		enrolledUsers: map[string]bool{},
	}
}

func (s *authTestMFAServiceStub) Enroll(userID string) (string, string, error) {
	if s.enrolledUsers[userID] {
		return "", "", ErrMFAAlreadyEnabled
	}
	return "SECRETSECRET", "otpauth://totp/Wiyata:user?secret=SECRETSECRET&issuer=Wiyata", nil
}
func (s *authTestMFAServiceStub) ConfirmEnrollment(userID string, code string) ([]string, error) {
	if s.enrolledUsers[userID] {
		return nil, ErrMFAAlreadyEnabled
	}
	if code != s.validCode {
		return nil, ErrMFAInvalidCode
	}
	s.enrolledUsers[userID] = true
	return []string{"AAAAA-BBBBB"}, nil
}
func (s *authTestMFAServiceStub) IsEnabled(userID string) (bool, error) { return s.enabled, nil }
func (s *authTestMFAServiceStub) VerifyCode(_ string, _ string) error   { return nil }
func (s *authTestMFAServiceStub) VerifyRecoveryCode(_ string, _ string) error {
	return nil
}
func (s *authTestMFAServiceStub) IssuePreAuthToken(userID string, purpose string, ttl time.Duration) (string, error) {
	raw := "raw-" + userID + "-" + purpose
	s.preAuthTokens[raw] = &domain.MFAPreAuthToken{
		ID:        raw,
		UserID:    userID,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
	}
	return raw, nil
}
func (s *authTestMFAServiceStub) ResolvePreAuthToken(rawToken string) (*domain.MFAPreAuthToken, error) {
	token, ok := s.preAuthTokens[rawToken]
	if !ok || s.consumed[rawToken] {
		return nil, errors.New("not found")
	}
	return token, nil
}
func (s *authTestMFAServiceStub) ConsumePreAuthToken(id string) error {
	s.consumed[id] = true
	return nil
}

// ─── tests ─────────────────────────────────────────────────────────────────

func TestRegister_FreshUser_NeverHitsMFABranches(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	svc := &authService{
		userRepo:       newAuthTestUserRepoStub(),
		schoolUserRepo: &authTestSchoolUserRepoStub{},
		logService:     &authTestLogServiceStub{},
		refreshTokenRepo: &authTestRefreshTokenRepoStub{},
		mfaService:       newAuthTestMFAServiceStub(),
	}

	result, err := svc.Register("Fresh User", "fresh@example.com", "Passw0rd!", RefreshTokenMetadata{})
	if err != nil {
		t.Fatalf("Register returned unexpected error: %v", err)
	}

	if result.Challenge != nil {
		t.Fatalf("fresh registration must never receive an MFA challenge, got: %+v", result.Challenge)
	}
	if result.Response == nil {
		t.Fatalf("expected a completed login response, got nil")
	}
	if result.Response.MFAGraceDaysRemaining == nil {
		t.Fatalf("expected MFAGraceDaysRemaining to be set for a user newly entering the grace period")
	}
	if *result.Response.MFAGraceDaysRemaining != 7 {
		t.Fatalf("expected a fresh grace period of 7 days, got %d", *result.Response.MFAGraceDaysRemaining)
	}
	if result.RawRefreshToken == "" {
		t.Fatalf("expected a raw refresh token to be issued")
	}
}

func TestCheckMFAGracePeriod_NilStart_StartsClockNotExpired(t *testing.T) {
	svc := &authService{userRepo: newAuthTestUserRepoStub()}
	user := &domain.User{ID: "user-1", MFAGraceStartedAt: nil}
	svc.userRepo.(*authTestUserRepoStub).byID[user.ID] = user

	days, expired, err := svc.checkMFAGracePeriod(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expired {
		t.Fatalf("a nil grace start must never be reported as expired")
	}
	if days == nil || *days != 7 {
		t.Fatalf("expected 7 days remaining right after starting the clock, got %v", days)
	}
	if user.MFAGraceStartedAt == nil {
		t.Fatalf("expected MFAGraceStartedAt to be persisted onto the user")
	}
}

func TestCheckMFAGracePeriod_WithinWindow_NotExpired(t *testing.T) {
	svc := &authService{userRepo: newAuthTestUserRepoStub()}
	started := time.Now().Add(-3 * 24 * time.Hour)
	user := &domain.User{ID: "user-1", MFAGraceStartedAt: &started}

	days, expired, err := svc.checkMFAGracePeriod(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expired {
		t.Fatalf("3 days into a 7-day window must not be expired")
	}
	if days == nil || *days <= 0 || *days > 4 {
		t.Fatalf("expected roughly 4 days remaining, got %v", days)
	}
}

func TestCheckMFAGracePeriod_PastWindow_Expired(t *testing.T) {
	svc := &authService{userRepo: newAuthTestUserRepoStub()}
	started := time.Now().Add(-8 * 24 * time.Hour)
	user := &domain.User{ID: "user-1", MFAGraceStartedAt: &started}

	_, expired, err := svc.checkMFAGracePeriod(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !expired {
		t.Fatalf("8 days past a 7-day window must be expired")
	}
}

// ─── forced MFA setup via preAuthToken (grace-period-expired flow) ────────

func newAuthServiceForMFASetupTests(userID string) (*authService, *authTestMFAServiceStub) {
	userRepo := newAuthTestUserRepoStub()
	userRepo.byID[userID] = &domain.User{ID: userID, Email: "user@example.com"}

	mfaService := newAuthTestMFAServiceStub()
	svc := &authService{
		userRepo:         userRepo,
		schoolUserRepo:   &authTestSchoolUserRepoStub{},
		logService:       &authTestLogServiceStub{},
		refreshTokenRepo: &authTestRefreshTokenRepoStub{},
		mfaService:       mfaService,
	}
	return svc, mfaService
}

func TestEnrollMFAViaPreAuthToken_RejectsVerifyPurposeToken(t *testing.T) {
	svc, mfaService := newAuthServiceForMFASetupTests("user-1")
	// Simulate the token a user who ALREADY has MFA enabled would receive
	// on login (MFARequired, not MFASetupRequired) — must never be usable
	// to (re)start enrollment.
	raw, _ := mfaService.IssuePreAuthToken("user-1", domain.MFAPreAuthPurposeVerify, mfaPreAuthTokenTTL)

	_, _, err := svc.EnrollMFAViaPreAuthToken(raw)
	if !errors.Is(err, ErrMFAPreAuthInvalid) {
		t.Fatalf("expected ErrMFAPreAuthInvalid for a mfa_verify-purpose token, got %v", err)
	}
}

func TestEnrollMFAViaPreAuthToken_AcceptsEnrollRequiredPurposeToken(t *testing.T) {
	svc, mfaService := newAuthServiceForMFASetupTests("user-1")
	raw, _ := mfaService.IssuePreAuthToken("user-1", domain.MFAPreAuthPurposeEnrollRequired, mfaPreAuthTokenTTL)

	secret, otpauthURL, err := svc.EnrollMFAViaPreAuthToken(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret == "" || otpauthURL == "" {
		t.Fatalf("expected a non-empty secret and otpauth URL")
	}
	if mfaService.consumed[raw] {
		t.Fatalf("EnrollMFAViaPreAuthToken must not consume the token — it's reused by CompleteMFASetup")
	}
}

func TestCompleteMFASetup_RejectsVerifyPurposeToken(t *testing.T) {
	svc, mfaService := newAuthServiceForMFASetupTests("user-1")
	raw, _ := mfaService.IssuePreAuthToken("user-1", domain.MFAPreAuthPurposeVerify, mfaPreAuthTokenTTL)

	_, _, err := svc.CompleteMFASetup(raw, "123456", RefreshTokenMetadata{})
	if !errors.Is(err, ErrMFAPreAuthInvalid) {
		t.Fatalf("expected ErrMFAPreAuthInvalid for a mfa_verify-purpose token, got %v", err)
	}
}

func TestCompleteMFASetup_SuccessIssuesSessionAndConsumesToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	svc, mfaService := newAuthServiceForMFASetupTests("user-1")
	raw, _ := mfaService.IssuePreAuthToken("user-1", domain.MFAPreAuthPurposeEnrollRequired, mfaPreAuthTokenTTL)

	response, rawRefreshToken, err := svc.CompleteMFASetup(raw, "123456", RefreshTokenMetadata{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.RecoveryCodes) == 0 {
		t.Fatalf("expected recovery codes to be returned")
	}
	if response.Token == "" {
		t.Fatalf("expected a real access token to be issued")
	}
	if rawRefreshToken == "" {
		t.Fatalf("expected a real refresh token to be issued")
	}

	// Single-use: the same token must not work a second time.
	if _, _, err := svc.CompleteMFASetup(raw, "123456", RefreshTokenMetadata{}); !errors.Is(err, ErrMFAPreAuthInvalid) {
		t.Fatalf("expected the preAuthToken to be consumed after a successful setup, got %v", err)
	}
}

func TestCompleteMFASetup_WrongCodeDoesNotConsumeToken(t *testing.T) {
	svc, mfaService := newAuthServiceForMFASetupTests("user-1")
	raw, _ := mfaService.IssuePreAuthToken("user-1", domain.MFAPreAuthPurposeEnrollRequired, mfaPreAuthTokenTTL)

	_, _, err := svc.CompleteMFASetup(raw, "000000", RefreshTokenMetadata{})
	if !errors.Is(err, ErrMFACodeInvalid) {
		t.Fatalf("expected ErrMFACodeInvalid for a wrong code, got %v", err)
	}
	if mfaService.consumed[raw] {
		t.Fatalf("a failed code attempt must not consume the preAuthToken")
	}
}
