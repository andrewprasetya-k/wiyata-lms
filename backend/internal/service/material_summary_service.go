package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/storage"
	"context"
	"errors"
	"regexp"
	"strings"
)

const (
	materialSummaryMaxFileBytes     int64 = 5 * 1024 * 1024
	materialSummaryMaxExtractedText       = 24000
	materialSummaryMinExtractedText       = 80
)

var (
	ErrMaterialSummaryUnsupported  = errors.New("material summary unsupported media type")
	ErrMaterialSummaryFileTooLarge = errors.New("material summary file too large")
	ErrMaterialSummaryExtraction   = errors.New("material summary extraction failed")
	ErrMaterialSummaryProvider     = errors.New("material summary provider unavailable")
	ErrMaterialSummaryAttachment   = errors.New("material summary attachment not found")
	ErrMaterialSummaryStoragePath  = errors.New("material summary storage path is required")
)

type PDFTextExtractor interface {
	ExtractText(ctx context.Context, data []byte) (string, error)
}

type MaterialAISummarizer interface {
	SummarizeMaterialDocument(ctx context.Context, text string) (string, error)
}

type MaterialSummaryService interface {
	Generate(ctx context.Context, material *domain.Material, mediaID string) (*dto.MaterialSummaryResponseDTO, error)
}

type materialSummaryService struct {
	attService AttachmentService
	storage    storage.Provider
	extractor  PDFTextExtractor
	ai         MaterialAISummarizer
	maxBytes   int64
}

func NewMaterialSummaryService(attService AttachmentService, storageProvider storage.Provider, extractor PDFTextExtractor, ai MaterialAISummarizer) MaterialSummaryService {
	if storageProvider == nil {
		storageProvider = storage.NewDisabledStorage()
	}
	if extractor == nil {
		extractor = PDFTextExtractorFunc(func(context.Context, []byte) (string, error) {
			return "", ErrMaterialSummaryExtraction
		})
	}
	if ai == nil {
		ai = disabledMaterialAISummarizer{}
	}
	return &materialSummaryService{
		attService: attService,
		storage:    storageProvider,
		extractor:  extractor,
		ai:         ai,
		maxBytes:   materialSummaryMaxFileBytes,
	}
}

func (s *materialSummaryService) Generate(ctx context.Context, material *domain.Material, mediaID string) (*dto.MaterialSummaryResponseDTO, error) {
	if material == nil || strings.TrimSpace(material.ID) == "" || strings.TrimSpace(mediaID) == "" {
		return nil, ErrMaterialSummaryAttachment
	}

	attachment, err := s.findMaterialAttachment(material, mediaID)
	if err != nil {
		return nil, err
	}
	media := attachment.Media
	if media.SchoolID != material.SchoolID || attachment.SchoolID != material.SchoolID {
		return nil, ErrMaterialSummaryAttachment
	}
	if !strings.EqualFold(strings.TrimSpace(media.MimeType), "application/pdf") {
		return nil, ErrMaterialSummaryUnsupported
	}
	if strings.TrimSpace(media.StoragePath) == "" {
		return nil, ErrMaterialSummaryStoragePath
	}
	if media.FileSize > s.maxBytes {
		return nil, ErrMaterialSummaryFileTooLarge
	}

	data, err := s.storage.Download(ctx, media.StoragePath, s.maxBytes)
	if err != nil {
		if errors.Is(err, storage.ErrFileTooLarge) {
			return nil, ErrMaterialSummaryFileTooLarge
		}
		if errors.Is(err, storage.ErrNotImplemented) || errors.Is(err, storage.ErrUnavailable) {
			return nil, ErrMaterialSummaryProvider
		}
		return nil, ErrMaterialSummaryExtraction
	}

	text, err := s.extractText(ctx, data)
	if err != nil {
		return nil, ErrMaterialSummaryExtraction
	}
	text = normalizeDocumentText(text)
	if len([]rune(text)) < materialSummaryMinExtractedText {
		return nil, ErrMaterialSummaryExtraction
	}
	text = truncateRunes(text, materialSummaryMaxExtractedText)

	summary, err := s.ai.SummarizeMaterialDocument(ctx, text)
	if err != nil {
		return nil, ErrMaterialSummaryProvider
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil, ErrMaterialSummaryProvider
	}

	return &dto.MaterialSummaryResponseDTO{
		Status:  "generated",
		Summary: summary,
		Source: dto.MaterialSummarySourceDTO{
			MaterialID: material.ID,
			MediaID:    media.ID,
			MediaName:  media.Name,
			MimeType:   media.MimeType,
		},
	}, nil
}

func (s *materialSummaryService) extractText(ctx context.Context, data []byte) (text string, err error) {
	defer func() {
		if recover() != nil {
			text = ""
			err = ErrMaterialSummaryExtraction
		}
	}()
	return s.extractor.ExtractText(ctx, data)
}

func (s *materialSummaryService) findMaterialAttachment(material *domain.Material, mediaID string) (*domain.Attachment, error) {
	attachments, err := s.attService.GetBySource(string(domain.SourceMaterial), material.ID)
	if err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}
		if attachment.SourceType != domain.SourceMaterial || attachment.SourceID != material.ID {
			continue
		}
		if attachment.MediaID == mediaID && attachment.Media.ID != "" {
			return attachment, nil
		}
	}
	return nil, ErrMaterialSummaryAttachment
}

type PDFTextExtractorFunc func(context.Context, []byte) (string, error)

func (f PDFTextExtractorFunc) ExtractText(ctx context.Context, data []byte) (string, error) {
	return f(ctx, data)
}

type disabledMaterialAISummarizer struct{}

func (disabledMaterialAISummarizer) SummarizeMaterialDocument(context.Context, string) (string, error) {
	return "", ErrMaterialSummaryProvider
}

var documentWhitespacePattern = regexp.MustCompile(`\s+`)

func normalizeDocumentText(value string) string {
	value = strings.ReplaceAll(value, "\x00", " ")
	value = documentWhitespacePattern.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
