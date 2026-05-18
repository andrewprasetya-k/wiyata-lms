package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"strings"
	"time"
)

type MaterialService interface {
	Create(mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline) error
	FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error)
	GetByID(id string) (*domain.Material, error)
	Update(mat *domain.Material, mediaIDs []string) error
	Delete(id string) error

	// Progress
	UpdateProgress(userID, matID string, status string) error
	GetProgress(userID, matID string) (*domain.MaterialProgress, error)
}

type materialService struct {
	repo       repository.MaterialRepository
	attService AttachmentService
	mediaRepo  repository.MediaRepository
}

func NewMaterialService(repo repository.MaterialRepository, attService AttachmentService, mediaRepo repository.MediaRepository) MaterialService {
	return &materialService{
		repo:       repo,
		attService: attService,
		mediaRepo:  mediaRepo,
	}
}

func (s *materialService) Create(mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline) error {
	mat.Title = strings.TrimSpace(mat.Title)

	err := s.repo.Create(mat)
	if err != nil {
		return err
	}

	// Create new medias if provided
	for _, m := range medias {
		media := &domain.Media{
			SchoolID:     mat.SchoolID,
			Name:         m.Name,
			FileSize:     m.FileSize,
			MimeType:     m.MimeType,
			FileURL:      m.FileURL,
			ThumbnailURL: m.ThumbnailURL,
			IsPublic:     true,
			OwnerType:    domain.OwnerMaterial,
			OwnerID:      mat.ID,
		}
		if err := s.mediaRepo.Create(media); err == nil {
			mediaIDs = append(mediaIDs, media.ID)
		}
	}

	// Link attachments
	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   mat.SchoolID,
			SourceID:   mat.ID,
			SourceType: domain.SourceMaterial,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}

	return nil
}

func (s *materialService) FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error) {
	return s.repo.FindAll(search, subjectClassID, page, limit)
}

func (s *materialService) GetByID(id string) (*domain.Material, error) {
	mat, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Load attachments metadata
	atts, _ := s.attService.GetBySource(string(domain.SourceMaterial), id)
	mat.Attachments = nil
	for _, a := range atts {
		mat.Attachments = append(mat.Attachments, *a)
	}

	return mat, nil
}

func (s *materialService) Update(mat *domain.Material, mediaIDs []string) error {
	mat.Title = strings.TrimSpace(mat.Title)
	err := s.repo.Update(mat)
	if err != nil {
		return err
	}

	if mediaIDs != nil {
		// 1. Unlink existing attachments for this material
		s.attService.UnlinkBySource(string(domain.SourceMaterial), mat.ID)

		// 2. Link new attachments
		for _, mID := range mediaIDs {
			att := &domain.Attachment{
				SchoolID:   mat.SchoolID,
				SourceID:   mat.ID,
				SourceType: domain.SourceMaterial,
				MediaID:    mID,
			}
			s.attService.Link(att)
		}
	}

	return nil
}

func (s *materialService) Delete(id string) error {
	// 1. Unlink all attachments associated with this material
	s.attService.UnlinkBySource(string(domain.SourceMaterial), id)

	// 2. Delete the material
	return s.repo.Delete(id)
}

func (s *materialService) UpdateProgress(userID, matID string, status string) error {
	now := time.Now()
	prog := &domain.MaterialProgress{
		UserID:       userID,
		MaterialID:   matID,
		Status:       domain.StatusProgress(status),
		LastOpenedAt: &now,
	}
	return s.repo.UpsertProgress(prog)
}

func (s *materialService) GetProgress(userID, matID string) (*domain.MaterialProgress, error) {
	return s.repo.GetProgress(userID, matID)
}
