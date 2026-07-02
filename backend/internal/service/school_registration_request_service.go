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
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SchoolRegistrationRequestService interface {
	Create(input dto.CreateSchoolRegistrationRequestDTO) (*dto.CreateSchoolRegistrationRequestResponseDTO, error)
	List(status string, page int, limit int) (*dto.SchoolRegistrationRequestListResponseDTO, error)
	GetByID(id string) (*dto.SchoolRegistrationRequestDetailDTO, error)
	Reject(id string, reviewerID string, input dto.RejectSchoolRegistrationRequestDTO) (*dto.SchoolRegistrationRequestDetailDTO, error)
	Approve(id string, reviewerID string, input dto.ApproveSchoolRegistrationRequestDTO) (*dto.ApproveSchoolRegistrationRequestResponseDTO, error)
}

type schoolRegistrationRequestService struct {
	repo         repository.SchoolRegistrationRequestRepository
	emailService EmailService
}

func NewSchoolRegistrationRequestService(repo repository.SchoolRegistrationRequestRepository, emailService EmailService) SchoolRegistrationRequestService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &schoolRegistrationRequestService{repo: repo, emailService: emailService}
}

func (s *schoolRegistrationRequestService) Create(input dto.CreateSchoolRegistrationRequestDTO) (*dto.CreateSchoolRegistrationRequestResponseDTO, error) {
	request := domain.SchoolRegistrationRequest{
		SchoolName: strings.TrimSpace(input.SchoolName),
		NPSN:       cleanOptional(input.NPSN),
		PICName:    strings.TrimSpace(input.PICName),
		PICEmail:   strings.ToLower(strings.TrimSpace(input.PICEmail)),
		PICPhone:   cleanOptional(input.PICPhone),
		PICRole:    cleanOptional(input.PICRole),
		Message:    cleanOptional(input.Message),
		Status:     domain.SchoolRegistrationPending,
	}

	if err := validateSchoolRegistrationRequest(&request); err != nil {
		return nil, err
	}

	duplicate, err := s.repo.HasPendingDuplicate(request.SchoolName, request.PICEmail)
	if err != nil {
		return nil, err
	}
	if duplicate {
		return nil, errors.New("school registration request pending duplicate")
	}

	if err := s.repo.Create(&request); err != nil {
		return nil, err
	}

	return &dto.CreateSchoolRegistrationRequestResponseDTO{
		Message: "School registration request submitted",
		Request: dto.SchoolRegistrationRequestSummaryDTO{
			RequestID:  request.ID,
			SchoolName: request.SchoolName,
			PICName:    request.PICName,
			PICEmail:   request.PICEmail,
			Status:     string(request.Status),
			CreatedAt:  formatAPITime(request.CreatedAt),
		},
	}, nil
}

func (s *schoolRegistrationRequestService) List(status string, page int, limit int) (*dto.SchoolRegistrationRequestListResponseDTO, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		status = string(domain.SchoolRegistrationPending)
	}
	if !isValidSchoolRegistrationStatus(status) {
		return nil, errors.New("school registration status is invalid")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	requests, total, err := s.repo.List(status, page, limit)
	if err != nil {
		return nil, err
	}

	data := make([]dto.SchoolRegistrationRequestDetailDTO, 0, len(requests))
	for _, request := range requests {
		data = append(data, mapSchoolRegistrationRequestDetail(request))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &dto.SchoolRegistrationRequestListResponseDTO{
		Data:       data,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}, nil
}

func (s *schoolRegistrationRequestService) GetByID(id string) (*dto.SchoolRegistrationRequestDetailDTO, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("school registration request id is required")
	}

	request, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := mapSchoolRegistrationRequestDetail(request)
	return &response, nil
}

func (s *schoolRegistrationRequestService) Reject(id string, reviewerID string, input dto.RejectSchoolRegistrationRequestDTO) (*dto.SchoolRegistrationRequestDetailDTO, error) {
	id = strings.TrimSpace(id)
	reviewerID = strings.TrimSpace(reviewerID)
	if id == "" {
		return nil, errors.New("school registration request id is required")
	}
	if reviewerID == "" {
		return nil, errors.New("school registration reviewer is required")
	}

	reason := cleanOptional(input.Reason)
	if reason != nil && len(*reason) > 1000 {
		return nil, errors.New("school registration rejection reason exceeds 1000 characters")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing.Status != domain.SchoolRegistrationPending {
		return nil, errors.New("school registration request is not pending")
	}

	if err := s.repo.RejectPending(id, reviewerID, time.Now(), reason); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("school registration request is not pending")
		}
		return nil, err
	}

	updated, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	response := mapSchoolRegistrationRequestDetail(updated)
	return &response, nil
}

func (s *schoolRegistrationRequestService) Approve(id string, reviewerID string, input dto.ApproveSchoolRegistrationRequestDTO) (*dto.ApproveSchoolRegistrationRequestResponseDTO, error) {
	id = strings.TrimSpace(id)
	reviewerID = strings.TrimSpace(reviewerID)
	if id == "" {
		return nil, errors.New("school registration request id is required")
	}
	if reviewerID == "" {
		return nil, errors.New("school registration reviewer is required")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing.Status != domain.SchoolRegistrationPending {
		return nil, errors.New("school registration request is not pending")
	}

	schoolCode := strings.TrimSpace(input.SchoolCode)
	schoolName := valueOrDefault(input.SchoolName, existing.SchoolName)
	adminName := valueOrDefault(input.AdminName, existing.PICName)
	adminEmail := strings.ToLower(valueOrDefault(input.AdminEmail, existing.PICEmail))
	note := cleanOptional(input.Note)

	if schoolCode == "" {
		return nil, errors.New("school registration school code is required")
	}
	if len(schoolCode) > 50 {
		return nil, errors.New("school registration school code exceeds 50 characters")
	}
	if schoolName == "" {
		return nil, errors.New("school registration school name is required")
	}
	if len(schoolName) > 150 {
		return nil, errors.New("school registration school name exceeds 150 characters")
	}
	if adminName == "" {
		return nil, errors.New("school registration admin name is required")
	}
	if len(adminName) > 150 {
		return nil, errors.New("school registration admin name exceeds 150 characters")
	}
	if adminEmail == "" {
		return nil, errors.New("school registration admin email is required")
	}
	if _, err := mail.ParseAddress(adminEmail); err != nil {
		return nil, errors.New("school registration admin email is invalid")
	}
	if note != nil && len(*note) > 1000 {
		return nil, errors.New("school registration approval note exceeds 1000 characters")
	}

	rawToken, tokenHash, err := generateInvitationToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	schoolPhone := ""
	if existing.PICPhone != nil {
		schoolPhone = *existing.PICPhone
	}

	school := &domain.School{
		Name:  schoolName,
		Code:  schoolCode,
		Email: adminEmail,
		Phone: schoolPhone,
	}
	invitation := &domain.Invitation{
		Email:     adminEmail,
		Role:      "admin",
		TokenHash: tokenHash,
		InvitedBy: reviewerID,
		ExpiresAt: expiresAt,
	}

	updatedRequest, err := s.repo.ApprovePending(id, school, invitation, reviewerID, now, note)
	if err != nil {
		return nil, err
	}

	acceptURL := "/invite/" + rawToken
	emailAcceptURL := buildInvitationAcceptURL(rawToken)
	if err := s.emailService.SendSchoolAdminInvitation(invitation.Email, school.Name, emailAcceptURL); err != nil {
		fmt.Printf("[Email Warning] failed to send school admin invitation invitation_id=%s email=%s error=%s\n", invitation.ID, maskEmail(invitation.Email), err.Error())
	}

	return &dto.ApproveSchoolRegistrationRequestResponseDTO{
		Message: "School registration request approved",
		Request: mapSchoolRegistrationRequestDetail(updatedRequest),
		School: dto.ApprovedSchoolDTO{
			SchoolID:   school.ID,
			SchoolCode: school.Code,
			SchoolName: school.Name,
		},
		Invitation: dto.AdminInvitationDTO{
			InvitationID: invitation.ID,
			Email:        invitation.Email,
			Role:         invitation.Role,
			ExpiresAt:    formatAPITime(invitation.ExpiresAt),
			AcceptURL:    acceptURL,
			Token:        rawToken,
		},
	}, nil
}

func validateSchoolRegistrationRequest(request *domain.SchoolRegistrationRequest) error {
	if request.SchoolName == "" {
		return errors.New("school registration school name is required")
	}
	if len(request.SchoolName) > 150 {
		return errors.New("school registration school name exceeds 150 characters")
	}
	if request.PICName == "" {
		return errors.New("school registration pic name is required")
	}
	if len(request.PICName) > 150 {
		return errors.New("school registration pic name exceeds 150 characters")
	}
	if request.PICEmail == "" {
		return errors.New("school registration pic email is required")
	}
	if _, err := mail.ParseAddress(request.PICEmail); err != nil {
		return errors.New("school registration pic email is invalid")
	}
	if request.NPSN != nil && len(*request.NPSN) > 50 {
		return errors.New("school registration npsn exceeds 50 characters")
	}
	if request.PICPhone != nil && len(*request.PICPhone) > 50 {
		return errors.New("school registration pic phone exceeds 50 characters")
	}
	if request.PICRole != nil && len(*request.PICRole) > 100 {
		return errors.New("school registration pic role exceeds 100 characters")
	}
	if request.Message != nil && len(*request.Message) > 1000 {
		return errors.New("school registration message exceeds 1000 characters")
	}
	return nil
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func valueOrDefault(value *string, fallback string) string {
	if value == nil {
		return strings.TrimSpace(fallback)
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return strings.TrimSpace(fallback)
	}
	return trimmed
}

func generateInvitationToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func buildInvitationAcceptURL(rawToken string) string {
	path := "/invite/" + rawToken
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if publicURL == "" {
		return path
	}
	return publicURL + path
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return "unknown"
	}
	local := parts[0]
	if len(local) <= 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}

func isValidSchoolRegistrationStatus(status string) bool {
	switch domain.SchoolRegistrationRequestStatus(status) {
	case domain.SchoolRegistrationPending, domain.SchoolRegistrationApproved, domain.SchoolRegistrationRejected:
		return true
	default:
		return false
	}
}

func mapSchoolRegistrationRequestDetail(request *domain.SchoolRegistrationRequest) dto.SchoolRegistrationRequestDetailDTO {
	return dto.SchoolRegistrationRequestDetailDTO{
		RequestID:  request.ID,
		SchoolName: request.SchoolName,
		NPSN:       request.NPSN,
		PICName:    request.PICName,
		PICEmail:   request.PICEmail,
		PICPhone:   request.PICPhone,
		PICRole:    request.PICRole,
		Message:    request.Message,
		Status:     string(request.Status),
		ReviewedBy: request.ReviewedBy,
		ReviewedAt: formatAPITimePtr(request.ReviewedAt),
		ReviewNote: request.ReviewNote,
		CreatedAt:  formatAPITime(request.CreatedAt),
		UpdatedAt:  formatAPITime(request.UpdatedAt),
	}
}
