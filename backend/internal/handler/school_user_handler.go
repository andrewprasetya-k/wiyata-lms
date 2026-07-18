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

type SchoolUserHandler struct {
	service       service.SchoolUserService
	schoolService service.SchoolService
	rbacService   service.RBACService
}

func NewSchoolUserHandler(service service.SchoolUserService, schoolService service.SchoolService, rbacService service.RBACService) *SchoolUserHandler {
	return &SchoolUserHandler{
		service:       service,
		schoolService: schoolService,
		rbacService:   rbacService,
	}
}

func (h *SchoolUserHandler) Enroll(c *gin.Context) {
	var input dto.AddSchoolUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	scu := domain.SchoolUser{
		UserID:   input.UserID,
		SchoolID: input.SchoolID,
	}

	if err := h.service.Enroll(&scu); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User enrolled to school successfully"})
}

func (h *SchoolUserHandler) GetMembersBySchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	search := c.Query("search")

	// 1. Ambil data sekolah (untuk header)
	school, err := h.schoolService.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Ambil daftar anggota
	members, total, err := h.service.GetMembersBySchool(schoolCode, search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	membersDTO := []dto.SchoolUserResponseDTO{}
	for _, m := range members {
		var roles []string
		for _, ur := range m.Roles {
			roles = append(roles, ur.Role.Name)
		}

		membersDTO = append(membersDTO, dto.SchoolUserResponseDTO{
			ID:        m.ID,
			UserID:    m.UserID,
			FullName:  m.User.FullName,
			Email:     m.User.Email,
			SchoolID:  m.SchoolID,
			Roles:     roles,
			CreatedAt: formatAPITime(m.CreatedAt),
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.SchoolWithMembersDTO{
		School: h.mapSchoolToHeader(school),
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

func (h *SchoolUserHandler) mapSchoolToHeader(s *domain.School) dto.SchoolHeaderDTO {
	return dto.SchoolHeaderDTO{
		ID:     s.ID,
		Name:   s.Name,
		Code:   s.Code,
		LogoID: s.LogoID,
	}
}

func (h *SchoolUserHandler) GetSchoolsByUser(c *gin.Context) {
	userID := c.Param("userId")

	requesterID := middleware.GetUserID(c)
	if requesterID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if requesterID != userID {
		isSuperAdmin, err := h.rbacService.IsSuperAdmin(requesterID)
		if err != nil {
			HandleError(c, err)
			return
		}
		if !isSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: cannot access another user's school memberships"})
			return
		}
	}

	schools, err := h.service.GetSchoolsByUser(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := []dto.SchoolUserResponseDTO{}
	for _, s := range schools {
		response = append(response, dto.SchoolUserResponseDTO{
			ID:         s.ID,
			UserID:     s.UserID,
			SchoolID:   s.SchoolID,
			SchoolName: s.School.Name,
			SchoolCode: s.School.Code,
			CreatedAt:  formatAPITime(s.CreatedAt),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchoolUserHandler) Unenroll(c *gin.Context) {
	userId := c.Param("userId")
	if err := h.service.Unenroll(userId); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User unenrolled from school successfully"})
}
