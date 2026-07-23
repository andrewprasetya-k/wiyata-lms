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
	enabled bool
}

func (s *authTestMFAServiceStub) Enroll(_ string) (string, string, error)             { return "", "", nil }
func (s *authTestMFAServiceStub) ConfirmEnrollment(_ string, _ string) ([]string, error) {
	return nil, nil
}
func (s *authTestMFAServiceStub) IsEnabled(_ string) (bool, error)          { return s.enabled, nil }
func (s *authTestMFAServiceStub) VerifyCode(_ string, _ string) error       { return nil }
func (s *authTestMFAServiceStub) VerifyRecoveryCode(_ string, _ string) error { return nil }
func (s *authTestMFAServiceStub) IssuePreAuthToken(_ string, _ string, _ time.Duration) (string, error) {
	return "preauth-token", nil
}
func (s *authTestMFAServiceStub) ResolvePreAuthToken(_ string) (*domain.MFAPreAuthToken, error) {
	return nil, errors.New("not found")
}
func (s *authTestMFAServiceStub) ConsumePreAuthToken(_ string) error { return nil }

// ─── tests ─────────────────────────────────────────────────────────────────

func TestRegister_FreshUser_NeverHitsMFABranches(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	svc := &authService{
		userRepo:       newAuthTestUserRepoStub(),
		schoolUserRepo: &authTestSchoolUserRepoStub{},
		logService:     &authTestLogServiceStub{},
		refreshTokenRepo: &authTestRefreshTokenRepoStub{},
		mfaService:       &authTestMFAServiceStub{enabled: false},
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
