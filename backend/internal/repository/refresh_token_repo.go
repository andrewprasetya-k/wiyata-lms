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

	FindValidByTokenHash(tokenHash string, now time.Time) (*domain.RefreshToken, error)
	FindByTokenHash(tokenHash string) (*domain.RefreshToken, error)

	Rotate(oldTokenHash string, newToken *domain.RefreshToken) (*domain.RefreshToken, error)
	RevokeFamily(familyID string) error
	RevokeByTokenHash(tokenHash string) error
	FindByID(id string) (*domain.RefreshToken, error)
	FindActiveByUserID(userID string, now time.Time) ([]*domain.RefreshToken, error)
	RevokeByID(id string) error
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
			// Only a token that was revoked because it was already rotated
			// (superseded by a newer one) counts as reuse — a token revoked
			// deliberately (user_revoked/logout) is an expected, benign
			// end-of-session, not a theft/replay signal. Anything else
			// (including a missing/unrecognized reason, e.g. a row from
			// before this column existed) fails closed as reuse, since
			// that's the safer default on a security-sensitive path.
			if old.RevokedReason != nil &&
				(*old.RevokedReason == domain.RefreshTokenRevokedReasonUserRevoked ||
					*old.RevokedReason == domain.RefreshTokenRevokedReasonLogout) {
				return ErrRefreshTokenInvalid
			}
			return &ReusedRefreshTokenError{FamilyID: old.FamilyID, UserID: old.UserID}
		}
		if time.Now().After(old.ExpiresAt) {
			return ErrRefreshTokenInvalid
		}

		if err := tx.Model(&domain.RefreshToken{}).
			Where("rft_id = ?", old.ID).
			Updates(map[string]any{
				"rft_revoked_at":     time.Now(),
				"rft_revoked_reason": domain.RefreshTokenRevokedReasonRotated,
			}).Error; err != nil {
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
		Updates(map[string]any{
			"rft_revoked_at":     time.Now(),
			"rft_revoked_reason": domain.RefreshTokenRevokedReasonReuseDetected,
		}).Error
}

func (r *refreshTokenRepository) RevokeByTokenHash(tokenHash string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("rft_token_hash = ? AND rft_revoked_at IS NULL", tokenHash).
		Updates(map[string]any{
			"rft_revoked_at":     time.Now(),
			"rft_revoked_reason": domain.RefreshTokenRevokedReasonLogout,
		}).Error
}

func (r *refreshTokenRepository) FindByID(id string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.Where("rft_id = ?", id).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) FindActiveByUserID(userID string, now time.Time) ([]*domain.RefreshToken, error) {
	var tokens []*domain.RefreshToken
	err := r.db.
		Where("rft_usr_id = ? AND rft_revoked_at IS NULL AND rft_expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

func (r *refreshTokenRepository) RevokeByID(id string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("rft_id = ? AND rft_revoked_at IS NULL", id).
		Updates(map[string]any{
			"rft_revoked_at":     time.Now(),
			"rft_revoked_reason": domain.RefreshTokenRevokedReasonUserRevoked,
		}).Error
}
