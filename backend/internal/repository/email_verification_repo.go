package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrEmailVerificationInvalid = errors.New("verification token is invalid or expired")

type EmailVerificationRepository interface {
	Create(verification *domain.EmailVerification) error
	InvalidateAllForUser(userID string) error
	ConsumeByTokenHash(tokenHash string, now time.Time) (*domain.EmailVerification, error)
}

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(verification *domain.EmailVerification) error {
	return r.db.Create(verification).Error
}

func (r *emailVerificationRepository) InvalidateAllForUser(userID string) error {
	return r.db.Model(&domain.EmailVerification{}).
		Where("evf_usr_id = ? AND evf_consumed_at IS NULL", userID).
		Update("evf_consumed_at", time.Now()).Error
}

func (r *emailVerificationRepository) ConsumeByTokenHash(tokenHash string, now time.Time) (*domain.EmailVerification, error) {
	var result *domain.EmailVerification

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var verification domain.EmailVerification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("evf_token_hash = ?", tokenHash).
			First(&verification).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEmailVerificationInvalid
			}
			return err
		}
		if verification.ConsumedAt != nil || now.After(verification.ExpiresAt) {
			return ErrEmailVerificationInvalid
		}

		if err := tx.Model(&domain.EmailVerification{}).
			Where("evf_usr_id = ? AND evf_consumed_at IS NULL", verification.UserID).
			Update("evf_consumed_at", now).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.User{}).
			Where("usr_id = ?", verification.UserID).
			Update("usr_email_verified_at", now).Error; err != nil {
			return err
		}

		verification.ConsumedAt = &now
		result = &verification
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
