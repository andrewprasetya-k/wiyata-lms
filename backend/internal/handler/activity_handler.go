package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	service service.ActivityService
}

func NewActivityHandler(service service.ActivityService) *ActivityHandler {
	return &ActivityHandler{service: service}
}

func (h *ActivityHandler) GetAcademicActivity(c *gin.Context) {
	from, ok := parseActivityDate(c, "from")
	if !ok {
		return
	}

	to, ok := parseActivityDate(c, "to")
	if !ok {
		return
	}

	result, err := h.service.GetAcademicActivity(
		middleware.GetUserID(c),
		getActivityActiveSchoolID(c),
		getActivityActiveRoles(c),
		from,
		to,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseActivityDate(c *gin.Context, key string) (*time.Time, bool) {
	value := c.Query(key)
	if value == "" {
		return nil, true
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " must use YYYY-MM-DD format"})
		return nil, false
	}

	return &parsed, true
}

func getActivityActiveSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func getActivityActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}
