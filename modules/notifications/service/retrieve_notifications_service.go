package notification_service

import (
	"errors"

	"ticket-zetu-api/modules/notifications/models"

	"github.com/google/uuid"
)

func (s *notificationService) GetUserNotifications(userID string, unreadOnly bool, module string, limit, offset int) ([]notifications.UserNotification, int64, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, 0, errors.New("invalid user ID format")
	}

	var userNotifications []notifications.UserNotification
	var totalCount int64

	query := s.db.Where("user_id = ? AND deleted_at IS NULL", userID)
	if unreadOnly {
		query = query.Where("status = ?", notifications.NotificationStatusUnread)
	}
	if module != "" {
		query = query.Joins("JOIN notifications ON user_notifications.notification_id = notifications.id").
			Where("notifications.module = ?", module)
	}

	// Count total
	if err := query.Model(&notifications.UserNotification{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Fetch notifications with preloaded Notification and Sender
	if err := query.
		Preload("Notification.Sender").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userNotifications).Error; err != nil {
		return nil, 0, err
	}

	return userNotifications, totalCount, nil
}
