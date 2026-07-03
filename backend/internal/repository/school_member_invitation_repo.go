package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSchoolMemberInvitationNotFound = errors.New("school member invitation not found")
var ErrSchoolMemberInvitationNotRevocable = errors.New("school member invitation cannot be revoked")

type SchoolMemberInvitationRepository interface {
	Create(invitation *domain.Invitation) error
	FindSchoolByID(schoolID string) (*domain.School, error)
	FindClassByCode(schoolID string, classCode string) (*domain.Class, error)
	HasPendingDuplicate(schoolID string, email string, role string, now time.Time) (bool, error)
	List(schoolID string, status string, page int, limit int, now time.Time) ([]domain.Invitation, int64, error)
	Revoke(schoolID string, invitationID string, now time.Time) (*domain.Invitation, error)
}

type schoolMemberInvitationRepository struct {
	db *gorm.DB
}

func NewSchoolMemberInvitationRepository(db *gorm.DB) SchoolMemberInvitationRepository {
	return &schoolMemberInvitationRepository{db: db}
}

func (r *schoolMemberInvitationRepository) Create(invitation *domain.Invitation) error {
	return r.db.Create(invitation).Error
}

func (r *schoolMemberInvitationRepository) FindSchoolByID(schoolID string) (*domain.School, error) {
	var school domain.School
	if err := r.db.Where("sch_id = ?", schoolID).First(&school).Error; err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *schoolMemberInvitationRepository) FindClassByCode(schoolID string, classCode string) (*domain.Class, error) {
	var class domain.Class
	if err := r.db.
		Where("cls_sch_id = ? AND cls_code = ? AND deleted_at IS NULL", schoolID, classCode).
		First(&class).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *schoolMemberInvitationRepository) HasPendingDuplicate(schoolID string, email string, role string, now time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Invitation{}).
		Where("inv_school_id = ? AND LOWER(inv_email) = LOWER(?) AND inv_role = ?", schoolID, email, role).
		Where("inv_accepted_at IS NULL AND inv_revoked_at IS NULL AND inv_expires_at > ?", now).
		Count(&count).Error
	return count > 0, err
}

func (r *schoolMemberInvitationRepository) List(schoolID string, status string, page int, limit int, now time.Time) ([]domain.Invitation, int64, error) {
	query := r.db.Model(&domain.Invitation{}).
		Preload("Class").
		Where("inv_school_id = ?", schoolID)
	query = applyInvitationStatusFilter(query, status, now)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var invitations []domain.Invitation
	offset := (page - 1) * limit
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invitations).Error; err != nil {
		return nil, 0, err
	}
	return invitations, total, nil
}

func (r *schoolMemberInvitationRepository) Revoke(schoolID string, invitationID string, now time.Time) (*domain.Invitation, error) {
	var invitation domain.Invitation
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("inv_id = ? AND inv_school_id = ?", invitationID, schoolID).
			First(&invitation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSchoolMemberInvitationNotFound
			}
			return err
		}
		if invitation.AcceptedAt != nil || invitation.RevokedAt != nil || !now.Before(invitation.ExpiresAt) {
			return ErrSchoolMemberInvitationNotRevocable
		}
		if err := tx.Model(&domain.Invitation{}).
			Where("inv_id = ?", invitation.ID).
			Updates(map[string]any{
				"inv_revoked_at": now,
				"updated_at":     now,
			}).Error; err != nil {
			return err
		}
		invitation.RevokedAt = &now
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := r.db.Preload("Class").Where("inv_id = ?", invitation.ID).First(&invitation).Error; err != nil {
		return nil, err
	}
	return &invitation, nil
}

func applyInvitationStatusFilter(query *gorm.DB, status string, now time.Time) *gorm.DB {
	switch status {
	case "accepted":
		return query.Where("inv_accepted_at IS NOT NULL")
	case "revoked":
		return query.Where("inv_revoked_at IS NOT NULL")
	case "expired":
		return query.Where("inv_accepted_at IS NULL AND inv_revoked_at IS NULL AND inv_expires_at <= ?", now)
	default:
		return query.Where("inv_accepted_at IS NULL AND inv_revoked_at IS NULL AND inv_expires_at > ?", now)
	}
}
