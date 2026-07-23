package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrMFANotEnrolled     = errors.New("mfa is not set up for this user")
	ErrMFARecoveryInvalid = errors.New("recovery code is invalid or already used")
	ErrMFAPreAuthInvalid  = errors.New("mfa pre-auth token is invalid or expired")
)

// MFARepository bundles the three MFA-related tables (user_mfa,
// mfa_recovery_codes, mfa_preauth_tokens) — they're always used together by
// MFAService and don't warrant three separate repositories.
type MFARepository interface {
	GetByUserID(userID string) (*domain.UserMFA, error)
	UpsertSecret(userID string, encryptedSecret string) error
	// SetEnabled marks enrollment complete.
	SetEnabled(userID string, now time.Time) error
	ReplaceRecoveryCodes(userID string, codeHashes []string) error
	ConsumeRecoveryCodeByHash(userID string, codeHash string) error

	CreatePreAuthToken(token *domain.MFAPreAuthToken) error
	FindValidPreAuthTokenByHash(tokenHash string, now time.Time) (*domain.MFAPreAuthToken, error)
	ConsumePreAuthTokenByID(id string) error
}

type mfaRepository struct {
	db *gorm.DB
}

func NewMFARepository(db *gorm.DB) MFARepository {
	return &mfaRepository{db: db}
}

func (r *mfaRepository) GetByUserID(userID string) (*domain.UserMFA, error) {
	var row domain.UserMFA
	err := r.db.Where("umf_usr_id = ?", userID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMFANotEnrolled
		}
		return nil, err
	}
	return &row, nil
}

func (r *mfaRepository) UpsertSecret(userID string, encryptedSecret string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing domain.UserMFA
		err := tx.Where("umf_usr_id = ?", userID).First(&existing).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return tx.Create(&domain.UserMFA{
				UserID:          userID,
				SecretEncrypted: encryptedSecret,
			}).Error
		}

		if existing.EnabledAt != nil {
			return errors.New("mfa is already enabled for this user")
		}
		return tx.Model(&domain.UserMFA{}).
			Where("umf_id = ?", existing.ID).
			Update("umf_secret_encrypted", encryptedSecret).Error
	})
}

func (r *mfaRepository) SetEnabled(userID string, now time.Time) error {
	return r.db.Model(&domain.UserMFA{}).
		Where("umf_usr_id = ?", userID).
		Update("umf_enabled_at", now).Error
}

func (r *mfaRepository) ReplaceRecoveryCodes(userID string, codeHashes []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("mrc_usr_id = ?", userID).Delete(&domain.MFARecoveryCode{}).Error; err != nil {
			return err
		}
		rows := make([]*domain.MFARecoveryCode, 0, len(codeHashes))
		for _, hash := range codeHashes {
			rows = append(rows, &domain.MFARecoveryCode{UserID: userID, CodeHash: hash})
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func (r *mfaRepository) ConsumeRecoveryCodeByHash(userID string, codeHash string) error {
	result := r.db.Model(&domain.MFARecoveryCode{}).
		Where("mrc_usr_id = ? AND mrc_code_hash = ? AND mrc_consumed_at IS NULL", userID, codeHash).
		Update("mrc_consumed_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMFARecoveryInvalid
	}
	return nil
}

func (r *mfaRepository) CreatePreAuthToken(token *domain.MFAPreAuthToken) error {
	return r.db.Create(token).Error
}

func (r *mfaRepository) FindValidPreAuthTokenByHash(tokenHash string, now time.Time) (*domain.MFAPreAuthToken, error) {
	var token domain.MFAPreAuthToken
	err := r.db.Where("mpt_token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMFAPreAuthInvalid
		}
		return nil, err
	}
	if token.ConsumedAt != nil || now.After(token.ExpiresAt) {
		return nil, ErrMFAPreAuthInvalid
	}
	return &token, nil
}

func (r *mfaRepository) ConsumePreAuthTokenByID(id string) error {
	return r.db.Model(&domain.MFAPreAuthToken{}).
		Where("mpt_id = ? AND mpt_consumed_at IS NULL", id).
		Update("mpt_consumed_at", time.Now()).Error
}
