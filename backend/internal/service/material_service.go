package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"backend/internal/storage"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UploadFile carries an open file reader and its metadata for material upload.
type UploadFile struct {
	Name     string
	Size     int64
	MimeType string
	Content  io.Reader
}

type MaterialService interface {
	Create(ctx context.Context, mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline, uploads []UploadFile) error
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
	storage    storage.Provider
}

func NewMaterialService(repo repository.MaterialRepository, attService AttachmentService, mediaRepo repository.MediaRepository, storageProvider storage.Provider) MaterialService {
	if storageProvider == nil {
		storageProvider = storage.NewDisabledStorage()
	}
	return &materialService{
		repo:       repo,
		attService: attService,
		mediaRepo:  mediaRepo,
		storage:    storageProvider,
	}
}

func (s *materialService) Create(ctx context.Context, mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline, uploads []UploadFile) error {
	mat.Title = strings.TrimSpace(mat.Title)

	if err := s.repo.Create(mat); err != nil {
		return err
	}

	// Upload files to storage and record media
	for _, u := range uploads {
		ext := filepath.Ext(u.Name)
		objectPath := fmt.Sprintf("schools/%s/%s%s", mat.SchoolID, uuid.NewString(), ext)
		mimeType := u.MimeType
		if strings.TrimSpace(mimeType) == "" {
			mimeType = "application/octet-stream"
		}

		publicURL, err := s.storage.Upload(ctx, objectPath, u.Content, mimeType)
		if err != nil {
			return err
		}

		media := &domain.Media{
			SchoolID:    mat.SchoolID,
			Name:        u.Name,
			FileSize:    u.Size,
			MimeType:    mimeType,
			StoragePath: objectPath,
			FileURL:     publicURL,
			IsPublic:    true,
			OwnerType:   domain.OwnerMaterial,
			OwnerID:     mat.ID,
		}
		if err := s.mediaRepo.Create(media); err != nil {
			// Best-effort cleanup of uploaded object
			_ = s.storage.Delete(ctx, objectPath)
			return err
		}
		mediaIDs = append(mediaIDs, media.ID)
	}

	// Create inline medias (pre-recorded, no upload needed)
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
