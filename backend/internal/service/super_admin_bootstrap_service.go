package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SuperAdminBootstrapService interface {
	BootstrapSchool(actor domain.ActorContext, input dto.SchoolBootstrapRequestDTO) (*dto.SchoolBootstrapResponseDTO, error)
}

type superAdminBootstrapService struct {
	db             *gorm.DB
	schoolRepo     repository.SchoolRepository
	schoolUserRepo repository.SchoolUserRepository
	rbacRepo       repository.RBACRepository
	logService     LogService
}

func NewSuperAdminBootstrapService(
	db *gorm.DB,
	schoolRepo repository.SchoolRepository,
	schoolUserRepo repository.SchoolUserRepository,
	rbacRepo repository.RBACRepository,
	logService LogService,
) SuperAdminBootstrapService {
	return &superAdminBootstrapService{
		db:             db,
		schoolRepo:     schoolRepo,
		schoolUserRepo: schoolUserRepo,
		rbacRepo:       rbacRepo,
		logService:     logService,
	}
}

func (s *superAdminBootstrapService) BootstrapSchool(actor domain.ActorContext, input dto.SchoolBootstrapRequestDTO) (*dto.SchoolBootstrapResponseDTO, error) {
	var response *dto.SchoolBootstrapResponseDTO

	err := s.db.Transaction(func(tx *gorm.DB) error {
		school, err := s.createBootstrapSchool(tx, input.School)
		if err != nil {
			return err
		}

		adminUser, err := s.resolveBootstrapAdminUser(tx, input.AdminUser)
		if err != nil {
			return err
		}

		schoolUser := domain.SchoolUser{
			UserID:   adminUser.ID,
			SchoolID: school.ID,
		}
		if err := s.schoolUserRepo.WithTx(tx).Create(&schoolUser); err != nil {
			return err
		}

		adminRole, err := s.rbacRepo.WithTx(tx).GetRoleByName("admin")
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("bootstrap admin role not found")
			}
			return err
		}

		userRole := domain.UserRole{
			SchoolUserID: schoolUser.ID,
			RoleID:       adminRole.ID,
		}
		if err := s.rbacRepo.WithTx(tx).AssignRole(&userRole); err != nil {
			return err
		}

		response = &dto.SchoolBootstrapResponseDTO{
			School: dto.SchoolBootstrapSchoolDTO{
				ID:   school.ID,
				Name: school.Name,
				Code: school.Code,
			},
			AdminUser: dto.BootstrapAdminUserResponseDTO{
				ID:       adminUser.ID,
				FullName: adminUser.FullName,
				Email:    adminUser.Email,
				IsActive: adminUser.IsActive,
			},
			SchoolUserID:  schoolUser.ID,
			AssignedRoles: []string{"admin"},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.logService.Log(actor, "platform.school.bootstrapped", "school", strPtr(response.School.ID), domain.LogSeverityHigh, map[string]any{
		"school_code": response.School.Code,
		"admin_email": response.AdminUser.Email,
		"admin_mode":  strings.ToLower(strings.TrimSpace(input.AdminUser.Mode)),
	})

	return response, nil
}

func (s *superAdminBootstrapService) createBootstrapSchool(tx *gorm.DB, input dto.CreateSchoolDTO) (*domain.School, error) {
	school := domain.School{
		Name:    strings.TrimSpace(input.Name),
		Code:    strings.TrimSpace(input.Code),
		LogoID:  input.LogoID,
		Address: strings.TrimSpace(input.Address),
		Email:   strings.TrimSpace(input.Email),
		Phone:   strings.TrimSpace(input.Phone),
		Website: input.Website,
	}
	if school.Website != nil {
		trimmed := strings.TrimSpace(*school.Website)
		school.Website = &trimmed
	}

	repo := s.schoolRepo.WithTx(tx)

	if school.Email != "" {
		exists, err := repo.CheckEmailExists(school.Email, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("bootstrap duplicate school email")
		}
	}
	if school.Phone != "" {
		exists, err := repo.CheckPhoneExists(school.Phone, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("bootstrap duplicate school phone")
		}
	}

	if school.Code == "" {
		code, err := repo.GenerateUniqueCode()
		if err != nil {
			return nil, fmt.Errorf("bootstrap school code generation failed")
		}
		school.Code = code
	} else {
		_, err := repo.GetSchoolByCode(school.Code)
		if err == nil {
			return nil, fmt.Errorf("bootstrap duplicate school code")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if err := repo.CreateSchool(&school); err != nil {
		return nil, err
	}
	return &school, nil
}

func (s *superAdminBootstrapService) resolveBootstrapAdminUser(tx *gorm.DB, input dto.BootstrapAdminUserDTO) (*domain.User, error) {
	mode := strings.TrimSpace(strings.ToLower(input.Mode))
	switch mode {
	case "new":
		return createBootstrapAdminUser(tx, input)
	case "existing":
		return findExistingBootstrapAdminUser(tx, input.UserID)
	default:
		return nil, fmt.Errorf("bootstrap invalid admin user mode")
	}
}

func createBootstrapAdminUser(tx *gorm.DB, input dto.BootstrapAdminUserDTO) (*domain.User, error) {
	fullName := strings.TrimSpace(input.FullName)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password

	if fullName == "" || email == "" || password == "" {
		return nil, fmt.Errorf("bootstrap new admin user fields are required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("bootstrap invalid admin user email")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("bootstrap admin user password too short")
	}

	var count int64
	if err := tx.Model(&domain.User{}).Where("usr_email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("bootstrap duplicate user email")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		FullName: fullName,
		Email:    email,
		Password: string(hashedPassword),
		IsActive: true,
	}
	if err := tx.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func findExistingBootstrapAdminUser(tx *gorm.DB, userID string) (*domain.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("bootstrap existing admin user id is required")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("bootstrap invalid existing admin user id")
	}

	var user domain.User
	if err := tx.Where("usr_id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bootstrap existing admin user not found")
		}
		return nil, err
	}
	return &user, nil
}
