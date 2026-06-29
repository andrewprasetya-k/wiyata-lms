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

type FeedHandler struct {
	service        service.FeedService
	commentService service.CommentService
	classService   service.ClassService
	notifService   service.NotificationService
}

func NewFeedHandler(service service.FeedService, commentService service.CommentService, classService service.ClassService, notifService service.NotificationService) *FeedHandler {
	return &FeedHandler{
		service:        service,
		commentService: commentService,
		classService:   classService,
		notifService:   notifService,
	}
}

var feedNotificationTypes = []string{
	domain.NotifFeedPosted,
	domain.NotifCommentAdded,
}

func (h *FeedHandler) Create(c *gin.Context) {
	var input dto.CreateFeedDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if input.SchoolID != "" && input.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: schoolId does not match active school"})
		return
	}
	if len(input.MediaIDs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feed attachments are not supported in this MVP"})
		return
	}

	feed := domain.Feed{
		SchoolID:  schoolID,
		ClassID:   input.ClassID,
		Content:   input.Content,
		CreatedBy: userID,
	}

	if err := h.service.Create(&feed, userID, getFeedActiveRoles(c)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feed posted"})
}

func (h *FeedHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	feed, err := h.service.GetByID(id, schoolID, middleware.GetUserID(c), getFeedActiveRoles(c))
	if err != nil {
		HandleError(c, err)
		return
	}

	count, _ := h.commentService.CountBySource(string(domain.SourceFeed), feed.ID, schoolID)
	c.JSON(http.StatusOK, h.mapToResponse(feed, count))
}

func (h *FeedHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateFeedDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if len(input.MediaIDs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feed attachments are not supported in this MVP"})
		return
	}

	if err := h.service.Update(id, schoolID, middleware.GetUserID(c), getFeedActiveRoles(c), input.Content); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed updated"})
}

func (h *FeedHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if err := h.service.Delete(id, schoolID, middleware.GetUserID(c), getFeedActiveRoles(c)); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feed deleted"})
}

func (h *FeedHandler) GetByClass(c *gin.Context) {
	classID := c.Param("classId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	schoolID, ok := getFeedActiveSchoolID(c)
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

	// 2. Get Feeds
	feeds, total, err := h.service.GetByClass(classID, schoolID, middleware.GetUserID(c), getFeedActiveRoles(c), page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var feedsDTO []dto.FeedResponseDTO
	for _, f := range feeds {
		count, _ := h.commentService.CountBySource(string(domain.SourceFeed), f.ID, schoolID)
		feedsDTO = append(feedsDTO, h.mapToResponse(f, count))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.ClassWithFeedsDTO{
		Class: dto.ClassHeaderDTO{
			ID:    class.ID,
			Title: class.Title,
			Code:  class.Code,
		},
		Data: dto.PaginatedResponse{
			Data:       feedsDTO,
			TotalItems: total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(totalPages),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *FeedHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	count, err := h.notifService.GetFeedUnreadCount(userID, schoolID, feedNotificationTypes)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, count)
}

func (h *FeedHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	schoolID, ok := getFeedActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.notifService.MarkFeedNotificationsRead(userID, schoolID, feedNotificationTypes); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed notifications marked as read"})
}

func (h *FeedHandler) mapToResponse(f *domain.Feed, commentCount int) dto.FeedResponseDTO {
	atts := make([]dto.MediaResponseDTO, 0, len(f.Attachments))
	for _, a := range f.Attachments {
		atts = append(atts, dto.MediaResponseDTO{
			ID:       a.Media.ID,
			Name:     a.Media.Name,
			FileSize: a.Media.FileSize,
			MimeType: a.Media.MimeType,
			FileURL:  a.Media.FileURL,
		})
	}

	return dto.FeedResponseDTO{
		ID:           f.ID,
		Content:      f.Content,
		CreatorName:  f.Creator.FullName,
		CreatedAt:    f.CreatedAt.Format("02-01-2006 15:04:05"),
		Attachments:  atts,
		CommentCount: commentCount,
	}
}

func getFeedActiveSchoolID(c *gin.Context) (string, bool) {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value, true
		}
	}
	if value := c.GetHeader("SchoolId"); value != "" {
		return value, true
	}
	return "", false
}

func getFeedActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}
