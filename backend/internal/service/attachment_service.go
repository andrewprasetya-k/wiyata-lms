package service

import (
	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

type AttachmentService interface {
	Link(att *domain.Attachment) error
	GetBySource(sourceType string, sourceID string) ([]*domain.Attachment, error)
	GetBySources(sourceType string, sourceIDs []string) (map[string][]*domain.Attachment, error)
	Unlink(id string) error
	UnlinkBySource(sourceType string, sourceID string) error
	ReplaceBySource(schoolID string, sourceType string, sourceID string, mediaIDs []string) error
	// WithTx returns a service instance bound to an existing transaction, so
	// callers can compose attachment operations with other repository writes
	// into one atomic unit.
	WithTx(tx *gorm.DB) AttachmentService
}

type attachmentService struct {
	repo repository.AttachmentRepository
}

func NewAttachmentService(repo repository.AttachmentRepository) AttachmentService {
	return &attachmentService{repo: repo}
}

func (s *attachmentService) WithTx(tx *gorm.DB) AttachmentService {
	return &attachmentService{repo: s.repo.WithTx(tx)}
}

func (s *attachmentService) Link(att *domain.Attachment) error {
	return s.repo.Create(att)
}

func (s *attachmentService) GetBySource(sourceType string, sourceID string) ([]*domain.Attachment, error) {
	return s.repo.GetBySource(domain.SourceType(sourceType), sourceID)
}

func (s *attachmentService) GetBySources(sourceType string, sourceIDs []string) (map[string][]*domain.Attachment, error) {
	grouped := make(map[string][]*domain.Attachment, len(sourceIDs))
	uniqueIDs := make([]string, 0, len(sourceIDs))
	seen := make(map[string]struct{}, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		if sourceID == "" {
			continue
		}
		if _, exists := seen[sourceID]; exists {
			continue
		}
		seen[sourceID] = struct{}{}
		uniqueIDs = append(uniqueIDs, sourceID)
		grouped[sourceID] = make([]*domain.Attachment, 0)
	}

	attachments, err := s.repo.GetBySources(domain.SourceType(sourceType), uniqueIDs)
	if err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		grouped[attachment.SourceID] = append(grouped[attachment.SourceID], attachment)
	}
	return grouped, nil
}

func (s *attachmentService) Unlink(id string) error {
	return s.repo.Delete(id)
}

func (s *attachmentService) UnlinkBySource(sourceType string, sourceID string) error {
	return s.repo.DeleteBySource(domain.SourceType(sourceType), sourceID)
}

func (s *attachmentService) ReplaceBySource(schoolID string, sourceType string, sourceID string, mediaIDs []string) error {
	return s.repo.ReplaceBySource(schoolID, domain.SourceType(sourceType), sourceID, mediaIDs)
}
