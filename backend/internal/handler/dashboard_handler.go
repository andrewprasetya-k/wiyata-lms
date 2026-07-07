package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetStudentDashboard(c *gin.Context) {
	userID := c.Param("userId")

	callerID := middleware.GetUserID(c)
	if callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if callerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	dashboard, err := h.service.GetStudentDashboard(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) GetTeacherDashboard(c *gin.Context) {
	schoolUserID := c.Param("schoolUserId")

	callerSchoolUserID, exists := c.Get("school_user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if callerSchoolUserID.(string) != schoolUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	dashboard, err := h.service.GetTeacherDashboard(schoolUserID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) GetAdminDashboard(c *gin.Context) {
	schoolID := c.Param("schoolId")

	activeSchoolID, exists := c.Get("school_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if activeSchoolID.(string) != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	dashboard, err := h.service.GetAdminDashboard(schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
