package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email string, password string) (*dto.LoginResponseDTO, error)
	Register(fullName string, email string, password string) (*dto.LoginResponseDTO, error)
}

type authService struct {
	userRepo       repository.UserRepository
	schoolUserRepo repository.SchoolUserRepository
}

func NewAuthService(userRepo repository.UserRepository, schoolUserRepo repository.SchoolUserRepository) AuthService {
	return &authService{userRepo: userRepo, schoolUserRepo: schoolUserRepo}
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

	return s.Login(email, password) // Auto-login after registration
}

func (s *authService) buildLoginResponse(token string, user *domain.User) (*dto.LoginResponseDTO, error) {
	response := &dto.LoginResponseDTO{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
		},
		Memberships: []dto.MembershipInfo{},
		GlobalRoles: []string{},
	}

	if s.schoolUserRepo == nil {
		return response, nil
	}

	schoolUsers, err := s.schoolUserRepo.GetByUser(user.ID)
	if err != nil {
		return nil, err
	}

	globalRoleSet := map[string]bool{}
	for i, schoolUser := range schoolUsers {
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
			IsDefault: i == 0,
		}
		response.Memberships = append(response.Memberships, membership)

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
