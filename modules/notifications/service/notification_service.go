package notification_service

import (
	"errors"
	"sync"
	"ticket-zetu-api/modules/notifications/dto"
	notifications "ticket-zetu-api/modules/notifications/models"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService interface {
	SendNotification(dto *dto.CreateNotificationDTO) error
	GetUserNotifications(userID string, unreadOnly bool, module string, limit, offset int) ([]dto.UserNotificationResponseDTO, int64, error)
	TriggerNotification(module, action, title, content string, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error
	DeleteUserNotifications(userID string, beforeDate *time.Time) error
	MarkNotificationsAsRead(userID string, notificationID *string) error
	CountUnreadNotifications(userID string) (int64, error)
}

type notificationService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	validator            *validator.Validate
	mu                   sync.Mutex
}

func NewNotificationService(db *gorm.DB, authService authorization_service.PermissionService) NotificationService {
	return &notificationService{
		db:                   db,
		authorizationService: authService,
		validator:            validator.New(),
	}
}

func (s *notificationService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	return s.authorizationService.HasPermission(userID, permission)
}

func (s *notificationService) SendNotification(dto *dto.CreateNotificationDTO) error {
	if err := s.validator.Struct(dto); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		notification := notifications.Notification{
			ID:        uuid.New().String(),
			Type:      dto.Type,
			Title:     dto.Title,
			Content:   dto.Content,
			SenderID:  dto.SenderID,
			RelatedID: dto.RelatedID,
			Module:    dto.Module,
			Metadata:  dto.Metadata,
			IsSystem:  dto.Type == notifications.NotificationTypeSystem,
		}

		if err := tx.Create(&notification).Error; err != nil {
			return err
		}

		return s.batchCreateUserNotifications(tx, notification.ID, dto.RecipientIDs)
	})
}

func (s *notificationService) batchCreateUserNotifications(tx *gorm.DB, notificationID string, recipientIDs []string) error {
	const batchSize = 100
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	semaphore := make(chan struct{}, 10) // Limit concurrent goroutines

	for i := 0; i < len(recipientIDs); i += batchSize {
		end := i + batchSize
		if end > len(recipientIDs) {
			end = len(recipientIDs)
		}
		batch := recipientIDs[i:end]

		wg.Add(1)
		go func(recipients []string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			var userNotifications []notifications.UserNotification
			for _, recipientID := range recipients {
				if _, err := uuid.Parse(recipientID); err != nil {
					errChan <- errors.New("invalid recipient ID: " + recipientID)
					return
				}

				userNotifications = append(userNotifications, notifications.UserNotification{
					ID:             uuid.New().String(),
					UserID:         recipientID,
					NotificationID: notificationID,
					Status:         notifications.NotificationStatusUnread,
				})
			}

			if err := tx.Create(&userNotifications).Error; err != nil {
				errChan <- err
			}
		}(batch)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *notificationService) TriggerNotification(module, action, title, content string, senderID, relatedID string, recipientIDs []string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["action"] = action

	notificationType := notifications.NotificationTypeAction
	switch action {
	case "system_message":
		notificationType = notifications.NotificationTypeSystem
	case "warning":
		notificationType = notifications.NotificationTypeWarning
	}

	return s.SendNotification(&dto.CreateNotificationDTO{
		Module:       module,
		Title:        title,
		Content:      content,
		Type:         notificationType,
		SenderID:     senderID,
		RelatedID:    relatedID,
		RecipientIDs: recipientIDs,
		Metadata:     metadata,
	})
}

func (s *notificationService) CountUnreadNotifications(userID string) (int64, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return 0, errors.New("invalid user ID format")
	}

	var count int64
	err := s.db.Model(&notifications.UserNotification{}).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, notifications.NotificationStatusUnread).
		Count(&count).Error

	return count, err
}
