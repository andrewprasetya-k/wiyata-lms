package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email string, password string) (*dto.LoginResponseDTO, error)
	Register(fullName string, email string, password string) (*dto.LoginResponseDTO, error)
	GetContext(userID string) (*dto.AuthContextResponseDTO, error)
}

type authService struct {
	userRepo             repository.UserRepository
	schoolUserRepo       repository.SchoolUserRepository
	emailVerificationSvc EmailVerificationService
}

func NewAuthService(userRepo repository.UserRepository, schoolUserRepo repository.SchoolUserRepository, emailVerificationSvc EmailVerificationService) AuthService {
	return &authService{userRepo: userRepo, schoolUserRepo: schoolUserRepo, emailVerificationSvc: emailVerificationSvc}
}

func (s *authService) Login(email string, password string) (*dto.LoginResponseDTO, error) {
	userEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userEmail.Password), []byte(password))
	if err != nil {
		// Return same generic error for password mismatch
		return nil, errors.New("invalid email or password")
	}

	payload := jwt.MapClaims{
		"user_id": userEmail.ID,
		"sub":     userEmail.ID,
		"email":   userEmail.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("server configuration error")
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(tokenString, userEmail)
}

func (s *authService) Register(fullName string, email string, password string) (*dto.LoginResponseDTO, error) {
	isEmailExists, err := s.userRepo.CheckEmailExists(email, "")
	if err != nil {
		return nil, err
	}
	if isEmailExists {
		return nil, errors.New("Email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName: fullName,
		Email:    email,
		Password: string(hashedPassword),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	if s.emailVerificationSvc != nil {
		if err := s.emailVerificationSvc.IssueForNewUser(user); err != nil {
			fmt.Printf("[Email Verification Warning] failed to issue token for user_id=%s error=%s\n", user.ID, err.Error())
		}
	}

	return s.Login(email, password) // Auto-login after registration
}

func (s *authService) GetContext(userID string) (*dto.AuthContextResponseDTO, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return s.buildAuthContext(user)
}

func (s *authService) buildLoginResponse(token string, user *domain.User) (*dto.LoginResponseDTO, error) {
	response := &dto.LoginResponseDTO{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
		},
	}

	context, err := s.buildAuthContext(user)
	if err != nil {
		return nil, err
	}
	response.Memberships = context.Memberships
	response.GlobalRoles = context.GlobalRoles
	response.DefaultContext = context.DefaultContext

	return response, nil
}

func (s *authService) buildAuthContext(user *domain.User) (*dto.AuthContextResponseDTO, error) {
	response := &dto.AuthContextResponseDTO{
		Memberships:     []dto.MembershipInfo{},
		GlobalRoles:     []string{},
		EmailVerified:   user.EmailVerifiedAt != nil,
		EmailVerifiedAt: formatAPITimePtr(user.EmailVerifiedAt),
	}

	if s.schoolUserRepo == nil {
		return response, nil
	}

	schoolUsers, err := s.schoolUserRepo.GetByUser(user.ID)
	if err != nil {
		return nil, err
	}

	globalRoleSet := map[string]bool{}
	activeMembershipIndex := 0
	for _, schoolUser := range schoolUsers {
		if schoolUser.DeletedAt.Valid {
			continue
		}

		roles := make([]string, 0, len(schoolUser.Roles))
		for _, userRole := range schoolUser.Roles {
			if userRole.Role.Name == "" {
				continue
			}
			roles = append(roles, userRole.Role.Name)
			if userRole.Role.Name == "super_admin" && !globalRoleSet[userRole.Role.Name] {
				response.GlobalRoles = append(response.GlobalRoles, userRole.Role.Name)
				globalRoleSet[userRole.Role.Name] = true
			}
		}

		membership := dto.MembershipInfo{
			SchoolUserID: schoolUser.ID,
			School: dto.SchoolInfo{
				ID:   schoolUser.School.ID,
				Code: schoolUser.School.Code,
				Name: schoolUser.School.Name,
			},
			Roles:     roles,
			IsDefault: activeMembershipIndex == 0,
		}
		response.Memberships = append(response.Memberships, membership)
		activeMembershipIndex++

		if response.DefaultContext == nil {
			response.DefaultContext = &dto.DefaultContext{
				SchoolID:     schoolUser.SchoolID,
				SchoolUserID: schoolUser.ID,
				Roles:        roles,
			}
		}
	}

	return response, nil
}
