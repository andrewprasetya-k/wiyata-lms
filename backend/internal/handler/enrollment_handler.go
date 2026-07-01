package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EnrollmentHandler struct {
	service      service.EnrollmentService
	classService service.ClassService
}

func NewEnrollmentHandler(service service.EnrollmentService, classService service.ClassService) *EnrollmentHandler {
	return &EnrollmentHandler{
		service:      service,
		classService: classService,
	}
}

func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	var input dto.CreateEnrollmentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if input.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: schoolId does not match active school"})
		return
	}

	if err := h.service.Enroll(input.SchoolID, input.ClassID, input.SchoolUserIDs, input.Role); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Users enrolled to class successfully"})
}

func (h *EnrollmentHandler) GetByClass(c *gin.Context) {
	classID := c.Param("classId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	// 1. Get Class Header
	class, err := h.classService.GetByID(classID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Get Members
	results, total, err := h.service.GetByClassInSchool(classID, schoolID, search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var membersDTO []dto.EnrollmentResponseDTO
	for _, r := range results {
		membersDTO = append(membersDTO, h.mapToResponse(r))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.ClassWithMembersDTO{
		Class: dto.ClassHeaderDTO{
			ID:    class.ID,
			Title: class.Title,
			Code:  class.Code,
		},
		Members: dto.PaginatedResponse{
			Data:       membersDTO,
			TotalItems: total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(totalPages),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *EnrollmentHandler) GetByMember(c *gin.Context) {
	schoolUserID := c.Param("schoolUserId")
	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	results, err := h.service.GetByMemberInSchool(schoolUserID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.EnrollmentResponseDTO
	for _, r := range results {
		response = append(response, h.mapToResponse(r))
	}

	c.JSON(http.StatusOK, response)
}

func (h *EnrollmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	r, err := h.service.GetByIDInSchool(id, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := h.mapToResponse(r)

	c.JSON(http.StatusOK, response)
}

func (h *EnrollmentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateEnrollmentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.service.Update(id, schoolID, input.Role); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Enrollment updated"})
}

func (h *EnrollmentHandler) Unenroll(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.service.Unenroll(id, schoolID); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Enrollment removed successfully"})
}

func getActiveSchoolID(c *gin.Context) (string, bool) {
	value, exists := c.Get("school_id")
	if !exists {
		return "", false
	}
	schoolID, ok := value.(string)
	return schoolID, ok && schoolID != ""
}

func (h *EnrollmentHandler) mapToResponse(r *domain.Enrollment) dto.EnrollmentResponseDTO {
	response := dto.EnrollmentResponseDTO{
		ID:           r.ID,
		SchoolID:     r.SchoolID,
		SchoolUserID: r.SchoolUserID,
		ClassID:      r.ClassID,
		Role:         r.Role,
		JoinedAt:     formatAPITime(r.JoinedAt),
	}
	if r.SchoolUser.User.ID != "" {
		response.UserFullName = r.SchoolUser.User.FullName
		response.UserEmail = r.SchoolUser.User.Email
	}
	if r.Class.ID != "" {
		response.ClassTitle = r.Class.Title
	}
	if r.LeftAt != nil {
		response.LeftAt = formatAPITime(*r.LeftAt)
	}
	return response
}
