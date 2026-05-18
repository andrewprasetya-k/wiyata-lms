package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email string, password string) (string, *domain.User, error)                     // returns JWT token and user
	Register(fullName string, email string, password string) (string, *domain.User, error) // returns JWT token and user
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(email string, password string) (string, *domain.User, error) {
	userEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userEmail.Password), []byte(password))
	if err != nil {
		// Return same generic error for password mismatch
		return "", nil, errors.New("invalid email or password")
	}

	payload := jwt.MapClaims{
		"user_id": userEmail.ID,
		"email":   userEmail.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", nil, errors.New("server configuration error")
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", nil, err
	}

	return tokenString, userEmail, nil
}

func (s *authService) Register(fullName string, email string, password string) (string, *domain.User, error) {
	isEmailExists, err := s.userRepo.CheckEmailExists(email, "")
	if err != nil {
		return "", nil, err
	}
	if isEmailExists {
		return "", nil, errors.New("Email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	user := &domain.User{
		FullName: fullName,
		Email:    email,
		Password: string(hashedPassword),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return "", nil, err
	}

	return s.Login(email, password) // Auto-login after registration
}
