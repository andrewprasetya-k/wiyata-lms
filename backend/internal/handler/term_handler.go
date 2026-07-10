package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TermHandler struct {
	service service.TermService
}

func NewTermHandler(service service.TermService) *TermHandler {
	return &TermHandler{service: service}
}

func (h *TermHandler) Create(c *gin.Context) {
	var input dto.CreateTermDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	term := domain.Term{
		AcademicYearID: input.AcademicYearID,
		Name:           input.Name,
	}

	if err := h.service.Create(&term); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&term))
}

func (h *TermHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	terms, total, err := h.service.FindAll(schoolID, search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.TermResponseDTO
	for _, t := range terms {
		response = append(response, h.mapToResponse(t))
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

func (h *TermHandler) GetByAcademicYear(c *gin.Context) {
	acyID := c.Param("academicYearId")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	terms, err := h.service.GetByAcademicYear(acyID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.TermResponseDTO
	for _, t := range terms {
		response = append(response, h.mapToResponse(t))
	}

	c.JSON(http.StatusOK, response)
}

func (h *TermHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	term, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if term.AcademicYear.School.ID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: term does not belong to active school"})
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(term))
}

func getTermSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *TermHandler) Update(c *gin.Context) {
	id := c.Param("id")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.UpdateTermDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	term, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if term.AcademicYear.School.ID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: term does not belong to active school"})
		return
	}

	if input.Name != nil {
		term.Name = *input.Name
	}

	if err := h.service.Update(term); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(term))
}

func (h *TermHandler) Activate(c *gin.Context) {
	id := c.Param("id")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	term, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if term.AcademicYear.School.ID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: term does not belong to active school"})
		return
	}

	if err := h.service.Activate(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Term activated successfully"})
}

func (h *TermHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	term, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if term.AcademicYear.School.ID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: term does not belong to active school"})
		return
	}

	if err := h.service.Deactivate(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Term deactivated successfully"})
}

func (h *TermHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	schoolID := getTermSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	term, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if term.AcademicYear.School.ID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: term does not belong to active school"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Term deleted successfully"})
}

func (h *TermHandler) mapToResponse(t *domain.Term) dto.TermResponseDTO {
	return dto.TermResponseDTO{
		ID:               t.ID,
		AcademicYearID:   t.AcademicYearID,
		AcademicYearName: t.AcademicYear.Name,
		SchoolName:       t.AcademicYear.School.Name,
		Name:             t.Name,
		IsActive:         t.IsActive,
		CreatedAt:        formatAPITime(t.CreatedAt),
	}
}
