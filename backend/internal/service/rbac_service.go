package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type RBACService interface {
	// Role management
	CreateRole(role *domain.Role) error
	GetAllRoles() ([]*domain.Role, error)
	GetRoleByID(id string) (*domain.Role, error)
	UpdateRole(role *domain.Role) error
	DeleteRole(id string) error

	// User-Role management
	AssignRoleToUser(schoolUserID string, roleID string) error
	RemoveRoleFromUser(schoolUserID string, roleID string) error
	GetUserRoles(schoolUserID string) ([]*domain.UserRole, error)
	SyncUserRoles(schoolUserID string, roleIDs []string) error

	// Super Admin management
	CreateSuperAdmin(name, email, password string) error
}

type rbacService struct {
	repo        repository.RBACRepository
	userService UserService
	schoolRepo  repository.SchoolRepository
}

func NewRBACService(repo repository.RBACRepository, userService UserService, schoolRepo repository.SchoolRepository) RBACService {
	return &rbacService{
		repo:        repo,
		userService: userService,
		schoolRepo:  schoolRepo,
	}
}

func (s *rbacService) CreateRole(role *domain.Role) error {
	role.Name = strings.TrimSpace(role.Name)

	// 1. Validasi Duplikasi Nama Role Global
	exists, err := s.repo.CheckDuplicateRoleName(role.Name, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("role '%s' sudah terdaftar", role.Name)
	}

	return s.repo.CreateRole(role)
}

func (s *rbacService) GetAllRoles() ([]*domain.Role, error) {
	return s.repo.GetAllRoles()
}

func (s *rbacService) GetRoleByID(id string) (*domain.Role, error) {
	return s.repo.GetRoleByID(id)
}

func (s *rbacService) UpdateRole(role *domain.Role) error {
	role.Name = strings.TrimSpace(role.Name)

	// Validasi Duplikasi Nama
	exists, err := s.repo.CheckDuplicateRoleName(role.Name, role.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("role '%s' sudah terdaftar", role.Name)
	}

	return s.repo.UpdateRole(role)
}

func (s *rbacService) DeleteRole(id string) error {
	return s.repo.DeleteRole(id)
}

func (s *rbacService) AssignRoleToUser(schoolUserID string, roleID string) error {
	userRole := &domain.UserRole{
		SchoolUserID: schoolUserID,
		RoleID:       roleID,
	}
	return s.repo.AssignRole(userRole)
}

func (s *rbacService) RemoveRoleFromUser(schoolUserID string, roleID string) error {
	return s.repo.RemoveRoleFromUser(schoolUserID, roleID)
}

func (s *rbacService) GetUserRoles(schoolUserID string) ([]*domain.UserRole, error) {
	return s.repo.GetUserRoles(schoolUserID)
}

func (s *rbacService) SyncUserRoles(schoolUserID string, roleIDs []string) error {
	return s.repo.SyncUserRoles(schoolUserID, roleIDs)
}

func (s *rbacService) CreateSuperAdmin(name, email, password string) error {
	// 1. Get admin school
	adminSchool, err := s.schoolRepo.GetSchoolByName("admin")
	if err != nil {
		return fmt.Errorf("admin school not found: %v", err)
	}

	// 2. Get super_admin role
	roles, err := s.repo.GetAllRoles()
	if err != nil {
		return fmt.Errorf("failed to get roles: %v", err)
	}

	var superAdminRoleID string
	for _, role := range roles {
		if role.Name == "super_admin" {
			superAdminRoleID = role.ID
			break
		}
	}
	if superAdminRoleID == "" {
		return fmt.Errorf("super_admin role not found")
	}

	// 3. Create user (userService.Create will hash password automatically)
	user := &domain.User{
		FullName: name,
		Email:    email,
		Password: password,
		IsActive: true,
	}
	if err := s.userService.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// 4. Enroll to admin school
	schoolUser := &domain.SchoolUser{
		UserID:   user.ID,
		SchoolID: adminSchool.ID,
	}
	if err := s.schoolRepo.EnrollUser(schoolUser); err != nil {
		return fmt.Errorf("failed to enroll user: %v", err)
	}

	// 5. Assign super_admin role
	userRole := &domain.UserRole{
		SchoolUserID: schoolUser.ID,
		RoleID:       superAdminRoleID,
	}
	if err := s.repo.AssignRole(userRole); err != nil {
		return fmt.Errorf("failed to assign role: %v", err)
	}

	return nil
}
