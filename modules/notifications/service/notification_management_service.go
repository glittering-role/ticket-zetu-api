package notification_service

import (
	"errors"
	notifications "ticket-zetu-api/modules/notifications/models"
	"time"

	"github.com/google/uuid"
)

func (s *notificationService) DeleteUserNotifications(userID string, beforeDate *time.Time) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid user ID format")
	}

	query := s.db.Model(&notifications.UserNotification{}).
		Where("user_id = ?", userID)

	if beforeDate != nil {
		query = query.Where("created_at < ?", beforeDate)
	}

	return query.Update("deleted_at", time.Now()).Error
}

func (s *notificationService) MarkNotificationsAsRead(userID string, notificationID *string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid user ID format")
	}

	now := time.Now()
	query := s.db.Model(&notifications.UserNotification{}).
		Where("user_id = ? AND status = ?", userID, notifications.NotificationStatusUnread)

	if notificationID != nil {
		if _, err := uuid.Parse(*notificationID); err != nil {
			return errors.New("invalid notification ID format")
		}
		query = query.Where("notification_id = ?", *notificationID)
	}

	return query.Updates(map[string]interface{}{
		"status":     notifications.NotificationStatusRead,
		"read_at":    now,
		"updated_at": now,
	}).Error
}
