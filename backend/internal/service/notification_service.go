package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/events"
	"backend/internal/repository"
	"math"
)

type NotificationService interface {
	Create(req *dto.CreateNotificationDTO) error
	GetByUserID(userID string, page, limit int, unreadOnly bool) (*dto.NotificationListDTO, error)
	GetUnreadCount(userID string) (*dto.UnreadCountDTO, error)
	GetFeedUnreadCount(userID string, schoolID string, types []string) (*dto.UnreadCountDTO, error)
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error
	MarkFeedNotificationsRead(userID string, schoolID string, types []string) error
	Delete(notificationID string, userID string) error
}

type notificationService struct {
	repo        repository.NotificationRepository
	broadcaster events.SidebarBroadcaster
}

func NewNotificationService(repo repository.NotificationRepository, broadcaster events.SidebarBroadcaster) NotificationService {
	return &notificationService{repo: repo, broadcaster: broadcaster}
}

func (s *notificationService) Create(req *dto.CreateNotificationDTO) error {
	notification := &domain.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		Link:      req.Link,
		RelatedID: req.RelatedID,
	}

	if err := s.repo.Create(notification); err != nil {
		return err
	}

	s.broadcastNotificationChanged(req.UserID)
	return nil
}

func (s *notificationService) GetByUserID(userID string, page, limit int, unreadOnly bool) (*dto.NotificationListDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	notifications, total, err := s.repo.GetByUserID(userID, page, limit, unreadOnly)
	if err != nil {
		return nil, err
	}

	// Get unread count
	unreadCount, err := s.repo.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	notificationDTOs := []dto.NotificationResponseDTO{}
	for _, notification := range notifications {
		notificationDTOs = append(notificationDTOs, dto.NotificationResponseDTO{
			NotificationID: notification.ID,
			Type:           notification.Type,
			Title:          notification.Title,
			Message:        notification.Message,
			Link:           notification.Link,
			IsRead:         notification.IsRead,
			CreatedAt:      formatAPITime(notification.CreatedAt),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.NotificationListDTO{
		Data:        notificationDTOs,
		UnreadCount: unreadCount,
		TotalItems:  total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
	}, nil
}

func (s *notificationService) GetUnreadCount(userID string) (*dto.UnreadCountDTO, error) {
	count, err := s.repo.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UnreadCountDTO{
		UnreadCount: count,
	}, nil
}

func (s *notificationService) GetFeedUnreadCount(userID string, schoolID string, types []string) (*dto.UnreadCountDTO, error) {
	count, err := s.repo.GetFeedUnreadCount(userID, schoolID, types)
	if err != nil {
		return nil, err
	}

	return &dto.UnreadCountDTO{
		UnreadCount: count,
	}, nil
}

func (s *notificationService) MarkAsRead(notificationID string, userID string) error {
	if err := s.repo.MarkAsRead(notificationID, userID); err != nil {
		return err
	}
	s.broadcastNotificationChanged(userID)
	return nil
}

func (s *notificationService) MarkAllAsRead(userID string) error {
	if err := s.repo.MarkAllAsRead(userID); err != nil {
		return err
	}
	s.broadcastNotificationChanged(userID)
	return nil
}

func (s *notificationService) MarkFeedNotificationsRead(userID string, schoolID string, types []string) error {
	if err := s.repo.MarkFeedNotificationsRead(userID, schoolID, types); err != nil {
		return err
	}
	s.broadcastNotificationChanged(userID)
	return nil
}

func (s *notificationService) Delete(notificationID string, userID string) error {
	if err := s.repo.Delete(notificationID, userID); err != nil {
		return err
	}
	s.broadcastNotificationChanged(userID)
	return nil
}

func (s *notificationService) broadcastNotificationChanged(userID string) {
	if s.broadcaster == nil || userID == "" {
		return
	}
	s.broadcaster.BroadcastToUser(userID, events.SidebarEvent{
		Type:   events.SidebarEventTypeNotificationChanged,
		UserID: userID,
	})
}
