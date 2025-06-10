package notification_service

import (
	"errors"
	"ticket-zetu-api/modules/notifications/dto"
	"ticket-zetu-api/modules/notifications/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *notificationService) GetUserNotifications(userID string, unreadOnly bool, module string, limit, offset int) ([]dto.UserNotificationResponseDTO, int64, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, 0, errors.New("invalid user ID format")
	}

	var totalCount int64
	query := s.db.Model(&notifications.UserNotification{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if unreadOnly {
		query = query.Where("status = ?", notifications.NotificationStatusUnread)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var userNotifications []notifications.UserNotification
	query = s.db.
		Preload("Notification").
		Preload("Notification.Sender", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, first_name, last_name, avatar_url")
		}).
		Where("user_notifications.user_id = ? AND user_notifications.deleted_at IS NULL", userID)

	if unreadOnly {
		query = query.Where("user_notifications.status = ?", notifications.NotificationStatusUnread)
	}

	if module != "" {
		query = query.Joins("JOIN notifications ON user_notifications.notification_id = notifications.id AND notifications.deleted_at IS NULL").
			Where("notifications.module = ?", module)
	}

	if err := query.
		Order("user_notifications.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userNotifications).Error; err != nil {
		return nil, 0, err
	}

	responseDTOs := make([]dto.UserNotificationResponseDTO, len(userNotifications))
	for i, un := range userNotifications {
		var metadata map[string]interface{}
		if un.Notification.Metadata != nil {
			metadata = un.Notification.Metadata
		}

		var sender *dto.SenderDTO
		if un.Notification.Sender.ID != "" {
			sender = &dto.SenderDTO{
				ID:        un.Notification.Sender.ID,
				Username:  un.Notification.Sender.Username,
				FirstName: un.Notification.Sender.FirstName,
				LastName:  un.Notification.Sender.LastName,
				AvatarURL: un.Notification.Sender.AvatarURL,
			}
		}

		responseDTOs[i] = dto.UserNotificationResponseDTO{
			ID:             un.ID,
			NotificationID: un.NotificationID,
			Notification: dto.NotificationResponseDTO{
				ID:        un.Notification.ID,
				Type:      un.Notification.Type,
				Title:     un.Notification.Title,
				Content:   un.Notification.Content,
				SenderID:  un.Notification.SenderID,
				Sender:    sender,
				RelatedID: un.Notification.RelatedID,
				Module:    un.Notification.Module,
				Metadata:  metadata,
				IsSystem:  un.Notification.IsSystem,
				CreatedAt: un.Notification.CreatedAt,
				UpdatedAt: un.Notification.UpdatedAt,
			},
			Status:    un.Status,
			ReadAt:    un.ReadAt,
			CreatedAt: un.CreatedAt,
			UpdatedAt: un.UpdatedAt,
		}
	}

	return responseDTOs, totalCount, nil
}
