package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SchoolRegistrationRequestService interface {
	Create(input dto.CreateSchoolRegistrationRequestDTO, requesterUserID string) (*dto.CreateSchoolRegistrationRequestResponseDTO, error)
	List(status string, page int, limit int) (*dto.SchoolRegistrationRequestListResponseDTO, error)
	GetByID(id string) (*dto.SchoolRegistrationRequestDetailDTO, error)
	Reject(id string, reviewerID string, input dto.RejectSchoolRegistrationRequestDTO) (*dto.SchoolRegistrationRequestDetailDTO, error)
	Approve(id string, reviewerID string, input dto.ApproveSchoolRegistrationRequestDTO) (*dto.ApproveSchoolRegistrationRequestResponseDTO, error)
}

type schoolRegistrationRequestService struct {
	repo         repository.SchoolRegistrationRequestRepository
	emailService EmailService
	userService  UserService
}

func NewSchoolRegistrationRequestService(repo repository.SchoolRegistrationRequestRepository, emailService EmailService, userService UserService) SchoolRegistrationRequestService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &schoolRegistrationRequestService{repo: repo, emailService: emailService, userService: userService}
}

func (s *schoolRegistrationRequestService) Create(input dto.CreateSchoolRegistrationRequestDTO, requesterUserID string) (*dto.CreateSchoolRegistrationRequestResponseDTO, error) {
	requesterUserID = strings.TrimSpace(requesterUserID)
	if requesterUserID == "" {
		return nil, errors.New("school registration requester is required")
	}

	request := domain.SchoolRegistrationRequest{
		RequesterUserID: requesterUserID,
		SchoolName:      strings.TrimSpace(input.SchoolName),
		NPSN:            cleanOptional(input.NPSN),
		PICName:         strings.TrimSpace(input.PICName),
		PICEmail:        strings.ToLower(strings.TrimSpace(input.PICEmail)),
		PICPhone:        cleanOptional(input.PICPhone),
		PICRole:         cleanOptional(input.PICRole),
		Message:         cleanOptional(input.Message),
		Status:          domain.SchoolRegistrationPending,
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

	// TODO(db-migration): once srr_usr_id actually exists in the database,
	// every newly created request will have RequesterUserID populated by
	// Create(). Requests that predate the migration will have an empty
	// RequesterUserID and are intentionally rejected here rather than
	// silently approved without an owning account.
	if strings.TrimSpace(existing.RequesterUserID) == "" {
		return nil, errors.New("school registration request has no linked requester account")
	}

	requester, err := s.userService.GetByID(existing.RequesterUserID)
	if err != nil {
		return nil, errors.New("school registration requester account not found")
	}

	schoolCode := strings.TrimSpace(input.SchoolCode)
	schoolName := valueOrDefault(input.SchoolName, existing.SchoolName)
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
	if note != nil && len(*note) > 1000 {
		return nil, errors.New("school registration approval note exceeds 1000 characters")
	}

	now := time.Now()
	schoolPhone := ""
	if existing.PICPhone != nil {
		schoolPhone = *existing.PICPhone
	}

	school := &domain.School{
		Name:  schoolName,
		Code:  schoolCode,
		Email: requester.Email,
		Phone: schoolPhone,
	}

	result, err := s.repo.ApprovePending(id, school, existing.RequesterUserID, "admin", reviewerID, now, note)
	if err != nil {
		return nil, err
	}

	if err := s.emailService.SendSchoolMemberAddedToSchool(requester.Email, result.School.Name, "admin"); err != nil {
		fmt.Printf("[Email Warning] failed to send school registration approval notification school_id=%s email=%s error=%s\n", result.School.ID, maskEmail(requester.Email), err.Error())
	}

	return &dto.ApproveSchoolRegistrationRequestResponseDTO{
		Message: "School registration request approved",
		Request: mapSchoolRegistrationRequestDetail(result.Request),
		School: dto.ApprovedSchoolDTO{
			SchoolID:   result.School.ID,
			SchoolCode: result.School.Code,
			SchoolName: result.School.Name,
		},
		Admin: dto.ApprovedAdminDTO{
			UserID:       requester.ID,
			FullName:     requester.FullName,
			Email:        requester.Email,
			SchoolUserID: result.SchoolUser.ID,
			Role:         "admin",
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
		RequestID:       request.ID,
		RequesterUserID: request.RequesterUserID,
		SchoolName:      request.SchoolName,
		NPSN:            request.NPSN,
		PICName:         request.PICName,
		PICEmail:        request.PICEmail,
		PICPhone:        request.PICPhone,
		PICRole:         request.PICRole,
		Message:         request.Message,
		Status:          string(request.Status),
		ReviewedBy:      request.ReviewedBy,
		ReviewedAt:      formatAPITimePtr(request.ReviewedAt),
		ReviewNote:      request.ReviewNote,
		CreatedAt:       formatAPITime(request.CreatedAt),
		UpdatedAt:       formatAPITime(request.UpdatedAt),
	}
}
