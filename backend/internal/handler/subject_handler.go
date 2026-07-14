package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubjectHandler struct {
	service       service.SubjectService
	schoolService service.SchoolService
}

func NewSubjectHandler(service service.SubjectService, schoolService service.SchoolService) *SubjectHandler {
	return &SubjectHandler{
		service:       service,
		schoolService: schoolService,
	}
}

func (h *SubjectHandler) Create(c *gin.Context) {
	var input dto.CreateSubjectDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	subject := domain.Subject{
		SchoolID: input.SchoolID,
		Name:     input.Name,
		Code:     input.Code,
		Color:    input.Color,
	}

	if err := h.service.Create(&subject); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&subject))
}

func (h *SubjectHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	search := c.Query("search")

	schoolID := getSubjectSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	subjects, total, err := h.service.FindAll(schoolID, search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.SubjectResponseDTO
	for _, s := range subjects {
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

func (h *SubjectHandler) GetBySchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")

	// 1. Ambil data sekolah (untuk header)
	school, err := h.schoolService.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Ambil daftar mata pelajaran
	subjects, err := h.service.GetBySchool(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	var subjectsDTO []dto.SubjectResponseDTO
	for _, s := range subjects {
		subjectsDTO = append(subjectsDTO, h.mapToResponse(s))
	}

	response := dto.SchoolWithSubjectsDTO{
		School:   h.mapSchoolToHeader(school),
		Subjects: subjectsDTO,
	}

	c.JSON(http.StatusOK, response)
}

func (h *SubjectHandler) mapSchoolToHeader(s *domain.School) dto.SchoolHeaderDTO {
	return dto.SchoolHeaderDTO{
		ID:     s.ID,
		Name:   s.Name,
		Code:   s.Code,
		LogoID: s.LogoID,
	}
}

func (h *SubjectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID := getSubjectSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	subject, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if subject.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: subject does not belong to active school"})
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(subject))
}

func (h *SubjectHandler) GetByCode(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	subjectCode := c.Param("subjectCode")

	subject, err := h.service.GetByCode(schoolCode, subjectCode)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(subject))
}

func getSubjectSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *SubjectHandler) Update(c *gin.Context) {
	id := c.Param("id")
	schoolID := getSubjectSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.UpdateSubjectDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	subject, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if subject.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: subject does not belong to active school"})
		return
	}

	if input.Name != nil {
		subject.Name = *input.Name
	}
	if input.Code != nil {
		subject.Code = *input.Code
	}
	if input.Color != nil {
		subject.Color = *input.Color
	}

	if err := h.service.Update(subject); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(subject))
}

func (h *SubjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	schoolID := getSubjectSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	subject, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if subject.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: subject does not belong to active school"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subject deleted successfully"})
}

func (h *SubjectHandler) mapToResponse(s *domain.Subject) dto.SubjectResponseDTO {
	return dto.SubjectResponseDTO{
		ID:         s.ID,
		SchoolID:   s.SchoolID,
		SchoolName: s.School.Name,
		SchoolCode: s.School.Code,
		Name:       s.Name,
		Code:       s.Code,
		Color:      s.Color,
		CreatedAt:  formatAPITime(s.CreatedAt),
	}
}
