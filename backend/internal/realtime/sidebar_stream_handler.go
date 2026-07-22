package realtime

import (
	"backend/internal/dto"
	"backend/internal/events"
	"backend/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SidebarStreamHandler struct {
	hub         *SidebarHub
	authService service.AuthService
	upgrader    websocket.Upgrader
}

func NewSidebarStreamHandler(hub *SidebarHub, authService service.AuthService) *SidebarStreamHandler {
	return &SidebarStreamHandler{
		hub:         hub,
		authService: authService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (h *SidebarStreamHandler) Stream(c *gin.Context) {
	tokenValue := extractHandshakeToken(c)
	if tokenValue == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := parseUserIDFromToken(tokenValue)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	schoolID := strings.TrimSpace(c.Query("schoolId"))
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	context, err := h.authService.GetContext(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	if !hasSchoolMembership(context.Memberships, schoolID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	client := &SidebarClient{
		UserID:   userID,
		SchoolID: schoolID,
		Events:   make(chan events.SidebarEvent, 16),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	headers := c.Writer.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	if _, err := fmt.Fprint(c.Writer, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	closeNotify := c.Request.Context().Done()

	for {
		select {
		case event, ok := <-client.Events:
			if !ok {
				return
			}
			if err := writeSidebarEvent(c.Writer, event); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-closeNotify:
			return
		}
	}
}

func writeSidebarEvent(writer http.ResponseWriter, event events.SidebarEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "data: %s\n\n", payload)
	return err
}

func hasSchoolMembership(memberships []dto.MembershipInfo, schoolID string) bool {
	for _, membership := range memberships {
		if membership.School.ID == schoolID {
			return true
		}
	}
	return false
}
