package dto

type CreateNotificationDTO struct {
	UserID    string `json:"userId" binding:"required,uuid"`
	Type      string `json:"type" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Message   string `json:"message"`
	Link      string `json:"link"`
	RelatedID string `json:"relatedId"`
}

type NotificationResponseDTO struct {
	NotificationID string `json:"notificationId"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Link           string `json:"link"`
	IsRead         bool   `json:"isRead"`
	CreatedAt      string `json:"createdAt"`
}

type NotificationListDTO struct {
	Data        []NotificationResponseDTO `json:"data"`
	UnreadCount int                       `json:"unreadCount"`
	TotalItems  int                       `json:"totalItems"`
	Page        int                       `json:"page"`
	Limit       int                       `json:"limit"`
	TotalPages  int                       `json:"totalPages"`
}

type UnreadCountDTO struct {
	UnreadCount int `json:"unreadCount"`
}
