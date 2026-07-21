package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type SchoolMemberInvitationService interface {
	Create(schoolID string, invitedBy string, input dto.CreateSchoolMemberInvitationDTO) (*dto.CreateSchoolMemberInvitationResponseDTO, error)
	List(schoolID string, status string, page int, limit int) (*dto.SchoolMemberInvitationListResponseDTO, error)
	Revoke(schoolID string, invitationID string) (*dto.SchoolMemberInvitationDTO, error)
}

type schoolMemberInvitationService struct {
	repo         repository.SchoolMemberInvitationRepository
	emailService EmailService
}

func NewSchoolMemberInvitationService(repo repository.SchoolMemberInvitationRepository, emailService EmailService) SchoolMemberInvitationService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &schoolMemberInvitationService{repo: repo, emailService: emailService}
}

func (s *schoolMemberInvitationService) Create(schoolID string, invitedBy string, input dto.CreateSchoolMemberInvitationDTO) (*dto.CreateSchoolMemberInvitationResponseDTO, error) {
	fullName := strings.TrimSpace(input.FullName)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	role := strings.ToLower(strings.TrimSpace(input.Role))
	classCode := strings.TrimSpace(input.ClassCode)

	if schoolID == "" {
		return nil, errors.New("active school context is required")
	}
	if invitedBy == "" {
		return nil, errors.New("inviting user is required")
	}
	// Full name is optional: the invited person supplies their own name when
	// they accept (AcceptInvitationDTO.Name), same as the existing Accept
	// flow already does for a brand-new account. Still bounded when given,
	// to match the domain column's constraint.
	if len(fullName) > 150 {
		return nil, errors.New("invitation full name exceeds 150 characters")
	}
	if email == "" {
		return nil, errors.New("invitation email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("invitation email is invalid")
	}
	if role != "student" && role != "teacher" {
		return nil, errors.New("invitation role must be student or teacher")
	}

	school, err := s.repo.FindSchoolByID(schoolID)
	if err != nil {
		return nil, err
	}

	var classID *string
	var invitedClass *domain.Class
	if role == "student" {
		if classCode == "" {
			return nil, errors.New("classCode is required for student invitation")
		}
		class, err := s.repo.FindClassByCode(schoolID, classCode)
		if err != nil {
			return nil, errors.New("classCode was not found in active school")
		}
		classID = &class.ID
		invitedClass = class
	} else if classCode != "" {
		return nil, errors.New("classCode is only allowed for student invitation")
	}

	now := time.Now()
	duplicate, err := s.repo.HasPendingDuplicate(schoolID, email, role, now)
	if err != nil {
		return nil, err
	}
	if duplicate {
		return nil, errors.New("pending invitation already exists for this email and role")
	}

	rawToken, tokenHash, err := generateSchoolMemberInvitationToken()
	if err != nil {
		return nil, err
	}

	var fullNamePtr *string
	if fullName != "" {
		fullNamePtr = &fullName
	}

	invitation := &domain.Invitation{
		SchoolID:  schoolID,
		Email:     email,
		Role:      role,
		FullName:  fullNamePtr,
		ClassID:   classID,
		TokenHash: tokenHash,
		InvitedBy: invitedBy,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	if err := s.repo.Create(invitation); err != nil {
		return nil, err
	}
	if invitedClass != nil {
		invitation.Class = *invitedClass
	}

	acceptURL := "/invite/" + rawToken
	emailAcceptURL := buildInvitationAcceptURL(rawToken)
	if err := s.emailService.SendSchoolMemberInvitation(invitation.Email, school.Name, invitation.Role, emailAcceptURL); err != nil {
		fmt.Printf("[Email Warning] failed to send school member invitation invitation_id=%s email=%s error=%s\n", invitation.ID, maskEmail(invitation.Email), err.Error())
	}

	return &dto.CreateSchoolMemberInvitationResponseDTO{
		Message:    "School member invitation created",
		Invitation: mapSchoolMemberInvitation(*invitation, now),
		AcceptURL:  acceptURL,
		Token:      rawToken,
	}, nil
}

func (s *schoolMemberInvitationService) List(schoolID string, status string, page int, limit int) (*dto.SchoolMemberInvitationListResponseDTO, error) {
	if schoolID == "" {
		return nil, errors.New("active school context is required")
	}
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		status = "pending"
	}
	if !isValidSchoolMemberInvitationStatus(status) {
		return nil, errors.New("invitation status is invalid")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	invitations, total, err := s.repo.List(schoolID, status, page, limit, now)
	if err != nil {
		return nil, err
	}

	data := make([]dto.SchoolMemberInvitationDTO, 0, len(invitations))
	for _, invitation := range invitations {
		data = append(data, mapSchoolMemberInvitation(invitation, now))
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &dto.SchoolMemberInvitationListResponseDTO{
		Data:       data,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}, nil
}

func (s *schoolMemberInvitationService) Revoke(schoolID string, invitationID string) (*dto.SchoolMemberInvitationDTO, error) {
	if schoolID == "" {
		return nil, errors.New("active school context is required")
	}
	invitationID = strings.TrimSpace(invitationID)
	if invitationID == "" {
		return nil, errors.New("invitation id is required")
	}
	now := time.Now()
	invitation, err := s.repo.Revoke(schoolID, invitationID, now)
	if err != nil {
		return nil, err
	}
	mapped := mapSchoolMemberInvitation(*invitation, now)
	return &mapped, nil
}

func generateSchoolMemberInvitationToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func isValidSchoolMemberInvitationStatus(status string) bool {
	switch status {
	case "pending", "accepted", "revoked", "expired":
		return true
	default:
		return false
	}
}

func mapSchoolMemberInvitation(invitation domain.Invitation, now time.Time) dto.SchoolMemberInvitationDTO {
	var fullName string
	if invitation.FullName != nil {
		fullName = *invitation.FullName
	}

	var classDTO *dto.InvitationClassDTO
	if invitation.ClassID != nil {
		classDTO = &dto.InvitationClassDTO{
			ClassID:    *invitation.ClassID,
			ClassCode:  invitation.Class.Code,
			ClassTitle: invitation.Class.Title,
		}
	}

	var acceptedAt *string
	if invitation.AcceptedAt != nil {
		formatted := formatAPITime(*invitation.AcceptedAt)
		acceptedAt = &formatted
	}
	var revokedAt *string
	if invitation.RevokedAt != nil {
		formatted := formatAPITime(*invitation.RevokedAt)
		revokedAt = &formatted
	}

	return dto.SchoolMemberInvitationDTO{
		InvitationID: invitation.ID,
		FullName:     fullName,
		Email:        invitation.Email,
		Role:         invitation.Role,
		Class:        classDTO,
		Status:       schoolMemberInvitationStatus(invitation, now),
		ExpiresAt:    formatAPITime(invitation.ExpiresAt),
		AcceptedAt:   acceptedAt,
		RevokedAt:    revokedAt,
		CreatedAt:    formatAPITime(invitation.CreatedAt),
	}
}

func schoolMemberInvitationStatus(invitation domain.Invitation, now time.Time) string {
	if invitation.AcceptedAt != nil {
		return "accepted"
	}
	if invitation.RevokedAt != nil {
		return "revoked"
	}
	if !now.Before(invitation.ExpiresAt) {
		return "expired"
	}
	return "pending"
}
