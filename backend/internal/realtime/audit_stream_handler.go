package realtime

import (
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AuditStreamHandler struct {
	hub             *Hub
	authService     service.AuthService
	wsTicketService service.WSTicketService
	upgrader        websocket.Upgrader
}

func NewAuditStreamHandler(hub *Hub, authService service.AuthService, wsTicketService service.WSTicketService) *AuditStreamHandler {
	return &AuditStreamHandler{
		hub:             hub,
		authService:     authService,
		wsTicketService: wsTicketService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (h *AuditStreamHandler) Stream(c *gin.Context) {
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

	room := strings.TrimSpace(c.Query("channel"))
	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel query is required"})
		return
	}

	context, err := h.authService.GetContext(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	isSuperAdmin := slices.Contains(context.GlobalRoles, "super_admin")

	if room == AuditPlatformRoom {
		if !isSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: platform audit channel requires super admin"})
			return
		}
	} else if !isSuperAdmin && !hasSchoolAdminMembership(context.Memberships, room) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: not an admin of this school"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, userID, room)
	h.hub.Register(client)
	client.ReadLoop()
}

func hasSchoolAdminMembership(memberships []dto.MembershipInfo, schoolID string) bool {
	for _, membership := range memberships {
		if membership.School.ID != schoolID {
			continue
		}
		if slices.Contains(membership.Roles, "admin") {
			return true
		}
	}
	return false
}
