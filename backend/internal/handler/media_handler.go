package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service service.MediaService
}

func NewMediaHandler(service service.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// Upload handles multipart file upload
func (h *MediaHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	schoolID := c.PostForm("schoolId")
	ownerType := c.PostForm("ownerType")
	ownerID := c.PostForm("ownerId")

	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schoolId is required"})
		return
	}

	// Auto-detect file info
	fileSize := file.Size / (1024 * 1024) // Convert to MB
	if fileSize > 10 {                    // Example limit: 100MB
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}
	mimeType := file.Header.Get("Content-Type")
	fileName := file.Filename
	ext := filepath.Ext(fileName)

	// Generate storage path (you can customize this)
	storagePath := fmt.Sprintf("materials/%s/%s", schoolID, fileName)

	// TODO: Upload to Supabase Storage here
	// For now, we'll create a placeholder URL
	fileURL := fmt.Sprintf("https://placeholder.supabase.co/storage/v1/object/public/%s", storagePath)

	media := domain.Media{
		SchoolID:    schoolID,
		Name:        fileName,
		FileSize:    fileSize,
		MimeType:    mimeType,
		StoragePath: storagePath,
		FileURL:     fileURL,
		IsPublic:    true,
		OwnerType:   domain.OwnerType(ownerType),
		OwnerID:     ownerID,
	}

	if err := h.service.RecordMetadata(&media); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "File uploaded successfully",
		"mediaId":  media.ID,
		"fileName": fileName,
		"fileSize": fileSize,
		"mimeType": mimeType,
		"fileUrl":  fileURL,
		"ext":      ext,
	})
}

// RecordMetadata records metadata of an already uploaded file (e.g., to Supabase/S3)
func (h *MediaHandler) RecordMetadata(c *gin.Context) {
	var input dto.RecordMediaDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	media := domain.Media{
		SchoolID:     input.SchoolID,
		Name:         input.Name,
		FileSize:     input.FileSize,
		MimeType:     input.MimeType,
		StoragePath:  input.StoragePath,
		FileURL:      input.FileURL,
		ThumbnailURL: input.ThumbnailURL,
		IsPublic:     input.IsPublic,
		OwnerType:    domain.OwnerType(input.OwnerType),
		OwnerID:      input.OwnerID,
	}

	if err := h.service.RecordMetadata(&media); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, media)
}

func (h *MediaHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	media, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, media)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Media record deleted"})
}
