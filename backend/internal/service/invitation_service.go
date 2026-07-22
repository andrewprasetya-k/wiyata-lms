package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type InvitationService interface {
	GetMetadata(token string) (*dto.InvitationMetadataDTO, error)
	Accept(token string, input dto.AcceptInvitationDTO) (*dto.AcceptInvitationResponseDTO, error)
	AcceptAuthenticated(token string, userID string) (*dto.AcceptInvitationResponseDTO, error)
}

type invitationService struct {
	repo       repository.InvitationRepository
	userRepo   repository.UserRepository
	logService LogService
}

func NewInvitationService(repo repository.InvitationRepository, userRepo repository.UserRepository, logService LogService) InvitationService {
	return &invitationService{repo: repo, userRepo: userRepo, logService: logService}
}

// invitationAcceptedActor builds the ActorContext for member.invitation.accepted.
// Unlike AcceptAuthenticated (where the actor is known from the JWT before the
// call even starts), Accept's actor only exists once repo.Accept has resolved
// or created the user — so this is built from the result, not from a request
// parameter, for both paths alike.
func invitationAcceptedActor(result *repository.InvitationAcceptResult) domain.ActorContext {
	return domain.ActorContext{
		UserID:   result.User.ID,
		SchoolID: &result.School.ID,
		Scope:    domain.LogScopeSchool,
	}
}

func (s *invitationService) GetMetadata(token string) (*dto.InvitationMetadataDTO, error) {
	tokenHash, err := hashInvitationToken(token)
	if err != nil {
		return nil, err
	}

	invitation, err := s.repo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, normalizeInvitationError(err)
	}

	existingUser, err := s.userRepo.CheckEmailExists(invitation.Email, "")
	if err != nil {
		return nil, err
	}

	return &dto.InvitationMetadataDTO{
		InvitationID: invitation.ID,
		Email:        invitation.Email,
		Role:         invitation.Role,
		School: dto.InvitationSchoolDTO{
			SchoolID:   invitation.School.ID,
			SchoolCode: invitation.School.Code,
			SchoolName: invitation.School.Name,
		},
		ExpiresAt:    formatAPITime(invitation.ExpiresAt),
		Status:       "valid",
		ExistingUser: existingUser,
	}, nil
}

func (s *invitationService) Accept(token string, input dto.AcceptInvitationDTO) (*dto.AcceptInvitationResponseDTO, error) {
	tokenHash, err := hashInvitationToken(token)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	password := input.Password
	confirmPassword := input.ConfirmPassword
	if name == "" {
		return nil, errors.New("invitation name is required")
	}
	if len(name) > 150 {
		return nil, errors.New("invitation name exceeds 150 characters")
	}
	if len(password) < 6 {
		return nil, errors.New("invitation password must be at least 6 characters")
	}
	if password != confirmPassword {
		return nil, errors.New("invitation password confirmation does not match")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.Accept(tokenHash, name, string(passwordHash), time.Now())
	if err != nil {
		return nil, normalizeInvitationError(err)
	}

	_ = s.logService.Log(invitationAcceptedActor(result), "member.invitation.accepted", "invitation", strPtr(result.Invitation.ID), domain.LogSeverityLow, map[string]any{
		"email": result.Invitation.Email,
		"role":  result.Role,
	})

	return &dto.AcceptInvitationResponseDTO{
		Message: "Invitation accepted",
		User: dto.InvitationAcceptedUserDTO{
			UserID:   result.User.ID,
			FullName: result.User.FullName,
			Email:    result.User.Email,
		},
		School: dto.InvitationSchoolDTO{
			SchoolID:   result.School.ID,
			SchoolCode: result.School.Code,
			SchoolName: result.School.Name,
		},
		Role: result.Role,
	}, nil
}

func (s *invitationService) AcceptAuthenticated(token string, userID string) (*dto.AcceptInvitationResponseDTO, error) {
	tokenHash, err := hashInvitationToken(token)
	if err != nil {
		return nil, err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("invitation authenticated user is required")
	}

	result, err := s.repo.AcceptAuthenticated(tokenHash, userID, time.Now())
	if err != nil {
		return nil, normalizeInvitationError(err)
	}

	_ = s.logService.Log(invitationAcceptedActor(result), "member.invitation.accepted", "invitation", strPtr(result.Invitation.ID), domain.LogSeverityLow, map[string]any{
		"email": result.Invitation.Email,
		"role":  result.Role,
	})

	return &dto.AcceptInvitationResponseDTO{
		Message: "Invitation accepted",
		User: dto.InvitationAcceptedUserDTO{
			UserID:   result.User.ID,
			FullName: result.User.FullName,
			Email:    result.User.Email,
		},
		School: dto.InvitationSchoolDTO{
			SchoolID:   result.School.ID,
			SchoolCode: result.School.Code,
			SchoolName: result.School.Name,
		},
		Role: result.Role,
	}, nil
}

func hashInvitationToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", repository.ErrInvitationInvalid
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func normalizeInvitationError(err error) error {
	if errors.Is(err, repository.ErrInvitationInvalid) {
		return repository.ErrInvitationInvalid
	}
	return err
}
