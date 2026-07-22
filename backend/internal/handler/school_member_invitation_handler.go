package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SchoolMemberInvitationHandler struct {
	service service.SchoolMemberInvitationService
}

func NewSchoolMemberInvitationHandler(service service.SchoolMemberInvitationService) *SchoolMemberInvitationHandler {
	return &SchoolMemberInvitationHandler{service: service}
}

func (h *SchoolMemberInvitationHandler) Create(c *gin.Context) {
	schoolID, ok := getSchoolMemberInvitationSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konteks sekolah aktif wajib tersedia."})
		return
	}
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.CreateSchoolMemberInvitationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	actor := buildActorContext(c, domain.LogScopeSchool)
	response, err := h.service.Create(actor, schoolID, input)
	if err != nil {
		handleSchoolMemberInvitationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *SchoolMemberInvitationHandler) List(c *gin.Context) {
	schoolID, ok := getSchoolMemberInvitationSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konteks sekolah aktif wajib tersedia."})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	response, err := h.service.List(schoolID, c.Query("status"), page, limit)
	if err != nil {
		handleSchoolMemberInvitationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchoolMemberInvitationHandler) Revoke(c *gin.Context) {
	schoolID, ok := getSchoolMemberInvitationSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konteks sekolah aktif wajib tersedia."})
		return
	}

	actor := buildActorContext(c, domain.LogScopeSchool)
	response, err := h.service.Revoke(actor, schoolID, c.Param("id"))
	if err != nil {
		handleSchoolMemberInvitationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "School member invitation revoked",
		"invitation": response,
	})
}

func getSchoolMemberInvitationSchoolID(c *gin.Context) (string, bool) {
	if value, exists := c.Get("school_id"); exists {
		if schoolID, ok := value.(string); ok && schoolID != "" {
			return schoolID, true
		}
	}
	schoolID := c.GetHeader("SchoolId")
	return schoolID, schoolID != ""
}

func handleSchoolMemberInvitationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrSchoolMemberInvitationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
	case errors.Is(err, repository.ErrSchoolMemberInvitationNotRevocable):
		c.JSON(http.StatusConflict, gin.H{"error": "Invitation cannot be revoked"})
	case err.Error() == "pending invitation already exists for this email and role":
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
