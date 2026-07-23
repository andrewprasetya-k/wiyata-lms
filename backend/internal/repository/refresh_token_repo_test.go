package repository

import (
	"backend/internal/domain"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newRefreshTokenRepoTestRepo(t *testing.T) (*refreshTokenRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm on top of sqlmock: %v", err)
	}

	return &refreshTokenRepository{db: gormDB}, mock, func() { sqlDB.Close() }
}

// refreshTokenRow builds a single already-revoked row (revoked 1 minute
// ago, still otherwise unexpired) with the given revoked_reason — nil
// means the column is NULL, simulating a pre-migration row.
func refreshTokenRow(revokedReason *string) *sqlmock.Rows {
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"rft_id", "rft_usr_id", "rft_token_hash", "rft_family_id",
		"rft_expires_at", "rft_revoked_at", "rft_revoked_reason",
		"rft_user_agent", "rft_ip_address", "created_at", "updated_at",
	})
	var reasonValue any
	if revokedReason != nil {
		reasonValue = *revokedReason
	}
	rows.AddRow(
		"token-1", "user-1", "hash-1", "family-1",
		now.Add(24*time.Hour), now.Add(-time.Minute), reasonValue,
		nil, nil, now.Add(-time.Hour), now.Add(-time.Minute),
	)
	return rows
}

// TestRotateTreatsRotatedReasonAsReuse confirms the actual reuse-detection
// path (the whole point of this feature) still fires for the genuine case:
// a token that was revoked because it was already exchanged for a newer
// one (reason "rotated"), then presented again.
func TestRotateTreatsRotatedReasonAsReuse(t *testing.T) {
	repo, mock, closeDB := newRefreshTokenRepoTestRepo(t)
	defer closeDB()

	reason := domain.RefreshTokenRevokedReasonRotated
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "edv"."refresh_tokens"`).
		WillReturnRows(refreshTokenRow(&reason))
	mock.ExpectRollback()

	_, err := repo.Rotate("hash-1", &domain.RefreshToken{})

	var reused *ReusedRefreshTokenError
	if !errors.As(err, &reused) {
		t.Fatalf("expected *ReusedRefreshTokenError for reason=%q, got: %v", reason, err)
	}
	if reused.FamilyID != "family-1" || reused.UserID != "user-1" {
		t.Fatalf("unexpected reused token details: %+v", reused)
	}
}

// TestRotateDoesNotTreatUserRevokedAsReuse is the false-positive this whole
// fix targets: a session ended deliberately via DELETE /me/sessions/:id
// should not trip reuse-detection when its (now-revoked) token is
// presented again by the browser.
func TestRotateDoesNotTreatUserRevokedAsReuse(t *testing.T) {
	repo, mock, closeDB := newRefreshTokenRepoTestRepo(t)
	defer closeDB()

	reason := domain.RefreshTokenRevokedReasonUserRevoked
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "edv"."refresh_tokens"`).
		WillReturnRows(refreshTokenRow(&reason))
	mock.ExpectRollback()

	_, err := repo.Rotate("hash-1", &domain.RefreshToken{})

	if !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("expected ErrRefreshTokenInvalid for reason=%q, got: %v", reason, err)
	}
	var reused *ReusedRefreshTokenError
	if errors.As(err, &reused) {
		t.Fatalf("did not expect reuse-detection for a user_revoked token, got: %+v", reused)
	}
}

// TestRotateDoesNotTreatLogoutAsReuse mirrors the user_revoked case for the
// other benign revocation cause: an explicit POST /logout.
func TestRotateDoesNotTreatLogoutAsReuse(t *testing.T) {
	repo, mock, closeDB := newRefreshTokenRepoTestRepo(t)
	defer closeDB()

	reason := domain.RefreshTokenRevokedReasonLogout
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "edv"."refresh_tokens"`).
		WillReturnRows(refreshTokenRow(&reason))
	mock.ExpectRollback()

	_, err := repo.Rotate("hash-1", &domain.RefreshToken{})

	if !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("expected ErrRefreshTokenInvalid for reason=%q, got: %v", reason, err)
	}
}

// TestRotateTreatsMissingReasonAsReuseFailClosed confirms the fail-closed
// default: a revoked row with no recognizable reason (e.g. a row from
// before this column existed, or any future revocation path that forgets
// to set one) still trips reuse-detection rather than silently passing.
func TestRotateTreatsMissingReasonAsReuseFailClosed(t *testing.T) {
	repo, mock, closeDB := newRefreshTokenRepoTestRepo(t)
	defer closeDB()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "edv"."refresh_tokens"`).
		WillReturnRows(refreshTokenRow(nil))
	mock.ExpectRollback()

	_, err := repo.Rotate("hash-1", &domain.RefreshToken{})

	var reused *ReusedRefreshTokenError
	if !errors.As(err, &reused) {
		t.Fatalf("expected fail-closed reuse-detection for a nil reason, got: %v", err)
	}
}
