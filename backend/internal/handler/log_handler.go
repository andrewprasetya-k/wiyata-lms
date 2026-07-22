package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"backend/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	service      service.LogService
	queryService service.LogQueryService
}

func NewLogHandler(service service.LogService, queryService service.LogQueryService) *LogHandler {
	return &LogHandler{service: service, queryService: queryService}
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

	response := []dto.LogResponseDTO{}
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

// List is the unrestricted, platform-wide audit log search — super-admin
// only (route-gated by RequireSystemSuperAdmin). schoolId is an optional
// query filter here, not a path param, so "all schools" is just "omit it".
func (h *LogHandler) List(c *gin.Context) {
	filter, err := parseLogFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.queryService.Search(filter)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, buildLogListResponse(logs, total, filter.Page, filter.Limit))
}

// GetByID is the unrestricted detail lookup — super-admin only, same gating
// as List. School admins use GetByIDInSchool instead.
func (h *LogHandler) GetByID(c *gin.Context) {
	log, err := h.queryService.GetByID(c.Param("id"))
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapLogDetail(log))
}

// SearchBySchool is the filtered/paginated audit log search pinned to one
// school — same middleware and same active-school double-check as the
// existing GetBySchool above, extended with the Phase 10.9 filter set.
func (h *LogHandler) SearchBySchool(c *gin.Context) {
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

	filter, err := parseLogFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Force the school scope from the URL — never trust a client-supplied
	// schoolId query value here.
	filter.SchoolID = &schoolID

	logs, total, err := h.queryService.Search(filter)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, buildLogListResponse(logs, total, filter.Page, filter.Limit))
}

// GetByIDInSchool is the detail lookup pinned to one school — same
// middleware/double-check pattern as SearchBySchool, plus a check that the
// fetched row actually belongs to that school (defense in depth, same shape
// as GetBySchool's existing activeSchoolID check).
func (h *LogHandler) GetByIDInSchool(c *gin.Context) {
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

	log, err := h.queryService.GetByID(c.Param("id"))
	if err != nil {
		HandleError(c, err)
		return
	}
	if log.SchoolID == nil || *log.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	c.JSON(http.StatusOK, mapLogDetail(log))
}

func parseLogFilter(c *gin.Context) (repository.LogFilter, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := repository.LogFilter{
		Scope:         strings.TrimSpace(c.Query("scope")),
		Action:        strings.TrimSpace(c.Query("action")),
		EntityType:    strings.TrimSpace(c.Query("entityType")),
		Severity:      strings.TrimSpace(c.Query("severity")),
		ActorUserID:   strings.TrimSpace(c.Query("actorUserId")),
		CorrelationID: strings.TrimSpace(c.Query("correlationId")),
		Search:        strings.TrimSpace(c.Query("search")),
		Page:          page,
		Limit:         limit,
	}

	if schoolID := strings.TrimSpace(c.Query("schoolId")); schoolID != "" {
		filter.SchoolID = &schoolID
	}

	dateFrom, err := parseLogDateParam(c.Query("dateFrom"), false)
	if err != nil {
		return filter, err
	}
	filter.DateFrom = dateFrom

	dateTo, err := parseLogDateParam(c.Query("dateTo"), true)
	if err != nil {
		return filter, err
	}
	filter.DateTo = dateTo

	return filter, nil
}

// parseLogDateParam accepts RFC3339 or a bare "2006-01-02" date. For the
// bare-date form on an end-of-range filter (endOfDay), the date is pushed to
// 23:59:59 so "dateTo=2026-01-01" includes the whole day.
func parseLogDateParam(raw string, endOfDay bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}
	return &t, nil
}

func mapLogListItem(l *domain.Log) dto.LogListItemDTO {
	item := dto.LogListItemDTO{
		ID:          l.ID,
		Action:      l.Action,
		ActorUserID: l.UserID,
		ActorName:   l.User.FullName,
		ActorEmail:  l.User.Email,
		CreatedAt:   formatAPITime(l.CreatedAt),
	}
	if l.EntityType != nil {
		item.EntityType = *l.EntityType
	}
	if l.EntityID != nil {
		item.EntityID = *l.EntityID
	}
	if l.Scope != nil {
		item.Scope = *l.Scope
	}
	if l.Severity != nil {
		item.Severity = *l.Severity
	}
	if l.SchoolID != nil {
		item.SchoolID = *l.SchoolID
	}
	if l.School != nil {
		item.SchoolName = l.School.Name
		item.SchoolCode = l.School.Code
	}
	if l.CorrelationID != nil {
		item.CorrelationID = *l.CorrelationID
	}
	return item
}

func mapLogDetail(l *domain.Log) dto.LogDetailDTO {
	detail := dto.LogDetailDTO{
		LogListItemDTO: mapLogListItem(l),
		Metadata:       l.Metadata,
	}
	if l.IPAddress != nil {
		detail.IPAddress = *l.IPAddress
	}
	if l.UserAgent != nil {
		detail.UserAgent = *l.UserAgent
	}
	return detail
}

func buildLogListResponse(logs []*domain.Log, total int64, page int, limit int) dto.PaginatedResponse {
	items := make([]dto.LogListItemDTO, 0, len(logs))
	for _, l := range logs {
		items = append(items, mapLogListItem(l))
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	return dto.PaginatedResponse{
		Data:       items,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}
}
