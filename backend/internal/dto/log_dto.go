package dto

type LogResponseDTO struct {
	ID        string `json:"logId"`
	UserID    string `json:"userId"`
	UserName  string `json:"userName,omitempty"`
	Action    string `json:"action"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"createdAt"`
}

// LogListItemDTO is the Phase 10.9 audit log list row — deliberately without
// Metadata (can be large; the viewer fetches it only when a row is opened
// via the detail endpoint).
type LogListItemDTO struct {
	ID            string `json:"logId"`
	Action        string `json:"action"`
	EntityType    string `json:"entityType,omitempty"`
	EntityID      string `json:"entityId,omitempty"`
	Scope         string `json:"scope,omitempty"`
	Severity      string `json:"severity,omitempty"`
	SchoolID      string `json:"schoolId,omitempty"`
	SchoolName    string `json:"schoolName,omitempty"`
	SchoolCode    string `json:"schoolCode,omitempty"`
	ActorUserID   string `json:"actorUserId"`
	ActorName     string `json:"actorName,omitempty"`
	ActorEmail    string `json:"actorEmail,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
	CreatedAt     string `json:"createdAt"`
}

// LogDetailDTO adds the fields only needed once a row is opened.
type LogDetailDTO struct {
	LogListItemDTO
	Metadata  string `json:"metadata"`
	IPAddress string `json:"ipAddress,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}
