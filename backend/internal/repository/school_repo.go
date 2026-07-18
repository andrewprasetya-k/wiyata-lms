package repository

import (
	"backend/internal/domain"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SchoolRepository interface {
	CreateSchool(school *domain.School) error
	CreateSchoolWithAdmin(school *domain.School, adminUserID string) (*domain.SchoolUser, error)
	GetSchools(search string, status string, page int, limit int, sortBy string, order string) ([]*domain.School, int64, error)
	GetSchoolByCode(schoolCode string) (*domain.School, error)
	GetSchoolByID(schoolID string) (*domain.School, error)
	GetSchoolByName(name string) (*domain.School, error)
	RestoreDeletedSchool(schoolID string) error
	UpdateSchool(school *domain.School) error
	DeleteSchool(schoolID string) error
	HardDeleteSchool(schoolID string) error
	CheckEmailExists(email string, excludeID string) (bool, error)
	CheckPhoneExists(phone string, excludeID string) (bool, error)
	GetSchoolSummary() (active int64, deleted int64, total int64, err error)
	EnrollUser(schoolUser *domain.SchoolUser) error
	// GenerateUniqueCode produces a random, currently-unused school code.
	// Shared by every caller that needs to auto-generate a code (self-service
	// creation, super-admin bootstrap) instead of each maintaining its own copy.
	GenerateUniqueCode() (string, error)
	// WithTx returns a repository instance bound to an existing transaction, so
	// callers can compose multiple repository operations into one atomic unit.
	WithTx(tx *gorm.DB) SchoolRepository
}

type schoolRepository struct {
	db *gorm.DB
}

// constructor
func NewSchoolRepository(db *gorm.DB) SchoolRepository {
	return &schoolRepository{db: db}
}

func (r *schoolRepository) WithTx(tx *gorm.DB) SchoolRepository {
	return &schoolRepository{db: tx}
}

func (r *schoolRepository) CreateSchool(school *domain.School) error {
	return r.db.Create(school).Error
}

func (r *schoolRepository) GenerateUniqueCode() (string, error) {
	alphabet := []rune("ABCDEFGHJKMNPQRSTUVWXYZ23456789")
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	for range 10 { // Coba maksimal 10 kali
		code := make([]rune, 6)
		for i := range code {
			code[i] = alphabet[seededRand.Intn(len(alphabet))]
		}

		var count int64
		if err := r.db.Unscoped().Model(&domain.School{}).Where("sch_code = ?", string(code)).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return string(code), nil
		}
	}

	return "", fmt.Errorf("failed to generate a unique school code")
}

func (r *schoolRepository) CreateSchoolWithAdmin(school *domain.School, adminUserID string) (*domain.SchoolUser, error) {
	var schoolUser *domain.SchoolUser

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(school).Error; err != nil {
			return err
		}

		su := domain.SchoolUser{
			UserID:   adminUserID,
			SchoolID: school.ID,
		}
		if err := tx.Create(&su).Error; err != nil {
			return err
		}

		var adminRole domain.Role
		if err := tx.Where("rol_name = ?", "admin").First(&adminRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("admin role not found")
			}
			return err
		}

		userRole := domain.UserRole{
			SchoolUserID: su.ID,
			RoleID:       adminRole.ID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}

		su.School = *school
		schoolUser = &su
		return nil
	})
	if err != nil {
		return nil, err
	}

	return schoolUser, nil
}

func (r *schoolRepository) GetSchools(search string, status string, page int, limit int, sortBy string, order string) ([]*domain.School, int64, error) {
	var schools []*domain.School
	var total int64

	query := r.db.Model(&domain.School{})

	// Filter by status
	switch status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	default:
		// "all" or empty -> Include soft-deleted records
		query = query.Unscoped()
	}

	// Filter by search term (Name or Code)
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("sch_name ILIKE ? OR sch_code ILIKE ?", searchTerm, searchTerm)
	}
	//hitung total data
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting Whitelist & Mapping
	sortMap := map[string]string{
		"name":      "sch_name",
		"code":      "sch_code",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}

	column, ok := sortMap[sortBy]
	if !ok {
		column = "created_at" // Default sort column
	}

	// Validate order
	if strings.ToLower(order) != "asc" {
		order = "desc"
	}

	query = query.Order(fmt.Sprintf("%s %s", column, order))

	//pagiatanion
	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Find(&schools).Error
	return schools, total, err
}

func (r *schoolRepository) GetSchoolByCode(schoolCode string) (*domain.School, error) {
	var school domain.School
	err := r.db.Unscoped().Where("sch_code = ?", schoolCode).First(&school).Error
	return &school, err
}

func (r *schoolRepository) GetSchoolByID(schoolID string) (*domain.School, error) {
	var school domain.School
	err := r.db.Unscoped().Where("sch_id = ?", schoolID).First(&school).Error
	return &school, err
}

func (r *schoolRepository) UpdateSchool(school *domain.School) error {
	result := r.db.Updates(school)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolRepository) RestoreDeletedSchool(schoolID string) error {
	result := r.db.Unscoped().Model(&domain.School{}).Where("sch_id = ?", schoolID).Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolRepository) DeleteSchool(schoolID string) error {
	result := r.db.Delete(&domain.School{}, "sch_id = ?", schoolID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolRepository) HardDeleteSchool(schoolID string) error {
	result := r.db.Unscoped().Delete(&domain.School{}, "sch_id = ?", schoolID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schoolRepository) CheckEmailExists(email string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.School{}).Where("sch_email = ?", email)
	if excludeID != "" {
		query = query.Where("sch_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *schoolRepository) CheckPhoneExists(phone string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.School{}).Where("sch_phone = ?", phone)
	if excludeID != "" {
		query = query.Where("sch_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *schoolRepository) GetSchoolSummary() (active int64, deleted int64, total int64, err error) {
	// Single scan over the table using FILTER instead of two separate Count() round-trips.
	row := r.db.Model(&domain.School{}).Unscoped().
		Select("COUNT(*) FILTER (WHERE deleted_at IS NULL) AS active, COUNT(*) FILTER (WHERE deleted_at IS NOT NULL) AS deleted").
		Row()
	if err = row.Scan(&active, &deleted); err != nil {
		return
	}
	total = active + deleted
	return
}

func (r *schoolRepository) GetSchoolByName(name string) (*domain.School, error) {
	var school domain.School
	err := r.db.Where("LOWER(sch_name) = LOWER(?)", name).First(&school).Error
	return &school, err
}

func (r *schoolRepository) EnrollUser(schoolUser *domain.SchoolUser) error {
	return r.db.Create(schoolUser).Error
}
