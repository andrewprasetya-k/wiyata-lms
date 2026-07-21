package repository

import (
	"backend/internal/domain"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvitationInvalid = errors.New("invitation is invalid or expired")
var ErrInvitationClassUnavailable = errors.New("invitation class is no longer available")
var ErrInvitationEmailMismatch = errors.New("invitation email does not match authenticated user")

type InvitationAcceptResult struct {
	Invitation domain.Invitation
	User       domain.User
	School     domain.School
	Role       string
}

type InvitationRepository interface {
	GetByTokenHash(tokenHash string) (*domain.Invitation, error)
	Accept(tokenHash string, name string, passwordHash string, now time.Time) (*InvitationAcceptResult, error)
	// AcceptAuthenticated accepts an invitation on behalf of an already
	// logged-in user (existing-account flow), instead of the name/password
	// registration path Accept uses. It verifies the caller's own account
	// email matches the invitation's email before attaching anything.
	AcceptAuthenticated(tokenHash string, userID string, now time.Time) (*InvitationAcceptResult, error)
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) GetByTokenHash(tokenHash string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	err := r.db.
		Preload("School").
		Where("inv_token_hash = ?", tokenHash).
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	if !isInvitationUsable(invitation, time.Now()) {
		return nil, ErrInvitationInvalid
	}
	if invitation.School.DeletedAt.Valid {
		return nil, ErrInvitationInvalid
	}
	return &invitation, nil
}

func (r *invitationRepository) Accept(tokenHash string, name string, passwordHash string, now time.Time) (*InvitationAcceptResult, error) {
	var result *InvitationAcceptResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		invitation, school, err := lockUsableInvitationWithSchool(tx, tokenHash, now)
		if err != nil {
			return err
		}

		user, err := resolveInvitationUser(tx, invitation.Email, name, passwordHash)
		if err != nil {
			return err
		}

		finalized, err := finalizeInvitationAcceptance(tx, *invitation, *school, user, now)
		if err != nil {
			return err
		}
		result = finalized
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *invitationRepository) AcceptAuthenticated(tokenHash string, userID string, now time.Time) (*InvitationAcceptResult, error) {
	var result *InvitationAcceptResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		invitation, school, err := lockUsableInvitationWithSchool(tx, tokenHash, now)
		if err != nil {
			return err
		}

		var user domain.User
		if err := tx.Where("usr_id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvitationInvalid
			}
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(invitation.Email)) {
			return ErrInvitationEmailMismatch
		}

		finalized, err := finalizeInvitationAcceptance(tx, *invitation, *school, &user, now)
		if err != nil {
			return err
		}
		result = finalized
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// lockUsableInvitationWithSchool loads the invitation row (row-locked, so
// concurrent accept attempts serialize) and its school, shared by both
// Accept and AcceptAuthenticated before they diverge on how they resolve
// the accepting user.
func lockUsableInvitationWithSchool(tx *gorm.DB, tokenHash string, now time.Time) (*domain.Invitation, *domain.School, error) {
	var invitation domain.Invitation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("inv_token_hash = ?", tokenHash).
		First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvitationInvalid
		}
		return nil, nil, err
	}
	if !isInvitationUsable(invitation, now) {
		return nil, nil, ErrInvitationInvalid
	}

	var school domain.School
	if err := tx.Where("sch_id = ?", invitation.SchoolID).First(&school).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvitationInvalid
		}
		return nil, nil, err
	}

	return &invitation, &school, nil
}

// finalizeInvitationAcceptance performs the part of acceptance that is
// identical regardless of how the user was resolved (registration-style via
// Accept, or an already-authenticated account via AcceptAuthenticated):
// attach/restore the SchoolUser, assign the invited role, enroll into the
// class if applicable, and mark the invitation accepted.
func finalizeInvitationAcceptance(tx *gorm.DB, invitation domain.Invitation, school domain.School, user *domain.User, now time.Time) (*InvitationAcceptResult, error) {
	schoolUser, err := resolveInvitationSchoolUser(tx, user.ID, invitation.SchoolID)
	if err != nil {
		return nil, err
	}

	var role domain.Role
	if err := tx.Where("rol_name = ?", invitation.Role).First(&role).Error; err != nil {
		return nil, err
	}

	userRole := domain.UserRole{
		SchoolUserID: schoolUser.ID,
		RoleID:       role.ID,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "urol_scu_id"}, {Name: "urol_rol_id"}},
		DoNothing: true,
	}).Create(&userRole).Error; err != nil {
		return nil, err
	}

	if invitation.Role == "student" && invitation.ClassID != nil {
		if err := ensureInvitationStudentEnrollment(tx, invitation, schoolUser.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Model(&domain.Invitation{}).
		Where("inv_id = ? AND inv_accepted_at IS NULL AND inv_revoked_at IS NULL", invitation.ID).
		Updates(map[string]interface{}{
			"inv_accepted_at":    now,
			"inv_target_user_id": user.ID,
			"updated_at":         now,
		}).Error; err != nil {
		return nil, err
	}

	invitation.AcceptedAt = &now
	invitation.TargetUserID = &user.ID
	return &InvitationAcceptResult{
		Invitation: invitation,
		User:       *user,
		School:     school,
		Role:       invitation.Role,
	}, nil
}

func isInvitationUsable(invitation domain.Invitation, now time.Time) bool {
	return invitation.AcceptedAt == nil &&
		invitation.RevokedAt == nil &&
		now.Before(invitation.ExpiresAt)
}

func resolveInvitationUser(tx *gorm.DB, email string, name string, passwordHash string) (*domain.User, error) {
	var user domain.User
	err := tx.Where("usr_email = ?", email).First(&user).Error
	if err == nil {
		updates := map[string]interface{}{}
		if user.Password == "" {
			updates["usr_password"] = passwordHash
			user.Password = passwordHash
		}
		if user.FullName == "" {
			updates["usr_nama_lengkap"] = name
			user.FullName = name
		}
		if len(updates) > 0 {
			if err := tx.Model(&domain.User{}).Where("usr_id = ?", user.ID).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user = domain.User{
		FullName: name,
		Email:    email,
		Password: passwordHash,
		IsActive: true,
	}
	if err := tx.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func resolveInvitationSchoolUser(tx *gorm.DB, userID string, schoolID string) (*domain.SchoolUser, error) {
	var schoolUser domain.SchoolUser
	err := tx.Unscoped().
		Where("scu_usr_id = ? AND scu_sch_id = ?", userID, schoolID).
		First(&schoolUser).Error
	if err == nil {
		if schoolUser.DeletedAt.Valid {
			if err := tx.Unscoped().Model(&domain.SchoolUser{}).
				Where("scu_id = ?", schoolUser.ID).
				Update("deleted_at", nil).Error; err != nil {
				return nil, err
			}
			schoolUser.DeletedAt.Valid = false
		}
		return &schoolUser, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	schoolUser = domain.SchoolUser{
		UserID:   userID,
		SchoolID: schoolID,
	}
	if err := tx.Create(&schoolUser).Error; err != nil {
		return nil, err
	}
	return &schoolUser, nil
}

func ensureInvitationStudentEnrollment(tx *gorm.DB, invitation domain.Invitation, schoolUserID string) error {
	var class domain.Class
	err := tx.Where("cls_id = ? AND cls_sch_id = ? AND deleted_at IS NULL", *invitation.ClassID, invitation.SchoolID).
		First(&class).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvitationClassUnavailable
		}
		return err
	}

	var enrollment domain.Enrollment
	err = tx.Where("enr_scu_id = ? AND enr_cls_id = ?", schoolUserID, class.ID).First(&enrollment).Error
	if err == nil {
		if enrollment.LeftAt == nil && enrollment.Role == "student" {
			return nil
		}
		return tx.Model(&domain.Enrollment{}).
			Where("enr_id = ?", enrollment.ID).
			Updates(map[string]any{
				"enr_role": "student",
				"left_at":  nil,
			}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	enrollment = domain.Enrollment{
		SchoolID:     invitation.SchoolID,
		SchoolUserID: schoolUserID,
		ClassID:      class.ID,
		Role:         "student",
	}
	return tx.Create(&enrollment).Error
}
