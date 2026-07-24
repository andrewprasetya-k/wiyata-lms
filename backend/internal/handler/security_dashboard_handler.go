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
