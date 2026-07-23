package realtime

import (
	"backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub             *Hub
	chatService     service.ChatService
	wsTicketService service.WSTicketService
	upgrader        websocket.Upgrader
}

func NewWebSocketHandler(hub *Hub, chatService service.ChatService, wsTicketService service.WSTicketService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:             hub,
		chatService:     chatService,
		wsTicketService: wsTicketService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) Chat(c *gin.Context) {
	ticketValue := extractHandshakeTicket(c)
	if ticketValue == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := h.wsTicketService.Consume(ticketValue)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired ticket"})
		return
	}

	schoolID := strings.TrimSpace(c.Query("schoolId"))
	if schoolID == "" {
		schoolID = strings.TrimSpace(c.GetHeader("SchoolId"))
	}
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	allowed, err := h.chatService.CanAccessSchoolChat(userID, schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify chat access"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: chat school access denied"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, userID, schoolID)
	h.hub.Register(client)
	client.ReadLoop()
}

// extractHandshakeTicket reads the short-lived, single-use WS ticket
// (issued via GET /me/ws-ticket) from the query string. Shared by chat,
// audit, and sidebar handshakes. Replaces the previous raw-JWT-via-?token=
// scheme — a real access token may now live somewhere JS on the frontend
// can't read it from, so it can no longer be attached to a WS URL at all.
func extractHandshakeTicket(c *gin.Context) string {
	return strings.TrimSpace(c.Query("ticket"))
}
