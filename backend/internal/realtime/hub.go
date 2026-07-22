package realtime

import "backend/internal/events"

type Hub struct {
	register      chan *Client
	unregister    chan *Client
	broadcast     chan broadcastRequest
	broadcastRoom chan roomBroadcastRequest
	clients       map[string]map[string]map[*Client]bool
}

type broadcastRequest struct {
	schoolID string
	userIDs  []string
	event    Event
}

type roomBroadcastRequest struct {
	room  string
	event events.AuditEvent
}

func NewHub() *Hub {
	return &Hub{
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan broadcastRequest, 32),
		broadcastRoom: make(chan roomBroadcastRequest, 32),
		clients:       make(map[string]map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case request := <-h.broadcast:
			h.broadcastToUsers(request)
		case request := <-h.broadcastRoom:
			h.broadcastToRoom(request)
		}
	}
}

func (h *Hub) Register(client *Client) {
	if h == nil || client == nil {
		return
	}
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	if h == nil || client == nil {
		return
	}
	h.unregister <- client
}

func (h *Hub) BroadcastToUsers(schoolID string, userIDs []string, event Event) {
	if h == nil || schoolID == "" || len(userIDs) == 0 {
		return
	}
	h.broadcast <- broadcastRequest{
		schoolID: schoolID,
		userIDs:  uniqueStrings(userIDs),
		event:    event,
	}
}

func (h *Hub) BroadcastToUser(schoolID string, userID string, event Event) {
	if h == nil || schoolID == "" || userID == "" {
		return
	}
	h.broadcast <- broadcastRequest{
		schoolID: schoolID,
		userIDs:  []string{userID},
		event:    event,
	}
}

func (h *Hub) BroadcastToRoom(room string, event events.AuditEvent) {
	if h == nil || room == "" {
		return
	}
	h.broadcastRoom <- roomBroadcastRequest{room: room, event: event}
}

func (h *Hub) addClient(client *Client) {
	if h.clients[client.SchoolID] == nil {
		h.clients[client.SchoolID] = make(map[string]map[*Client]bool)
	}
	if h.clients[client.SchoolID][client.UserID] == nil {
		h.clients[client.SchoolID][client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.SchoolID][client.UserID][client] = true
}

func (h *Hub) removeClient(client *Client) {
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
		client.Close()
	}
	if len(connections) == 0 {
		delete(users, client.UserID)
	}
	if len(users) == 0 {
		delete(h.clients, client.SchoolID)
	}
}

func (h *Hub) broadcastToUsers(request broadcastRequest) {
	users := h.clients[request.schoolID]
	if users == nil {
		return
	}
	for _, userID := range request.userIDs {
		for client := range users[userID] {
			if err := client.WriteEvent(request.event); err != nil {
				h.removeClient(client)
			}
		}
	}
}

func (h *Hub) broadcastToRoom(request roomBroadcastRequest) {
	users := h.clients[request.room]
	if users == nil {
		return
	}
	for _, connections := range users {
		for client := range connections {
			if err := client.WriteJSON(request.event); err != nil {
				h.removeClient(client)
			}
		}
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
