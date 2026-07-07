package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AcademicYearHandler struct {
	service       service.AcademicYearService
	schoolService service.SchoolService
}

func NewAcademicYearHandler(service service.AcademicYearService, schoolService service.SchoolService) *AcademicYearHandler {
	return &AcademicYearHandler{
		service:       service,
		schoolService: schoolService,
	}
}

func (h *AcademicYearHandler) Create(c *gin.Context) {
	var input dto.CreateAcademicYearDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	acy := domain.AcademicYear{
		SchoolID: input.SchoolID,
		Name:     input.Name,
	}

	if err := h.service.Create(&acy); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&acy))
}

func (h *AcademicYearHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	years, total, err := h.service.FindAll(search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.AcademicYearResponseDTO
	for _, y := range years {
		response = append(response, h.mapToResponse(y))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	paginatedResponse := dto.PaginatedResponse{
		Data:       response,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}
	c.JSON(http.StatusOK, paginatedResponse)
}

func (h *AcademicYearHandler) GetBySchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")

	// Get school header
	school, err := h.schoolService.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	years, err := h.service.GetBySchool(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	var data []dto.AcademicYearResponseDTO
	for _, y := range years {
		data = append(data, h.mapToResponse(y))
	}

	response := dto.AcademicYearWithSchoolDTO{
		School: h.mapSchoolToHeader(school),
		Data:   data,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AcademicYearHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	acy, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(acy))
}

func getAcademicYearSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *AcademicYearHandler) Update(c *gin.Context) {
	id := c.Param("id")
	schoolID := getAcademicYearSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.UpdateAcademicYearDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	acy, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if acy.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: academic year does not belong to active school"})
		return
	}

	if input.Name != nil {
		acy.Name = *input.Name
	}

	if err := h.service.Update(acy); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(acy))
}

func (h *AcademicYearHandler) Activate(c *gin.Context) {
	id := c.Param("id")
	schoolID := getAcademicYearSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	acy, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if acy.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: academic year does not belong to active school"})
		return
	}

	if err := h.service.Activate(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Academic year activated successfully"})
}

func (h *AcademicYearHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")
	schoolID := getAcademicYearSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	acy, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if acy.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: academic year does not belong to active school"})
		return
	}

	if err := h.service.Deactivate(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Academic year deactivated successfully"})
}

func (h *AcademicYearHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	schoolID := getAcademicYearSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	acy, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if acy.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: academic year does not belong to active school"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Academic year deleted successfully"})
}

func (h *AcademicYearHandler) mapToResponse(acy *domain.AcademicYear) dto.AcademicYearResponseDTO {
	return dto.AcademicYearResponseDTO{
		ID:         acy.ID,
		SchoolID:   acy.SchoolID,
		SchoolName: acy.School.Name,
		SchoolCode: acy.School.Code,
		Name:       acy.Name,
		IsActive:   acy.IsActive,
		CreatedAt:  formatAPITime(acy.CreatedAt),
	}
}

func (h *AcademicYearHandler) mapSchoolToHeader(s *domain.School) dto.SchoolHeaderDTO {
	return dto.SchoolHeaderDTO{
		ID:     s.ID,
		Name:   s.Name,
		Code:   s.Code,
		LogoID: s.LogoID,
	}
}
