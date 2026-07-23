package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")
)

// ReusedRefreshTokenError is returned by Rotate when the presented token has
// already been rotated once before (its row exists but is already revoked)
// — i.e. someone is presenting a token that was already exchanged for a
// newer one. FamilyID/UserID let the caller act on the whole session family,
// not just this one token.
type ReusedRefreshTokenError struct {
	FamilyID string
	UserID   string
}

func (e *ReusedRefreshTokenError) Error() string {
	return "refresh token reuse detected"
}

type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) error

	// FindValidByTokenHash is a non-mutating lookup that only returns a row
	// currently valid (not revoked, not expired) — mirrors
	// PasswordResetRepository.GetValidByTokenHash's shape.
	FindValidByTokenHash(tokenHash string, now time.Time) (*domain.RefreshToken, error)

	// FindByTokenHash is unfiltered — it returns the row regardless of
	// revoked/expired status. Needed because reuse-detection and rate-limit
	// key resolution both need a token's family_id even when the token
	// itself is no longer valid.
	FindByTokenHash(tokenHash string) (*domain.RefreshToken, error)

	// Rotate atomically: locks the row by tokenHash, and
	//   - if not found: returns ErrRefreshTokenInvalid
	//   - if already revoked: returns *ReusedRefreshTokenError (reuse of an
	//     already-rotated token)
	//   - if expired (but not revoked): returns ErrRefreshTokenInvalid
	//   - otherwise: marks the old row revoked and inserts newToken with the
	//     same UserID/FamilyID as the old row, all in one transaction.
	Rotate(oldTokenHash string, newToken *domain.RefreshToken) (*domain.RefreshToken, error)

	// RevokeFamily marks every not-yet-revoked token sharing familyID as
	// revoked — the reuse-detection response, ending the whole session.
	RevokeFamily(familyID string) error

	// RevokeByTokenHash revokes a single token (logout — only this session
	// ends, siblings in the same family are untouched).
	RevokeByTokenHash(tokenHash string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindValidByTokenHash(tokenHash string, now time.Time) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.Where("rft_token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	if token.RevokedAt != nil || now.After(token.ExpiresAt) {
		return nil, ErrRefreshTokenInvalid
	}
	return &token, nil
}

func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.Where("rft_token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) Rotate(oldTokenHash string, newToken *domain.RefreshToken) (*domain.RefreshToken, error) {
	var result *domain.RefreshToken

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var old domain.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("rft_token_hash = ?", oldTokenHash).
			First(&old).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRefreshTokenInvalid
			}
			return err
		}

		if old.RevokedAt != nil {
			return &ReusedRefreshTokenError{FamilyID: old.FamilyID, UserID: old.UserID}
		}
		if time.Now().After(old.ExpiresAt) {
			return ErrRefreshTokenInvalid
		}

		if err := tx.Model(&domain.RefreshToken{}).
			Where("rft_id = ?", old.ID).
			Update("rft_revoked_at", time.Now()).Error; err != nil {
			return err
		}

		newToken.UserID = old.UserID
		newToken.FamilyID = old.FamilyID
		if err := tx.Create(newToken).Error; err != nil {
			return err
		}
		result = newToken
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *refreshTokenRepository) RevokeFamily(familyID string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("rft_family_id = ? AND rft_revoked_at IS NULL", familyID).
		Update("rft_revoked_at", time.Now()).Error
}

func (r *refreshTokenRepository) RevokeByTokenHash(tokenHash string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("rft_token_hash = ? AND rft_revoked_at IS NULL", tokenHash).
		Update("rft_revoked_at", time.Now()).Error
}
