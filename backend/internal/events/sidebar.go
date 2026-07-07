package events

const (
	SidebarEventTypeNotificationChanged = "notification_changed"
	SidebarEventTypeFeedChanged         = "feed_changed"
)

type SidebarEvent struct {
	Type     string         `json:"type"`
	SchoolID string         `json:"schoolId,omitempty"`
	UserID   string         `json:"userId,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
}

type SidebarBroadcaster interface {
	BroadcastToUser(userID string, event SidebarEvent)
	BroadcastToUsers(schoolID string, userIDs []string, event SidebarEvent)
}