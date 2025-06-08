package notification_service

import (
	"encoding/json"
	"errors"
	"time"

	"ticket-zetu-api/modules/notifications/models"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService interface {
	SendNotification(module, title, content string, typeVal notifications.NotificationType, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error
	GetUserNotifications(userID string, unreadOnly bool, module string, limit, offset int) ([]notifications.UserNotification, int64, error)
	TriggerNotification(module, action, title, content string, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error
}

type notificationService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
}

func NewNotificationService(db *gorm.DB, authService authorization_service.PermissionService) NotificationService {
	return &notificationService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *notificationService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *notificationService) SendNotification(module, title, content string, typeVal notifications.NotificationType, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error {
	// Validate inputs
	if module == "" {
		return errors.New("module is required")
	}
	if len(recipientIDs) == 0 {
		return errors.New("at least one recipient is required")
	}
	for _, recipientID := range recipientIDs {
		if _, err := uuid.Parse(recipientID); err != nil {
			return errors.New("invalid recipient ID format")
		}
	}
	if senderID != "" {
		if _, err := uuid.Parse(senderID); err != nil {
			return errors.New("invalid sender ID format")
		}
	}
	if relatedID != "" {
		if _, err := uuid.Parse(relatedID); err != nil {
			return errors.New("invalid related ID format")
		}
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return errors.New("invalid metadata format")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create notification
		notification := notifications.Notification{
			ID:        uuid.New().String(),
			Type:      typeVal,
			Title:     title,
			Content:   content,
			SenderID:  senderID,
			RelatedID: relatedID,
			Module:    module,
			Metadata:  string(metadataJSON),
			IsSystem:  typeVal == notifications.NotificationTypeSystem,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&notification).Error; err != nil {
			return err
		}

		// Create UserNotification for each recipient
		for _, recipientID := range recipientIDs {
			userNotification := notifications.UserNotification{
				ID:             uuid.New().String(),
				UserID:         recipientID,
				NotificationID: notification.ID,
				Status:         notifications.NotificationStatusUnread,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := tx.Create(&userNotification).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *notificationService) TriggerNotification(module, action, title, content string, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error {
	// Add action to metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["action"] = action

	// Use action-specific type
	typeVal := notifications.NotificationTypeAction
	if action == "system_message" {
		typeVal = notifications.NotificationTypeSystem
	} else if action == "warning" {
		typeVal = notifications.NotificationTypeWarning
	}

	return s.SendNotification(module, title, content, typeVal, senderID, relatedID, recipientIDs, metadata)
}
