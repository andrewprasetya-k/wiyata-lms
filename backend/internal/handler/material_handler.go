package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/storage"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MaterialHandler struct {
	service             service.MaterialService
	subjectClassService service.SubjectClassService
}

func NewMaterialHandler(service service.MaterialService, subjectClassService service.SubjectClassService) *MaterialHandler {
	return &MaterialHandler{
		service:             service,
		subjectClassService: subjectClassService,
	}
}

func (h *MaterialHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if it's multipart form (with files) or JSON
	contentType := c.GetHeader("Content-Type")

	if contentType == "application/json" {
		// Original JSON flow
		var input dto.CreateMaterialDTO
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleBindingError(c, err)
			return
		}
		if !h.validateRequestSchool(c, input.SchoolID) {
			return
		}
		if !h.authorizeTeacherForSubjectClass(c, input.SubjectClassID) {
			return
		}

		mat := domain.Material{
			SchoolID:       input.SchoolID,
			SubjectClassID: input.SubjectClassID,
			Title:          input.Title,
			Description:    input.Description,
			Type:           domain.MaterialType(input.Type),
			CreatedBy:      userID,
		}

		if err := h.service.Create(c.Request.Context(), &mat, input.MediaIDs, input.Medias, nil, userID, h.hasActiveRole(c, "admin")); err != nil {
			HandleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Material created successfully"})
		return
	}

	// Multipart form flow (with file uploads)
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	// Parse form fields
	schoolID := c.PostForm("schoolId")
	subjectClassID := c.PostForm("subjectClassId")
	title := c.PostForm("materialTitle")
	description := c.PostForm("materialDesc")
	materialType := c.PostForm("materialType")

	if schoolID == "" || subjectClassID == "" || title == "" || materialType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required fields: schoolId, subjectClassId, materialTitle, materialType"})
		return
	}
	if !h.validateRequestSchool(c, schoolID) {
		return
	}
	if !h.authorizeTeacherForSubjectClass(c, subjectClassID) {
		return
	}

	mat := domain.Material{
		SchoolID:       schoolID,
		SubjectClassID: subjectClassID,
		Title:          title,
		Description:    description,
		Type:           domain.MaterialType(materialType),
		CreatedBy:      userID,
	}

	// Process uploaded files
	files := form.File["files"]
	var uploads []service.UploadFile

	const maxUploadSize = 10 * 1024 * 1024
	for _, file := range files {
		if file.Size > maxUploadSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
			return
		}
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file"})
			return
		}
		mimeType := file.Header.Get("Content-Type")
		uploads = append(uploads, service.UploadFile{
			Name:     file.Filename,
			Size:     file.Size,
			MimeType: mimeType,
			Content:  src,
		})
	}
	defer func() {
		for _, u := range uploads {
			if c, ok := u.Content.(interface{ Close() error }); ok {
				c.Close()
			}
		}
	}()

	if err := h.service.Create(c.Request.Context(), &mat, nil, nil, uploads, userID, h.hasActiveRole(c, "admin")); err != nil {
		if errors.Is(err, storage.ErrNotImplemented) || errors.Is(err, storage.ErrUnavailable) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "File upload to storage is not configured"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Material created successfully with files"})
}

func (h *MaterialHandler) getSchoolContext(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *MaterialHandler) getActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}

func (h *MaterialHandler) hasActiveRole(c *gin.Context, role string) bool {
	for _, activeRole := range h.getActiveRoles(c) {
		if activeRole == role {
			return true
		}
	}
	return false
}

func (h *MaterialHandler) validateRequestSchool(c *gin.Context, requestSchoolID string) bool {
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}
	if requestSchoolID != "" && requestSchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: schoolId does not match active school"})
		return false
	}
	return true
}

func (h *MaterialHandler) authorizeUserForSubjectClassAccess(c *gin.Context, subjectClassID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	allowed, err := h.subjectClassService.UserCanAccessSubjectClass(userID, schoolID, subjectClassID, h.getActiveRoles(c))
	if err != nil {
		HandleError(c, err)
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you cannot access this subject class"})
		return false
	}
	return true
}

func (h *MaterialHandler) authorizeMaterialMutation(c *gin.Context, material *domain.Material) bool {
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}
	if material.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: material does not belong to active school"})
		return false
	}

	if h.hasActiveRole(c, "admin") {
		return true
	}

	if h.hasActiveRole(c, "teacher") {
		return h.authorizeTeacherForSubjectClass(c, material.SubjectClassID)
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
	return false
}

func (h *MaterialHandler) authorizeTeacherForSubjectClass(c *gin.Context, subjectClassID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	ownsSubjectClass, err := h.subjectClassService.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
	if err != nil {
		HandleError(c, err)
		return false
	}
	if !ownsSubjectClass {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: teacher does not teach this subject class"})
		return false
	}
	return true
}

func (h *MaterialHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	subjectClassID := c.Query("subjectClassId")
	if subjectClassID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subjectClassId is required"})
		return
	}
	if !h.authorizeUserForSubjectClassAccess(c, subjectClassID) {
		return
	}

	materials, total, err := h.service.FindAll(search, subjectClassID, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.MaterialResponseDTO
	for _, m := range materials {
		response = append(response, h.mapToResponse(m))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	paginatedResponse := dto.PaginatedResponse{
		Data:       response,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}

	// If subjectClassID is provided, fetch header and wrap response
	if subjectClassID != "" {
		subjectClass, err := h.subjectClassService.GetByID(subjectClassID)
		if err != nil {
			HandleError(c, err)
			return
		}

		c.JSON(http.StatusOK, dto.MaterialListWithSubjectDTO{
			SubjectClass: dto.SubjectClassHeaderDTO{
				ID:           subjectClass.ID,
				SubjectCode:  subjectClass.Subject.Code,
				SubjectName:  subjectClass.Subject.Name,
				SubjectColor: subjectClass.Subject.Color,
				TeacherID:    subjectClass.Teacher.ID,
				TeacherName:  subjectClass.Teacher.User.FullName,
			},
			Data: paginatedResponse,
		})
		return
	}

	c.JSON(http.StatusOK, paginatedResponse)
}

func (h *MaterialHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	mat, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeUserForSubjectClassAccess(c, mat.SubjectClassID) {
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(mat))
}

func (h *MaterialHandler) UpdateProgress(c *gin.Context) {
	var input dto.UpdateProgressDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.UpdateProgress(userID, input.MaterialID, input.Status); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated"})
}

func (h *MaterialHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateMaterialDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	mat, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeMaterialMutation(c, mat) {
		return
	}

	if input.Title != nil {
		mat.Title = *input.Title
	}
	if input.Description != nil {
		mat.Description = *input.Description
	}
	if input.Type != nil {
		mat.Type = domain.MaterialType(*input.Type)
	}

	if err := h.service.Update(mat, input.MediaIDs, middleware.GetUserID(c), h.hasActiveRole(c, "admin")); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material updated successfully"})
}

func (h *MaterialHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	mat, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeMaterialMutation(c, mat) {
		return
	}
	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
}

func (h *MaterialHandler) mapToResponse(m *domain.Material) dto.MaterialResponseDTO {
	atts := make([]dto.MediaResponseDTO, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		if attachment, ok := mapAttachmentMedia(a, m.SchoolID); ok {
			atts = append(atts, attachment)
		}
	}

	return dto.MaterialResponseDTO{
		ID:             m.ID,
		SubjectClassID: m.SubjectClassID,
		SubjectName:    m.SubjectClass.Subject.Name,
		SubjectColor:   m.SubjectClass.Subject.Color,
		Title:          m.Title,
		Description:    m.Description,
		Type:           string(m.Type),
		CreatorName:    m.Creator.FullName,
		CreatedAt:      formatAPITime(m.CreatedAt),
		Attachments:    atts,
	}
}
