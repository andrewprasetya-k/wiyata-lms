package handler

import (
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	service service.LogService
}

func NewLogHandler(service service.LogService) *LogHandler {
	return &LogHandler{service: service}
}

func (h *LogHandler) GetBySchool(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	logs, total, err := h.service.GetBySchool(schoolID, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.LogResponseDTO
	for _, l := range logs {
		response = append(response, dto.LogResponseDTO{
			ID:        l.ID,
			UserID:    l.UserID,
			UserName:  l.User.FullName,
			Action:    l.Action,
			Metadata:  l.Metadata,
			CreatedAt: formatAPITime(l.CreatedAt),
		})
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
