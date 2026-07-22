package realtime

import (
	"backend/internal/events"
	"sync"
)

type SidebarHub struct {
	register   chan *SidebarClient
	unregister chan *SidebarClient
	broadcast  chan sidebarBroadcastRequest
	clients    map[string]map[string]map[*SidebarClient]bool
	mu         sync.RWMutex
}

type sidebarBroadcastRequest struct {
	schoolID string
	userIDs  []string
	event    events.SidebarEvent
}

type SidebarClient struct {
	UserID   string
	SchoolID string
	Events   chan events.SidebarEvent
}

func NewSidebarHub() *SidebarHub {
	return &SidebarHub{
		register:   make(chan *SidebarClient),
		unregister: make(chan *SidebarClient),
		broadcast:  make(chan sidebarBroadcastRequest, 64),
		clients:    make(map[string]map[string]map[*SidebarClient]bool),
	}
}

func (h *SidebarHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case request := <-h.broadcast:
			h.broadcastToUsers(request)
		}
	}
}

func (h *SidebarHub) Register(client *SidebarClient) {
	if h == nil || client == nil {
		return
	}
	h.register <- client
}

func (h *SidebarHub) Unregister(client *SidebarClient) {
	if h == nil || client == nil {
		return
	}
	h.unregister <- client
}

func (h *SidebarHub) BroadcastToUsers(schoolID string, userIDs []string, event events.SidebarEvent) {
	if h == nil || len(userIDs) == 0 {
		return
	}
	h.broadcast <- sidebarBroadcastRequest{
		schoolID: schoolID,
		userIDs:  uniqueStrings(userIDs),
		event:    event,
	}
}

func (h *SidebarHub) BroadcastToUser(userID string, event events.SidebarEvent) {
	if h == nil || userID == "" {
		return
	}
	h.broadcast <- sidebarBroadcastRequest{
		userIDs: []string{userID},
		event:   event,
	}
}

func (h *SidebarHub) addClient(client *SidebarClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.SchoolID] == nil {
		h.clients[client.SchoolID] = make(map[string]map[*SidebarClient]bool)
	}
	if h.clients[client.SchoolID][client.UserID] == nil {
		h.clients[client.SchoolID][client.UserID] = make(map[*SidebarClient]bool)
	}
	h.clients[client.SchoolID][client.UserID][client] = true
}

func (h *SidebarHub) removeClient(client *SidebarClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	users := h.clients[client.SchoolID]
	if users == nil {
		return
	}
	connections := users[client.UserID]
	if connections == nil {
		return
	}
	if _, ok := connections[client]; ok {
		delete(connections, client)
		close(client.Events)
	}
	if len(connections) == 0 {
		delete(users, client.UserID)
	}
	if len(users) == 0 {
		delete(h.clients, client.SchoolID)
	}
}

func (h *SidebarHub) broadcastToUsers(request sidebarBroadcastRequest) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if request.schoolID != "" {
		users := h.clients[request.schoolID]
		if users == nil {
			return
		}
		for _, userID := range request.userIDs {
			for client := range users[userID] {
				h.trySend(client, request.event)
			}
		}
		return
	}

	for _, users := range h.clients {
		for _, userID := range request.userIDs {
			for client := range users[userID] {
				h.trySend(client, request.event)
			}
		}
	}
}

func (h *SidebarHub) trySend(client *SidebarClient, event events.SidebarEvent) {
	if client == nil {
		return
	}
	select {
	case client.Events <- event:
	default:
	}
}
