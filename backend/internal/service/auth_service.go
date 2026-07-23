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
)

type PasswordAttemptLimiter interface {
	// Allow records one attempt for key and reports whether it is still
	// within budget. Returns false once the key's budget is exhausted.
	Allow(key string) bool
	// Reset clears any tracked attempts for key, restoring full budget.
	Reset(key string)
}

var (
	ErrInvalidCurrentPassword  = errors.New("current password is incorrect")
	ErrTooManyPasswordAttempts = errors.New("too many failed attempts, try again later")
)

const changePasswordLockKeyPrefix = "change_password:"

type AuthService interface {
	Login(email string, password string) (*dto.LoginResponseDTO, error)
	Register(fullName string, email string, password string) (*dto.LoginResponseDTO, error)
	GetContext(userID string) (*dto.AuthContextResponseDTO, error)
	ChangePassword(userID string, currentPassword string, newPassword string) error
}

type authService struct {
	userRepo             repository.UserRepository
	schoolUserRepo       repository.SchoolUserRepository
	emailVerificationSvc EmailVerificationService
	logService           LogService
	passwordAttemptLimit PasswordAttemptLimiter
}

func NewAuthService(userRepo repository.UserRepository, schoolUserRepo repository.SchoolUserRepository, emailVerificationSvc EmailVerificationService, logService LogService, passwordAttemptLimit PasswordAttemptLimiter) AuthService {
	return &authService{userRepo: userRepo, schoolUserRepo: schoolUserRepo, emailVerificationSvc: emailVerificationSvc, logService: logService, passwordAttemptLimit: passwordAttemptLimit}
}

func (s *authService) logLoginFailed(email string, reason string) {
	_ = s.logService.Log(domain.ActorContext{Scope: domain.LogScopePlatform}, "auth.login.failed", "user", nil, domain.LogSeverityMedium, map[string]any{
		"email":  email,
		"reason": reason,
	})
}

func (s *authService) Login(email string, password string) (*dto.LoginResponseDTO, error) {
	userEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Return generic error to prevent user enumeration
		s.logLoginFailed(email, "user_not_found")
		return nil, errors.New("invalid email or password")
	}

	err = verifyPassword(userEmail.Password, password)
	if err != nil {
		// Return same generic error for password mismatch
		s.logLoginFailed(email, "invalid_password")
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

	response, err := s.buildLoginResponse(tokenString, userEmail)
	if err != nil {
		return nil, err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: userEmail.ID, Scope: domain.LogScopePlatform}, "auth.login.success", "user", strPtr(userEmail.ID), domain.LogSeverityLow, map[string]any{
		"user_id":      userEmail.ID,
		"login_method": "password",
	})

	if response.DefaultContext != nil {
		schoolID := response.DefaultContext.SchoolID
		schoolUserID := response.DefaultContext.SchoolUserID
		_ = s.logService.Log(domain.ActorContext{
			UserID:       userEmail.ID,
			SchoolID:     &schoolID,
			SchoolUserID: &schoolUserID,
			Scope:        domain.LogScopeSchool,
		}, "member.login", "school_user", strPtr(schoolUserID), domain.LogSeverityLow, map[string]any{
			"login_method": "password",
			"user_id":      userEmail.ID,
			"school_id":    schoolID,
		})
	}

	return response, nil
}

func (s *authService) Register(fullName string, email string, password string) (*dto.LoginResponseDTO, error) {
	isEmailExists, err := s.userRepo.CheckEmailExists(email, "")
	if err != nil {
		return nil, err
	}
	if isEmailExists {
		return nil, errors.New("Email already registered")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName: fullName,
		Email:    email,
		Password: hashedPassword,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	_ = s.logService.Log(domain.ActorContext{UserID: user.ID, Scope: domain.LogScopePlatform}, "auth.registered", "user", strPtr(user.ID), domain.LogSeverityMedium, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})

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

// ChangePassword is the self-service counterpart to UserService.ChangePassword
// (which is a super-admin-on-behalf-of-another-user reset). userID always
// comes from the caller's own JWT claims, never a path/body-supplied ID.
func (s *authService) ChangePassword(userID string, currentPassword string, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if err := verifyPassword(user.Password, currentPassword); err != nil {
		lockKey := changePasswordLockKeyPrefix + userID
		reason := "invalid_current_password"
		failErr := ErrInvalidCurrentPassword
		if s.passwordAttemptLimit != nil && !s.passwordAttemptLimit.Allow(lockKey) {
			reason = "rate_limited"
			failErr = ErrTooManyPasswordAttempts
		}
		s.logChangePasswordFailed(userID, reason)
		return failErr
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if s.passwordAttemptLimit != nil {
		s.passwordAttemptLimit.Reset(changePasswordLockKeyPrefix + userID)
	}

	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.password.changed", "user", strPtr(userID), domain.LogSeverityHigh, map[string]any{
		"user_id": userID,
		"method":  "self_service",
	})
	return nil
}

func (s *authService) logChangePasswordFailed(userID string, reason string) {
	_ = s.logService.Log(domain.ActorContext{UserID: userID, Scope: domain.LogScopePlatform}, "auth.password.change.failed", "user", strPtr(userID), domain.LogSeverityMedium, map[string]any{
		"user_id": userID,
		"reason":  reason,
	})
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
