package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"net/mail"
	"strings"
)

type SchoolRegistrationRequestService interface {
	Create(input dto.CreateSchoolRegistrationRequestDTO) (*dto.CreateSchoolRegistrationRequestResponseDTO, error)
}

type schoolRegistrationRequestService struct {
	repo repository.SchoolRegistrationRequestRepository
}

func NewSchoolRegistrationRequestService(repo repository.SchoolRegistrationRequestRepository) SchoolRegistrationRequestService {
	return &schoolRegistrationRequestService{repo: repo}
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
