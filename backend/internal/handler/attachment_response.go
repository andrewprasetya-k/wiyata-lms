package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"net/url"
	"strings"
)

func mapAttachmentMedia(attachment domain.Attachment, schoolID string) (dto.MediaResponseDTO, bool) {
	media := attachment.Media
	if media.ID == "" || media.SchoolID != schoolID || attachment.SchoolID != schoolID {
		return dto.MediaResponseDTO{}, false
	}

	name := strings.TrimSpace(media.Name)
	if name == "" {
		return dto.MediaResponseDTO{}, false
	}

	return dto.MediaResponseDTO{
		ID:           media.ID,
		Name:         name,
		FileSize:     media.FileSize,
		MimeType:     strings.TrimSpace(media.MimeType),
		FileURL:      safeHTTPURL(media.FileURL),
		ThumbnailURL: safeHTTPURL(media.ThumbnailURL),
		OwnerType:    string(media.OwnerType),
		CreatedAt:    media.CreatedAt.Format("02-01-2006 15:04:05"),
	}, true
}

func safeHTTPURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.String()
}
