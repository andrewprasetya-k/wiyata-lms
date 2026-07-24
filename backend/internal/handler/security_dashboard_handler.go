package handler

import (
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SecurityDashboardHandler struct {
	service service.SecurityDashboardService
}

func NewSecurityDashboardHandler(service service.SecurityDashboardService) *SecurityDashboardHandler {
	return &SecurityDashboardHandler{service: service}
}

// GetAdminSecurityDashboard is GET /dashboard/admin/:schoolId/security —
// same middleware (RequireSchoolMember + RequireRole "admin") and the same
// active-school double-check as GetAdminDashboard/log_handler.go's
// school-pinned routes: the path's schoolId must match the caller's active
// school context, never a client-supplied value trusted on its own.
func (h *SecurityDashboardHandler) GetAdminSecurityDashboard(c *gin.Context) {
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

	dashboard, err := h.service.GetDashboard(&schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetSuperAdminSecurityDashboard is GET /dashboard/super-admin/security —
// unrestricted, platform-wide — route-gated by RequireSystemSuperAdmin,
// same as GetSuperAdminDashboard and the unrestricted /logs endpoints.
func (h *SecurityDashboardHandler) GetSuperAdminSecurityDashboard(c *gin.Context) {
	dashboard, err := h.service.GetDashboard(nil)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
