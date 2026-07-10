package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	service       service.ClassService
	schoolService service.SchoolService
}

func NewClassHandler(service service.ClassService, schoolService service.SchoolService) *ClassHandler {
	return &ClassHandler{
		service:       service,
		schoolService: schoolService,
	}
}

func (h *ClassHandler) Create(c *gin.Context) {
	var input dto.CreateClassDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	class := domain.Class{
		SchoolID:    input.SchoolID,
		TermID:      input.TermID,
		Code:        input.Code,
		Title:       input.Title,
		Description: input.Description,
		CreatedBy:   userID,
	}

	if err := h.service.Create(&class); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&class))
}

func (h *ClassHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	schoolCode := c.Query("schoolCode")
	termID := c.Query("termId")

	schoolID := getClassSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	classes, total, err := h.service.FindAll(search, schoolID, termID, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var data []dto.ClassResponseDTO
	for _, cls := range classes {
		data = append(data, h.mapToResponse(cls))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	paginatedResponse := dto.PaginatedResponse{
		Data:       data,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}

	// If schoolCode is provided, wrap with school header
	if schoolCode != "" {
		school, err := h.schoolService.GetSchoolByCode(schoolCode)
		if err == nil {
			response := dto.ClassListWithSchoolDTO{
				School: h.mapSchoolToHeader(school),
				Data:   paginatedResponse,
			}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	c.JSON(http.StatusOK, paginatedResponse)
}

func (h *ClassHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID := getClassSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	class, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if class.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: class does not belong to active school"})
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(class))
}

func getClassSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *ClassHandler) Update(c *gin.Context) {
	id := c.Param("id")
	schoolID := getClassSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.UpdateClassDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	class, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if class.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: class does not belong to active school"})
		return
	}

	if input.Title != nil {
		class.Title = *input.Title
	}
	if input.Description != nil {
		class.Description = *input.Description
	}
	if input.IsActive != nil {
		class.IsActive = *input.IsActive
	}

	if err := h.service.Update(class); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(class))
}

func (h *ClassHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	schoolID := getClassSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	class, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if class.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: class does not belong to active school"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Class deleted successfully"})
}

func (h *ClassHandler) mapToResponse(c *domain.Class) dto.ClassResponseDTO {
	return dto.ClassResponseDTO{
		ID:               c.ID,
		SchoolID:         c.SchoolID,
		SchoolName:       c.School.Name,
		TermID:           c.TermID,
		TermName:         c.Term.Name,
		AcademicYearName: c.Term.AcademicYear.Name,
		Code:             c.Code,
		Title:            c.Title,
		Description:      c.Description,
		CreatedBy:        c.CreatedBy,
		CreatorName:      c.Creator.FullName,
		IsActive:         c.IsActive,
		CreatedAt:        formatAPITime(c.CreatedAt),
		UpdatedAt:        formatAPITime(c.UpdatedAt),
	}
}

func (h *ClassHandler) mapSchoolToHeader(s *domain.School) dto.SchoolHeaderDTO {
	return dto.SchoolHeaderDTO{
		ID:     s.ID,
		Name:   s.Name,
		Code:   s.Code,
		LogoID: s.LogoID,
	}
}
