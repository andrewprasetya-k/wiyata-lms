package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	service service.CommentService
}

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Create(c *gin.Context) {
	var input dto.CreateCommentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	schoolID, ok := getCommentActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	if input.SchoolID != "" && input.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: schoolId does not match active school"})
		return
	}

	comment := domain.Comment{
		SchoolID:   schoolID,
		SourceType: domain.SourceType(input.SourceType),
		SourceID:   input.SourceID,
		UserID:     userID,
		Content:    input.Content,
	}

	if err := h.service.Create(&comment, schoolID, userID, getCommentActiveRoles(c)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Comment posted"})
}

func (h *CommentHandler) GetBySource(c *gin.Context) {
	sourceType := c.Query("type")
	sourceID := c.Query("id")
	userID := middleware.GetUserID(c)
	schoolID, ok := getCommentActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	comments, err := h.service.GetBySource(sourceType, sourceID, schoolID, userID, getCommentActiveRoles(c))
	if err != nil {
		HandleError(c, err)
		return
	}

	response := make([]dto.CommentResponseDTO, 0, len(comments))
	for _, c := range comments {
		response = append(response, h.mapCommentToResponse(c, userID))
	}

	c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	schoolID, ok := getCommentActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	comment, err := h.service.GetByID(id, schoolID, userID, getCommentActiveRoles(c))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapCommentToResponse(comment, userID))
}

func (h *CommentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateCommentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}
	userID := middleware.GetUserID(c)
	schoolID, ok := getCommentActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.service.Update(id, schoolID, userID, getCommentActiveRoles(c), input.Content); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment updated"})
}

func (h *CommentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	schoolID, ok := getCommentActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.service.Delete(id, schoolID, userID, getCommentActiveRoles(c)); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

func (h *CommentHandler) mapCommentToResponse(comment *domain.Comment, userID string) dto.CommentResponseDTO {
	return dto.CommentResponseDTO{
		ID:          comment.ID,
		SourceType:  string(comment.SourceType),
		SourceID:    comment.SourceID,
		Content:     comment.Content,
		CreatorName: comment.User.FullName,
		CreatedAt:   formatAPITime(comment.CreatedAt),
		IsMine:      comment.UserID == userID,
	}
}

func getCommentActiveSchoolID(c *gin.Context) (string, bool) {
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

func getCommentActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}
