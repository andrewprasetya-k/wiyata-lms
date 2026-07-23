package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"
	"strings"
)

type UserService interface {
	Create(actor domain.ActorContext, user *domain.User) error
	FindAll(search string, page int, limit int) ([]*domain.User, int64, error)
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(actor domain.ActorContext, user *domain.User) error
	Delete(actor domain.ActorContext, id string) error
	ChangePassword(actor domain.ActorContext, id string, oldPassword string, newPassword string) error
}

type userService struct {
	repo       repository.UserRepository
	logService LogService
}

func NewUserService(repo repository.UserRepository, logService LogService) UserService {
	return &userService{repo: repo, logService: logService}
}

func (s *userService) Create(actor domain.ActorContext, user *domain.User) error {
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
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	if err := s.repo.Create(user); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "user.created", "user", strPtr(user.ID), domain.LogSeverityHigh, map[string]any{
		"email": user.Email,
	})

	return nil
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

func (s *userService) Update(actor domain.ActorContext, user *domain.User) error {
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

	if err := s.repo.Update(user); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "user.updated", "user", strPtr(user.ID), domain.LogSeverityHigh, map[string]any{
		"email": user.Email,
	})

	return nil
}

func (s *userService) Delete(actor domain.ActorContext, id string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "user.deleted", "user", strPtr(id), domain.LogSeverityCritical, map[string]any{
		"email": user.Email,
	})

	return nil
}

func (s *userService) ChangePassword(actor domain.ActorContext, id string, oldPassword string, newPassword string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. Verifikasi Password Lama
	err = verifyPassword(user.Password, oldPassword)
	if err != nil {
		return fmt.Errorf("password lama salah")
	}

	// 2. Hash Password Baru
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	if err := s.repo.Update(user); err != nil {
		return err
	}

	_ = s.logService.Log(actor, "auth.password.changed", "user", strPtr(id), domain.LogSeverityHigh, map[string]any{
		"user_id": id,
	})
	return nil
}
