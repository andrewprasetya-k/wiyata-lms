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
}

func NewFeedHandler(service service.FeedService, commentService service.CommentService, classService service.ClassService) *FeedHandler {
	return &FeedHandler{
		service:        service,
		commentService: commentService,
		classService:   classService,
	}
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

	// Get user roles from middleware context (set by RequireRole middleware)
	var userRole string
	if roles, exists := c.Get("user_roles"); exists {
		if roleList, ok := roles.([]string); ok && len(roleList) > 0 {
			userRole = roleList[0]
		}
	}

	feed := domain.Feed{
		SchoolID:  input.SchoolID,
		ClassID:   input.ClassID,
		Content:   input.Content,
		CreatedBy: userID,
	}

	if err := h.service.Create(&feed, input.MediaIDs, userID, userRole); err != nil {
		// Check if it's an authorization error
		if err.Error() == "teacher does not teach any subject in this class" ||
			err.Error() == "class does not belong to this school" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feed posted"})
}

func (h *FeedHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	feed, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	count, _ := h.commentService.CountBySource(string(domain.SourceFeed), feed.ID)
	c.JSON(http.StatusOK, h.mapToResponse(feed, count))
}

func (h *FeedHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateFeedDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if input.Content != nil {
		existing.Content = *input.Content
	}

	if err := h.service.Update(id, existing, input.MediaIDs); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed updated"})
}

func (h *FeedHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feed deleted"})
}

func (h *FeedHandler) GetByClass(c *gin.Context) {
	classID := c.Param("classId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 1. Get Class Header
	class, err := h.classService.GetByID(classID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Get Feeds
	feeds, total, err := h.service.GetByClass(classID, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var feedsDTO []dto.FeedResponseDTO
	for _, f := range feeds {
		count, _ := h.commentService.CountBySource(string(domain.SourceFeed), f.ID)
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

func (h *FeedHandler) mapToResponse(f *domain.Feed, commentCount int) dto.FeedResponseDTO {
	var atts []dto.MediaResponseDTO
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
