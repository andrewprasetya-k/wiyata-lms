package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type RBACRepository interface {
	// Role operations
	CreateRole(role *domain.Role) error
	GetRoleByID(id string) (*domain.Role, error)
	GetRoleByName(name string) (*domain.Role, error)
	GetAllRoles() ([]*domain.Role, error)
	UpdateRole(role *domain.Role) error
	DeleteRole(id string) error
	CheckDuplicateRoleName(name string, excludeID string) (bool, error)

	// User-Role association
	AssignRole(userRole *domain.UserRole) error
	RemoveRoleFromUser(schoolUserID string, roleID string) error
	GetUserRoles(schoolUserID string) ([]*domain.UserRole, error)
	SyncUserRoles(schoolUserID string, roleIDs []string) error

	// RBAC Helpers
	GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error)
	IsUserInSchool(userID, schoolID string) (bool, error)
	GetSchoolUserID(userID, schoolID string) (string, error)
	IsSuperAdmin(userID string) (bool, error)

	// WithTx returns a repository instance bound to an existing transaction, so
	// callers can compose multiple repository operations into one atomic unit.
	WithTx(tx *gorm.DB) RBACRepository
}

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{db: db}
}

func (r *rbacRepository) WithTx(tx *gorm.DB) RBACRepository {
	return &rbacRepository{db: tx}
}

func (r *rbacRepository) CreateRole(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *rbacRepository) GetRoleByID(id string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("rol_id = ?", id).First(&role).Error
	return &role, err
}

func (r *rbacRepository) GetRoleByName(name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("rol_name = ?", name).First(&role).Error
	return &role, err
}

func (r *rbacRepository) GetAllRoles() ([]*domain.Role, error) {
	var roles []*domain.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) UpdateRole(role *domain.Role) error {
	result := r.db.Save(role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *rbacRepository) DeleteRole(id string) error {
	result := r.db.Delete(&domain.Role{}, "rol_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *rbacRepository) CheckDuplicateRoleName(name string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Role{}).Where("rol_name = ?", name)
	if excludeID != "" {
		query = query.Where("rol_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *rbacRepository) AssignRole(userRole *domain.UserRole) error {
	return r.db.Create(userRole).Error
}

func (r *rbacRepository) RemoveRoleFromUser(schoolUserID string, roleID string) error {
	result := r.db.Where("urol_scu_id = ? AND urol_rol_id = ?", schoolUserID, roleID).Delete(&domain.UserRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *rbacRepository) GetUserRoles(schoolUserID string) ([]*domain.UserRole, error) {
	var userRoles []*domain.UserRole
	err := r.db.Preload("Role").Where("urol_scu_id = ?", schoolUserID).Find(&userRoles).Error
	return userRoles, err
}

func (r *rbacRepository) SyncUserRoles(schoolUserID string, roleIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete existing roles
		if err := tx.Where("urol_scu_id = ?", schoolUserID).Delete(&domain.UserRole{}).Error; err != nil {
			return err
		}

		// 2. Add new roles
		if len(roleIDs) > 0 {
			var userRoles []domain.UserRole
			for _, rid := range roleIDs {
				userRoles = append(userRoles, domain.UserRole{
					SchoolUserID: schoolUserID,
					RoleID:       rid,
				})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rbacRepository) GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error) {
	var roleNames []string
	err := r.db.Table("edv.user_roles").
		Select("roles.rol_name").
		Joins("JOIN edv.school_users ON edv.school_users.scu_id = edv.user_roles.urol_scu_id").
		Joins("JOIN edv.roles ON edv.roles.rol_id = edv.user_roles.urol_rol_id").
		Where("edv.school_users.scu_usr_id = ? AND edv.school_users.scu_sch_id = ? AND edv.school_users.deleted_at IS NULL", userID, schoolID).
		Pluck("edv.roles.rol_name", &roleNames).Error
	return roleNames, err
}

// IsUserInSchool checks if user belongs to a school
func (r *rbacRepository) IsUserInSchool(userID, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.school_users").
		Where("scu_usr_id = ? AND scu_sch_id = ? AND deleted_at IS NULL", userID, schoolID).
		Count(&count).Error
	return count > 0, err
}

// GetSchoolUserID returns school_user ID for a user in a school
func (r *rbacRepository) GetSchoolUserID(userID, schoolID string) (string, error) {
	var scuID string
	err := r.db.Table("edv.school_users").
		Select("scu_id").
		Where("scu_usr_id = ? AND scu_sch_id = ? AND deleted_at IS NULL", userID, schoolID).
		Pluck("scu_id", &scuID).Error
	if err != nil {
		return "", err
	}
	if scuID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return scuID, nil
}

// IsSuperAdmin checks if user has super_admin role
func (r *rbacRepository) IsSuperAdmin(userID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.user_roles").
		Joins("JOIN edv.roles ON edv.roles.rol_id = edv.user_roles.urol_rol_id").
		Joins("JOIN edv.school_users ON edv.school_users.scu_id = edv.user_roles.urol_scu_id").
		Where("edv.school_users.scu_usr_id = ? AND edv.school_users.deleted_at IS NULL AND edv.roles.rol_name = ?", userID, "super_admin").
		Count(&count).Error
	return count > 0, err
}
