package repository

import (
	"backend/internal/domain"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SchoolRegistrationRequestRepository interface {
	Create(request *domain.SchoolRegistrationRequest) error
	HasPendingDuplicate(schoolName string, requesterUserID string) (bool, error)
	List(status string, page int, limit int) ([]*domain.SchoolRegistrationRequest, int64, error)
	GetByID(id string) (*domain.SchoolRegistrationRequest, error)
	RejectPending(id string, reviewerID string, reviewedAt time.Time, reviewNote *string) error
	ApprovePending(id string, school *domain.School, requesterUserID string, adminRoleName string, reviewerID string, reviewedAt time.Time, reviewNote *string) (*SchoolRegistrationApprovalResult, error)
}

type SchoolRegistrationApprovalResult struct {
	Request    *domain.SchoolRegistrationRequest
	School     *domain.School
	SchoolUser *domain.SchoolUser
}

type schoolRegistrationRequestRepository struct {
	db *gorm.DB
}

func NewSchoolRegistrationRequestRepository(db *gorm.DB) SchoolRegistrationRequestRepository {
	return &schoolRegistrationRequestRepository{db: db}
}

func (r *schoolRegistrationRequestRepository) Create(request *domain.SchoolRegistrationRequest) error {
	return r.db.Create(request).Error
}

func (r *schoolRegistrationRequestRepository) HasPendingDuplicate(schoolName string, requesterUserID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SchoolRegistrationRequest{}).
		Where("srr_status = ?", domain.SchoolRegistrationPending).
		Where(
			"LOWER(TRIM(srr_school_name)) = ? OR srr_usr_id = ?",
			strings.ToLower(strings.TrimSpace(schoolName)),
			requesterUserID,
		).
		Count(&count).Error
	return count > 0, err
}

func (r *schoolRegistrationRequestRepository) List(status string, page int, limit int) ([]*domain.SchoolRegistrationRequest, int64, error) {
	var requests []*domain.SchoolRegistrationRequest
	var total int64

	query := r.db.Model(&domain.SchoolRegistrationRequest{})
	if status != "" {
		query = query.Where("srr_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&requests).Error
	return requests, total, err
}

func (r *schoolRegistrationRequestRepository) GetByID(id string) (*domain.SchoolRegistrationRequest, error) {
	var request domain.SchoolRegistrationRequest
	err := r.db.Where("srr_id = ?", id).First(&request).Error
	return &request, err
}

func (r *schoolRegistrationRequestRepository) RejectPending(id string, reviewerID string, reviewedAt time.Time, reviewNote *string) error {
	result := r.db.Model(&domain.SchoolRegistrationRequest{}).
		Where("srr_id = ? AND srr_status = ?", id, domain.SchoolRegistrationPending).
		Updates(map[string]interface{}{
			"srr_status":      domain.SchoolRegistrationRejected,
			"srr_reviewed_by": reviewerID,
			"srr_reviewed_at": reviewedAt,
			"srr_review_note": reviewNote,
			"updated_at":      reviewedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolRegistrationRequestRepository) ApprovePending(id string, school *domain.School, requesterUserID string, adminRoleName string, reviewerID string, reviewedAt time.Time, reviewNote *string) (*SchoolRegistrationApprovalResult, error) {
	var result *SchoolRegistrationApprovalResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var request domain.SchoolRegistrationRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("srr_id = ?", id).
			First(&request).Error; err != nil {
			return err
		}
		if request.Status != domain.SchoolRegistrationPending {
			return errors.New("school registration request is not pending")
		}

		var schoolCount int64
		if err := tx.Unscoped().Model(&domain.School{}).
			Where("sch_code = ?", school.Code).
			Count(&schoolCount).Error; err != nil {
			return err
		}
		if schoolCount > 0 {
			return errors.New("school registration duplicate school code")
		}

		if err := tx.Create(school).Error; err != nil {
			return err
		}

		schoolUser := &domain.SchoolUser{
			UserID:   requesterUserID,
			SchoolID: school.ID,
		}
		if err := tx.Create(schoolUser).Error; err != nil {
			return err
		}

		var adminRole domain.Role
		if err := tx.Where("rol_name = ?", adminRoleName).First(&adminRole).Error; err != nil {
			return err
		}
		userRole := domain.UserRole{
			SchoolUserID: schoolUser.ID,
			RoleID:       adminRole.ID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.SchoolRegistrationRequest{}).
			Where("srr_id = ? AND srr_status = ?", id, domain.SchoolRegistrationPending).
			Updates(map[string]interface{}{
				"srr_status":      domain.SchoolRegistrationApproved,
				"srr_reviewed_by": reviewerID,
				"srr_reviewed_at": reviewedAt,
				"srr_review_note": reviewNote,
				"updated_at":      reviewedAt,
			}).Error; err != nil {
			return err
		}

		var refreshed domain.SchoolRegistrationRequest
		if err := tx.Where("srr_id = ?", id).First(&refreshed).Error; err != nil {
			return err
		}
		result = &SchoolRegistrationApprovalResult{
			Request:    &refreshed,
			School:     school,
			SchoolUser: schoolUser,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
