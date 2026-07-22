package events

type AuditEvent struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Payload any    `json:"payload"`
}

const AuditEventTypeCreated = "audit_log_created"

type AuditBroadcaster interface {
	BroadcastSchoolEvent(schoolID string, event AuditEvent)
	BroadcastPlatformEvent(event AuditEvent)
}
