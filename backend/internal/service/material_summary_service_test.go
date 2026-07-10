package service

import (
	"backend/internal/domain"
	"backend/internal/storage"
	"context"
	"errors"
	"io"
	"testing"

	"gorm.io/gorm"
)

type materialSummaryAttachmentServiceStub struct {
	attachments []*domain.Attachment
}

func (s *materialSummaryAttachmentServiceStub) Link(*domain.Attachment) error {
	return nil
}

func (s *materialSummaryAttachmentServiceStub) GetBySource(string, string) ([]*domain.Attachment, error) {
	return s.attachments, nil
}

func (s *materialSummaryAttachmentServiceStub) GetBySources(string, []string) (map[string][]*domain.Attachment, error) {
	return nil, nil
}

func (s *materialSummaryAttachmentServiceStub) Unlink(string) error {
	return nil
}

func (s *materialSummaryAttachmentServiceStub) UnlinkBySource(string, string) error {
	return nil
}

func (s *materialSummaryAttachmentServiceStub) ReplaceBySource(string, string, string, []string) error {
	return nil
}

func (s *materialSummaryAttachmentServiceStub) WithTx(*gorm.DB) AttachmentService {
	return s
}

type materialSummaryStorageStub struct {
	data           []byte
	err            error
	downloadCalled bool
	objectPath     string
	maxBytes       int64
}

func (s *materialSummaryStorageStub) Upload(context.Context, string, io.Reader, string) (string, error) {
	return "", nil
}

func (s *materialSummaryStorageStub) Delete(context.Context, string) error {
	return nil
}

func (s *materialSummaryStorageStub) Download(_ context.Context, objectPath string, maxBytes int64) ([]byte, error) {
	s.downloadCalled = true
	s.objectPath = objectPath
	s.maxBytes = maxBytes
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func (s *materialSummaryStorageStub) HealthCheck(context.Context) error {
	return nil
}

func (s *materialSummaryStorageStub) GetPublicURL(string) string {
	return ""
}

type materialSummaryExtractorStub struct {
	text          string
	err           error
	panicOnCall   bool
	extractCalled bool
	data          []byte
}

func (s *materialSummaryExtractorStub) ExtractText(_ context.Context, data []byte) (string, error) {
	s.extractCalled = true
	s.data = data
	if s.panicOnCall {
		panic("malformed pdf")
	}
	if s.err != nil {
		return "", s.err
	}
	return s.text, nil
}

type materialSummaryAIStub struct {
	summary string
	err     error
	called  bool
	text    string
}

func (s *materialSummaryAIStub) SummarizeMaterialDocument(_ context.Context, text string) (string, error) {
	s.called = true
	s.text = text
	if s.err != nil {
		return "", s.err
	}
	return s.summary, nil
}

func TestMaterialSummaryMediaNotAttachedRejectedBeforeStorageAndAI(t *testing.T) {
	storageStub := &materialSummaryStorageStub{}
	aiStub := &materialSummaryAIStub{}
	svc := newMaterialSummaryTestService(nil, storageStub, &materialSummaryExtractorStub{}, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryAttachment) {
		t.Fatalf("expected attachment error, got %v", err)
	}
	if storageStub.downloadCalled {
		t.Fatal("storage should not be called when media is not attached")
	}
	if aiStub.called {
		t.Fatal("AI should not be called when media is not attached")
	}
}

func TestMaterialSummaryCrossSchoolMediaRejectedBeforeStorageAndAI(t *testing.T) {
	storageStub := &materialSummaryStorageStub{}
	aiStub := &materialSummaryAIStub{}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-2", "material-1", materialSummaryPDFMedia("school-2", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, &materialSummaryExtractorStub{}, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryAttachment) {
		t.Fatalf("expected attachment error, got %v", err)
	}
	if storageStub.downloadCalled {
		t.Fatal("storage should not be called for cross-school media")
	}
	if aiStub.called {
		t.Fatal("AI should not be called for cross-school media")
	}
}

func TestMaterialSummaryNonPDFRejectedAsUnsupported(t *testing.T) {
	storageStub := &materialSummaryStorageStub{}
	aiStub := &materialSummaryAIStub{}
	media := materialSummaryPDFMedia("school-1", "media-1", 100)
	media.MimeType = "image/png"
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", media),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, &materialSummaryExtractorStub{}, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryUnsupported) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
	if storageStub.downloadCalled {
		t.Fatal("storage should not be called for unsupported MIME")
	}
	if aiStub.called {
		t.Fatal("AI should not be called for unsupported MIME")
	}
}

func TestMaterialSummaryFileTooLargeRejectedBeforeStorageAndAI(t *testing.T) {
	storageStub := &materialSummaryStorageStub{}
	aiStub := &materialSummaryAIStub{}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", materialSummaryMaxFileBytes+1)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, &materialSummaryExtractorStub{}, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryFileTooLarge) {
		t.Fatalf("expected file too large error, got %v", err)
	}
	if storageStub.downloadCalled {
		t.Fatal("storage should not be called when metadata already exceeds limit")
	}
	if aiStub.called {
		t.Fatal("AI should not be called when metadata exceeds limit")
	}
}

func TestMaterialSummaryStorageTooLargeMapsToFileTooLarge(t *testing.T) {
	storageStub := &materialSummaryStorageStub{err: storage.ErrFileTooLarge}
	aiStub := &materialSummaryAIStub{}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, &materialSummaryExtractorStub{}, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryFileTooLarge) {
		t.Fatalf("expected file too large error, got %v", err)
	}
	if aiStub.called {
		t.Fatal("AI should not be called when storage reports file too large")
	}
}

func TestMaterialSummaryExtractorErrorMapsToExtractionError(t *testing.T) {
	storageStub := &materialSummaryStorageStub{data: []byte("pdf bytes")}
	aiStub := &materialSummaryAIStub{}
	extractorStub := &materialSummaryExtractorStub{err: errors.New("corrupt pdf")}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, extractorStub, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryExtraction) {
		t.Fatalf("expected extraction error, got %v", err)
	}
	if aiStub.called {
		t.Fatal("AI should not be called when extraction fails")
	}
}

func TestMaterialSummaryExtractorPanicMapsToExtractionError(t *testing.T) {
	storageStub := &materialSummaryStorageStub{data: []byte("pdf bytes")}
	aiStub := &materialSummaryAIStub{}
	extractorStub := &materialSummaryExtractorStub{panicOnCall: true}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, extractorStub, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryExtraction) {
		t.Fatalf("expected extraction error, got %v", err)
	}
	if aiStub.called {
		t.Fatal("AI should not be called when extractor panics")
	}
}

func TestMaterialSummaryProviderUnavailableMapsToProviderError(t *testing.T) {
	storageStub := &materialSummaryStorageStub{data: []byte("pdf bytes")}
	aiStub := &materialSummaryAIStub{err: ErrMaterialSummaryProvider}
	extractorStub := &materialSummaryExtractorStub{text: materialSummaryReadableText()}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, extractorStub, aiStub)

	_, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if !errors.Is(err, ErrMaterialSummaryProvider) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestMaterialSummarySuccessSendsNormalizedTextToAI(t *testing.T) {
	storageStub := &materialSummaryStorageStub{data: []byte("pdf bytes")}
	aiStub := &materialSummaryAIStub{summary: "Ringkasan dokumen"}
	extractorStub := &materialSummaryExtractorStub{text: "  Baris pertama.\n\n\tBaris kedua   dengan spasi. \x00 Baris ketiga yang membuat teks cukup panjang untuk diringkas oleh AI.  "}
	attachments := []*domain.Attachment{
		materialSummaryAttachment("school-1", "material-1", materialSummaryPDFMedia("school-1", "media-1", 100)),
	}
	svc := newMaterialSummaryTestService(attachments, storageStub, extractorStub, aiStub)

	result, err := svc.Generate(context.Background(), materialSummaryTestMaterial(), "media-1")

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.Status != "generated" || result.Summary != "Ringkasan dokumen" {
		t.Fatalf("unexpected result: %+v", result)
	}
	expectedText := "Baris pertama. Baris kedua dengan spasi. Baris ketiga yang membuat teks cukup panjang untuk diringkas oleh AI."
	if aiStub.text != expectedText {
		t.Fatalf("expected normalized AI text %q, got %q", expectedText, aiStub.text)
	}
	if string(extractorStub.data) != "pdf bytes" {
		t.Fatalf("extractor received unexpected data: %q", string(extractorStub.data))
	}
	if storageStub.objectPath != "schools/school-1/materials/file.pdf" {
		t.Fatalf("storage used unexpected path: %s", storageStub.objectPath)
	}
	if storageStub.maxBytes != materialSummaryMaxFileBytes {
		t.Fatalf("storage max bytes = %d, want %d", storageStub.maxBytes, materialSummaryMaxFileBytes)
	}
}

func newMaterialSummaryTestService(attachments []*domain.Attachment, storageStub *materialSummaryStorageStub, extractorStub *materialSummaryExtractorStub, aiStub *materialSummaryAIStub) MaterialSummaryService {
	return NewMaterialSummaryService(
		&materialSummaryAttachmentServiceStub{attachments: attachments},
		storageStub,
		extractorStub,
		aiStub,
	)
}

func materialSummaryTestMaterial() *domain.Material {
	return &domain.Material{
		ID:       "material-1",
		SchoolID: "school-1",
	}
}

func materialSummaryPDFMedia(schoolID string, mediaID string, fileSize int64) domain.Media {
	return domain.Media{
		ID:          mediaID,
		SchoolID:    schoolID,
		Name:        "Materi.pdf",
		MimeType:    "application/pdf",
		FileSize:    fileSize,
		StoragePath: "schools/school-1/materials/file.pdf",
	}
}

func materialSummaryAttachment(schoolID string, materialID string, media domain.Media) *domain.Attachment {
	return &domain.Attachment{
		ID:         "attachment-1",
		SchoolID:   schoolID,
		SourceID:   materialID,
		SourceType: domain.SourceMaterial,
		MediaID:    media.ID,
		Media:      media,
	}
}

func materialSummaryReadableText() string {
	return "Teks PDF yang cukup panjang untuk lolos batas minimum ekstraksi dan dapat dikirim ke AI sebagai konten dokumen."
}
