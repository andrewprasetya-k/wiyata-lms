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
	Create(ctx context.Context, mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline, uploads []UploadFile, actorUserID string, isAdmin bool) error
	FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error)
	GetByID(id string) (*domain.Material, error)
	Update(mat *domain.Material, mediaIDs []string, actorUserID string, isAdmin bool) error
	Delete(id string) error

	// Progress
	UpdateProgress(userID, matID string, status string) error
	GetProgress(userID, matID string) (*domain.MaterialProgress, error)
}

type materialService struct {
	repo         repository.MaterialRepository
	attService   AttachmentService
	mediaRepo    repository.MediaRepository
	storage      storage.Provider
	notifService NotificationService
	sclRepo      repository.SubjectClassRepository
	enrRepo      repository.EnrollmentRepository
}

func NewMaterialService(repo repository.MaterialRepository, attService AttachmentService, mediaRepo repository.MediaRepository, storageProvider storage.Provider, notifService NotificationService, sclRepo repository.SubjectClassRepository, enrRepo repository.EnrollmentRepository) MaterialService {
	if storageProvider == nil {
		storageProvider = storage.NewDisabledStorage()
	}
	return &materialService{
		repo:         repo,
		attService:   attService,
		mediaRepo:    mediaRepo,
		storage:      storageProvider,
		notifService: notifService,
		sclRepo:      sclRepo,
		enrRepo:      enrRepo,
	}
}

func (s *materialService) Create(ctx context.Context, mat *domain.Material, mediaIDs []string, medias []dto.CreateMediaInline, uploads []UploadFile, actorUserID string, isAdmin bool) error {
	mat.Title = strings.TrimSpace(mat.Title)

	if err := validateAttachableMedia(s.mediaRepo, mediaIDs, mat.SchoolID, actorUserID, isAdmin); err != nil {
		return err
	}

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

	if err := replaceSourceAttachments(s.attService, mat.SchoolID, domain.SourceMaterial, mat.ID, mediaIDs); err != nil {
		return err
	}

	// Best-effort: notify students in the class
	if classID, err := s.sclRepo.GetClassIDBySubjectClass(mat.SubjectClassID); err == nil && classID != "" {
		if userIDs, err := s.enrRepo.GetStudentUserIDsByClass(classID); err == nil {
			for _, uid := range userIDs {
				_ = s.notifService.Create(&dto.CreateNotificationDTO{
					UserID:    uid,
					Type:      domain.NotifMaterialAdded,
					Title:     "Materi baru",
					Message:   mat.Title,
					Link:      fmt.Sprintf("/student/subjects/%s/materials/%s", mat.SubjectClassID, mat.ID),
					RelatedID: mat.ID,
				})
			}
		}
	}

	return nil
}

func (s *materialService) FindAll(search string, subjectClassID string, page int, limit int) ([]*domain.Material, int64, error) {
	materials, total, err := s.repo.FindAll(search, subjectClassID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	sourceIDs := make([]string, 0, len(materials))
	for _, mat := range materials {
		sourceIDs = append(sourceIDs, mat.ID)
	}
	attachmentsBySource, err := s.attService.GetBySources(string(domain.SourceMaterial), sourceIDs)
	if err != nil {
		return nil, 0, err
	}

	for _, mat := range materials {
		atts := attachmentsBySource[mat.ID]
		mat.Attachments = make([]domain.Attachment, 0, len(atts))
		for _, a := range atts {
			mat.Attachments = append(mat.Attachments, *a)
		}
	}

	return materials, total, nil
}

func (s *materialService) GetByID(id string) (*domain.Material, error) {
	mat, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Load attachments metadata
	atts, err := s.attService.GetBySource(string(domain.SourceMaterial), id)
	if err != nil {
		return nil, err
	}

	mat.Attachments = nil
	for _, a := range atts {
		mat.Attachments = append(mat.Attachments, *a)
	}

	return mat, nil
}

func (s *materialService) Update(mat *domain.Material, mediaIDs []string, actorUserID string, isAdmin bool) error {
	mat.Title = strings.TrimSpace(mat.Title)
	var attachmentMediaIDs []string
	if mediaIDs != nil {
		var err error
		attachmentMediaIDs, err = prepareAttachableMediaIDs(s.mediaRepo, mediaIDs, mat.SchoolID, actorUserID, isAdmin)
		if err != nil {
			return err
		}
	}

	err := s.repo.Update(mat)
	if err != nil {
		return err
	}

	if mediaIDs != nil {
		if err := replaceSourceAttachments(s.attService, mat.SchoolID, domain.SourceMaterial, mat.ID, attachmentMediaIDs); err != nil {
			return err
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
