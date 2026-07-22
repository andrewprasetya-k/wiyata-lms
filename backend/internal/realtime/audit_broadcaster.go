package realtime

import "backend/internal/events"

const AuditPlatformRoom = "platform"

type AuditHubBroadcaster struct {
	hub *Hub
}

func NewAuditHubBroadcaster(hub *Hub) *AuditHubBroadcaster {
	return &AuditHubBroadcaster{hub: hub}
}

func (b *AuditHubBroadcaster) BroadcastSchoolEvent(schoolID string, event events.AuditEvent) {
	b.hub.BroadcastToRoom(schoolID, event)
}

func (b *AuditHubBroadcaster) BroadcastPlatformEvent(event events.AuditEvent) {
	b.hub.BroadcastToRoom(AuditPlatformRoom, event)
}
