package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service service.ChatService
}

func NewChatHandler(service service.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) ListRooms(c *gin.Context) {
	userID := middleware.GetUserID(c)
	schoolID, ok := getChatActiveSchoolID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	rooms, err := h.service.ListMyRooms(userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if rooms == nil {
		rooms = make([]dto.ChatRoomDTO, 0)
	}
	c.JSON(http.StatusOK, dto.ChatRoomsResponseDTO{Rooms: rooms})
}

func (h *ChatHandler) OpenSchoolRoom(c *gin.Context) {
	userID := middleware.GetUserID(c)
	schoolID, ok := getChatActiveSchoolID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	room, err := h.service.OpenSchoolRoom(userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ChatRoomResponseDTO{Room: *room})
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	schoolID, ok := getChatActiveSchoolID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var before *time.Time
	if raw := c.Query("before"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid before timestamp"})
			return
		}
		before = &parsed
	}

	response, err := h.service.ListMessages(userID, schoolID, c.Param("roomId"), limit, before)
	if err != nil {
		HandleError(c, err)
		return
	}
	if response.Messages == nil {
		response.Messages = make([]dto.ChatMessageDTO, 0)
	}
	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) CreateMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	schoolID, ok := getChatActiveSchoolID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.CreateChatMessageDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	message, err := h.service.CreateMessage(userID, schoolID, c.Param("roomId"), input.Content)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	schoolID, ok := getChatActiveSchoolID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	var input dto.MarkChatRoomReadDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := h.service.MarkRead(userID, schoolID, c.Param("roomId"), input.LastReadMessageID); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat room marked as read"})
}

func getChatActiveSchoolID(c *gin.Context) (string, bool) {
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
