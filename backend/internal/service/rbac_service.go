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
	DeleteRole(actor domain.ActorContext, id string) error

	// User-Role management
	AssignRoleToUser(actor domain.ActorContext, schoolUserID string, roleID string) error
	RemoveRoleFromUser(actor domain.ActorContext, schoolUserID string, roleID string) error
	GetUserRoles(schoolUserID string) ([]*domain.UserRole, error)
	SyncUserRoles(actor domain.ActorContext, schoolUserID string, roleIDs []string) error

	// Super Admin management
	CreateSuperAdmin(actor domain.ActorContext, name, email, password string) error
	IsSuperAdmin(userID string) (bool, error)
}

type rbacService struct {
	repo        repository.RBACRepository
	userService UserService
	schoolRepo  repository.SchoolRepository
	logService  LogService
}

func NewRBACService(repo repository.RBACRepository, userService UserService, schoolRepo repository.SchoolRepository, logService LogService) RBACService {
	return &rbacService{
		repo:        repo,
		userService: userService,
		schoolRepo:  schoolRepo,
		logService:  logService,
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

func (s *rbacService) DeleteRole(actor domain.ActorContext, id string) error {
	// Best-effort name lookup for the audit metadata only — deletion itself
	// must behave exactly as before even if this lookup fails.
	role, roleErr := s.repo.GetRoleByID(id)

	if err := s.repo.DeleteRole(id); err != nil {
		return err
	}

	roleName := id
	if roleErr == nil && role != nil {
		roleName = role.Name
	}
	_ = s.logService.Log(actor, "rbac.role.deleted", "role", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"role_name": roleName,
	})
	return nil
}

func (s *rbacService) AssignRoleToUser(actor domain.ActorContext, schoolUserID string, roleID string) error {
	existingRoles, err := s.repo.GetUserRoles(schoolUserID)
	if err != nil {
		return err
	}
	newRole, err := s.repo.GetRoleByID(roleID)
	if err != nil {
		return err
	}

	roleNames := make([]string, 0, len(existingRoles)+1)
	for _, ur := range existingRoles {
		roleNames = append(roleNames, ur.Role.Name)
	}
	roleNames = append(roleNames, newRole.Name)
	if err := domain.ValidateSchoolRoleCombination(roleNames); err != nil {
		return err
	}

	userRole := &domain.UserRole{
		SchoolUserID: schoolUserID,
		RoleID:       roleID,
	}
	if err := s.repo.AssignRole(userRole); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "member.role.assigned", "school_user", strPtr(schoolUserID), domain.LogSeverityHigh, map[string]any{
		"role_name": newRole.Name,
	})
	return nil
}

func (s *rbacService) RemoveRoleFromUser(actor domain.ActorContext, schoolUserID string, roleID string) error {
	// Best-effort name lookup for the audit metadata only — removal itself
	// must behave exactly as before even if this lookup fails.
	role, roleErr := s.repo.GetRoleByID(roleID)

	if err := s.repo.RemoveRoleFromUser(schoolUserID, roleID); err != nil {
		return err
	}

	roleName := roleID
	if roleErr == nil && role != nil {
		roleName = role.Name
	}
	_ = s.logService.Log(actor, "member.role.removed", "school_user", strPtr(schoolUserID), domain.LogSeverityHigh, map[string]any{
		"role_name": roleName,
	})
	return nil
}

func (s *rbacService) GetUserRoles(schoolUserID string) ([]*domain.UserRole, error) {
	return s.repo.GetUserRoles(schoolUserID)
}

func (s *rbacService) SyncUserRoles(actor domain.ActorContext, schoolUserID string, roleIDs []string) error {
	beforeRoles, err := s.repo.GetUserRoles(schoolUserID)
	if err != nil {
		return err
	}
	beforeNames := make([]string, 0, len(beforeRoles))
	for _, ur := range beforeRoles {
		beforeNames = append(beforeNames, ur.Role.Name)
	}

	afterNames, err := s.roleNamesByIDs(roleIDs)
	if err != nil {
		return err
	}
	if err := domain.ValidateSchoolRoleCombination(afterNames); err != nil {
		return err
	}
	if err := s.repo.SyncUserRoles(schoolUserID, roleIDs); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "member.role.synced", "school_user", strPtr(schoolUserID), domain.LogSeverityHigh, map[string]any{
		"before_roles": beforeNames,
		"after_roles":  afterNames,
	})
	return nil
}

func (s *rbacService) roleNamesByIDs(roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	roles, err := s.repo.GetAllRoles()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]string, len(roles))
	for _, role := range roles {
		byID[role.ID] = role.Name
	}
	names := make([]string, 0, len(roleIDs))
	for _, id := range roleIDs {
		if name, ok := byID[id]; ok {
			names = append(names, name)
		}
	}
	return names, nil
}

func (s *rbacService) IsSuperAdmin(userID string) (bool, error) {
	return s.repo.IsSuperAdmin(userID)
}

func (s *rbacService) CreateSuperAdmin(actor domain.ActorContext, name, email, password string) error {
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

	_ = s.logService.Log(actor, "platform.super_admin.created", "user", strPtr(user.ID), domain.LogSeverityHigh, map[string]any{
		"email": email,
	})
	return nil
}
