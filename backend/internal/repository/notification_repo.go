package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *domain.Notification) error
	GetByUserID(userID string, page, limit int, unreadOnly bool) ([]*domain.Notification, int, error)
	GetUnreadCount(userID string) (int, error)
	GetFeedUnreadCount(userID string, schoolID string, types []string) (int, error)
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error
	MarkFeedNotificationsRead(userID string, schoolID string, types []string) error
	Delete(notificationID string, userID string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByUserID(userID string, page, limit int, unreadOnly bool) ([]*domain.Notification, int, error) {
	var notifications []*domain.Notification
	var total int64

	query := r.db.Where("ntf_usr_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Count total
	query.Model(&domain.Notification{}).Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error

	return notifications, int(total), err
}

func (r *notificationRepository) GetUnreadCount(userID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).Where("ntf_usr_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return int(count), err
}

func (r *notificationRepository) GetFeedUnreadCount(userID string, schoolID string, types []string) (int, error) {
	var count int64
	query := r.db.Model(&domain.Notification{}).Where("ntf_usr_id = ? AND is_read = ?", userID, false)
	if len(types) > 0 {
		query = query.Where("ntf_type IN ?", types)
	}
	query = query.Where(`
		EXISTS (
			SELECT 1
			FROM edv.feeds f
			WHERE f.fds_id = ntf_related_id
				AND f.fds_sch_id = ?
				AND f.deleted_at IS NULL
		)
	`, schoolID)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *notificationRepository) MarkAsRead(notificationID string, userID string) error {
	result := r.db.Model(&domain.Notification{}).
		Where("ntf_id = ? AND ntf_usr_id = ?", notificationID, userID).
		Update("is_read", true)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	return r.db.Model(&domain.Notification{}).Where("ntf_usr_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

func (r *notificationRepository) MarkFeedNotificationsRead(userID string, schoolID string, types []string) error {
	query := r.db.Model(&domain.Notification{}).Where("ntf_usr_id = ? AND is_read = ?", userID, false)
	if len(types) > 0 {
		query = query.Where("ntf_type IN ?", types)
	}
	query = query.Where(`
		EXISTS (
			SELECT 1
			FROM edv.feeds f
			WHERE f.fds_id = ntf_related_id
				AND f.fds_sch_id = ?
				AND f.deleted_at IS NULL
		)
	`, schoolID)
	return query.Update("is_read", true).Error
}

func (r *notificationRepository) Delete(notificationID string, userID string) error {
	result := r.db.Where("ntf_id = ? AND ntf_usr_id = ?", notificationID, userID).
		Delete(&domain.Notification{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
