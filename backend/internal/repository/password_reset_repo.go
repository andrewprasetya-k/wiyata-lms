package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPasswordResetInvalid = errors.New("password reset token is invalid or expired")

type PasswordResetRepository interface {
	Create(reset *domain.PasswordResetToken) error
	InvalidateAllForUser(userID string) error
	// GetValidByTokenHash is a non-consuming lookup, used by the metadata
	// check (GET /reset-password/:token) so the frontend can show a valid
	// form vs. an error state before the user submits anything.
	GetValidByTokenHash(tokenHash string, now time.Time) (*domain.PasswordResetToken, error)
	// ConsumeAndSetPassword atomically validates the token, writes the new
	// password hash onto the user row, and marks the token (and any other
	// outstanding tokens for the same user) consumed — all in one
	// transaction, so a reset can never partially apply.
	ConsumeAndSetPassword(tokenHash string, newPasswordHash string, now time.Time) (*domain.PasswordResetToken, error)
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(reset *domain.PasswordResetToken) error {
	return r.db.Create(reset).Error
}

func (r *passwordResetRepository) InvalidateAllForUser(userID string) error {
	return r.db.Model(&domain.PasswordResetToken{}).
		Where("prt_usr_id = ? AND prt_consumed_at IS NULL", userID).
		Update("prt_consumed_at", time.Now()).Error
}

func (r *passwordResetRepository) GetValidByTokenHash(tokenHash string, now time.Time) (*domain.PasswordResetToken, error) {
	var reset domain.PasswordResetToken
	err := r.db.Where("prt_token_hash = ?", tokenHash).First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPasswordResetInvalid
		}
		return nil, err
	}
	if reset.ConsumedAt != nil || now.After(reset.ExpiresAt) {
		return nil, ErrPasswordResetInvalid
	}
	return &reset, nil
}

func (r *passwordResetRepository) ConsumeAndSetPassword(tokenHash string, newPasswordHash string, now time.Time) (*domain.PasswordResetToken, error) {
	var result *domain.PasswordResetToken

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var reset domain.PasswordResetToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("prt_token_hash = ?", tokenHash).
			First(&reset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPasswordResetInvalid
			}
			return err
		}
		if reset.ConsumedAt != nil || now.After(reset.ExpiresAt) {
			return ErrPasswordResetInvalid
		}

		if err := tx.Model(&domain.PasswordResetToken{}).
			Where("prt_usr_id = ? AND prt_consumed_at IS NULL", reset.UserID).
			Update("prt_consumed_at", now).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.User{}).
			Where("usr_id = ?", reset.UserID).
			Update("usr_password", newPasswordHash).Error; err != nil {
			return err
		}

		reset.ConsumedAt = &now
		result = &reset
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
