package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SchoolHandler struct {
	service service.SchoolService
}

func NewSchoolHandler(service service.SchoolService) *SchoolHandler {
	return &SchoolHandler{service: service}
}

// Create
func (h *SchoolHandler) CreateSchool(c *gin.Context) {
	var input dto.CreateSchoolDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	school := domain.School{
		Name:    input.Name,
		Code:    input.Code,
		LogoID:  input.LogoID,
		Address: input.Address,
		Email:   input.Email,
		Phone:   input.Phone,
		Website: input.Website,
	}

	if err := h.service.CreateSchool(&school); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&school))
}

// Get Schools (with filter)
func (h *SchoolHandler) GetSchools(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	search := c.Query("search")
	sortBy := c.Query("sortBy")
	order := c.Query("order")

	schools, total, err := h.service.GetSchools(search, status, page, limit, sortBy, order)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.SchoolResponseDTO
	for _, s := range schools {
		response = append(response, h.mapToResponse(s))
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

// Get Summary
func (h *SchoolHandler) GetSchoolSummary(c *gin.Context) {
	summary, err := h.service.GetSchoolSummary()
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// Check Availability
func (h *SchoolHandler) CheckCodeAvailability(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	available, err := h.service.CheckCodeAvailability(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"schoolCode": schoolCode,
		"available":  available,
	})
}

// Get By Code
func (h *SchoolHandler) GetSchoolByCode(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	school, err := h.service.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(school))
}

// Update
func (h *SchoolHandler) UpdateSchool(c *gin.Context) {
	var input dto.UpdateSchoolDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	schoolCode := c.Param("schoolCode")
	school, err := h.service.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	if input.Name != nil {
		school.Name = *input.Name
	}
	if input.Code != nil {
		school.Code = *input.Code
	}
	if input.LogoID != nil {
		school.LogoID = input.LogoID
	}
	if input.Address != nil {
		school.Address = *input.Address
	}
	if input.Email != nil {
		school.Email = *input.Email
	}
	if input.Phone != nil {
		school.Phone = *input.Phone
	}
	if input.Website != nil {
		school.Website = input.Website
	}

	if err := h.service.UpdateSchool(school); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(school))
}

// Helper to map domain to DTO
func (h *SchoolHandler) mapToResponse(s *domain.School) dto.SchoolResponseDTO {
	return dto.SchoolResponseDTO{
		ID:        s.ID,
		Name:      s.Name,
		Code:      s.Code,
		LogoID:    s.LogoID,
		Address:   s.Address,
		Email:     s.Email,
		Phone:     s.Phone,
		Website:   s.Website,
		IsDeleted: s.DeletedAt.Valid,
		CreatedAt: s.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}

// restore deleted school
func (h *SchoolHandler) RestoreDeletedSchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	if err := h.service.RestoreDeletedSchool(schoolCode); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "School restored successfully"})
}

// Delete
func (h *SchoolHandler) DeleteSchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	if err := h.service.DeleteSchool(schoolCode); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "School deleted successfully"})
}

// Hard Delete
func (h *SchoolHandler) HardDeleteSchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	if err := h.service.HardDeleteSchool(schoolCode); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "School permanently deleted successfully"})
}
