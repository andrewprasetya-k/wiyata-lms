package handler

import (
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

	dashboard, err := h.service.GetStudentDashboard(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) GetTeacherDashboard(c *gin.Context) {
	schoolUserID := c.Param("schoolUserId")

	dashboard, err := h.service.GetTeacherDashboard(schoolUserID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) GetAdminDashboard(c *gin.Context) {
	schoolID := c.Param("schoolId")

	dashboard, err := h.service.GetAdminDashboard(schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
