package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *domain.User) error
	FindAll(search string, page int, limit int) ([]*domain.User, int64, error)
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
	ChangePassword(actor domain.ActorContext, id string, oldPassword string, newPassword string) error
}

type userService struct {
	repo       repository.UserRepository
	logService LogService
}

func NewUserService(repo repository.UserRepository, logService LogService) UserService {
	return &userService{repo: repo, logService: logService}
}

func (s *userService) Create(user *domain.User) error {
	user.FullName = strings.TrimSpace(user.FullName)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	// 1. Validasi Email Unik
	exists, err := s.repo.CheckEmailExists(user.Email, "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email '%s' sudah terdaftar", user.Email)
	}

	// 2. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.Create(user)
}

func (s *userService) FindAll(search string, page int, limit int) ([]*domain.User, int64, error) {
	return s.repo.FindAll(search, page, limit)
}

func (s *userService) GetByID(id string) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetByEmail(email string) (*domain.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *userService) Update(user *domain.User) error {
	user.FullName = strings.TrimSpace(user.FullName)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	// 1. Validasi Email Unik (jika diubah)
	exists, err := s.repo.CheckEmailExists(user.Email, user.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email '%s' sudah terdaftar", user.Email)
	}

	return s.repo.Update(user)
}

func (s *userService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *userService) ChangePassword(actor domain.ActorContext, id string, oldPassword string, newPassword string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Verifikasi Password Lama
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("password lama salah")
	}

	// 2. Hash Password Baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	if err := s.repo.Update(user); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "auth.password.changed", "user", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"user_id": id,
	})
	return nil
}
