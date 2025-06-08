package organizers_services

import (
	"errors"
	"fmt"

	//"fmt"
	"time"

	notification_service "ticket-zetu-api/modules/notifications/service"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers "ticket-zetu-api/modules/organizers/models"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubscriptionService interface {
	SubscribeToOrganization(UserID, organizerID string) (*organizer_dto.SubscriptionInfo, error)
	UnsubscribeFromOrganization(UserID, organizerID string) error
	UpdateSubscriptionPreferences(UserID, organizerID string, preferences organizer_dto.SubscriptionInfo) (*organizer_dto.SubscriptionInfo, error)
	BanSubscriber(UserID, subscriberID string, reason string) error
	GetSubscriptionsForOrganizer(UserID string, page, pageSize int) ([]organizer_dto.SubscriberInfo, int64, error)
	GetSubscriptionsForUser(UserID string, page, pageSize int) ([]organizer_dto.OrganizerSubscriptionInfo, int64, error)
}

type subscriptionService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	notificationService  notification_service.NotificationService
}

func NewSubscriptionService(
	db *gorm.DB,
	authService authorization_service.PermissionService,
	notificationService notification_service.NotificationService,
) SubscriptionService {
	return &subscriptionService{
		db:                   db,
		authorizationService: authService,
		notificationService:  notificationService,
	}
}

// Helper functions
func (s *subscriptionService) validateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid ID format")
	}
	return nil
}

func (s *subscriptionService) checkPermission(userID, permission string) error {
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil || !hasPerm {
		return errors.New("user lacks required permission")
	}
	return nil
}

func (s *subscriptionService) getActiveSubscription(organizerID, userID string) (*organizers.OrganizationSubscription, error) {
	var subscription organizers.OrganizationSubscription
	err := s.db.Where("organizer_id = ? AND subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL",
		organizerID, userID, true).First(&subscription).Error
	if err != nil {
		return nil, errors.New("subscription not found")
	}
	return &subscription, nil
}

func (s *subscriptionService) getOrganizerIDForUser(userID string) (string, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("no organizer found for this user")
		}
		return "", err
	}
	return organizer.ID, nil
}

func (s *subscriptionService) toSubscriptionInfo(subscription *organizers.OrganizationSubscription) *organizer_dto.SubscriptionInfo {
	if subscription == nil {
		return &organizer_dto.SubscriptionInfo{IsSubscribed: false}
	}
	return &organizer_dto.SubscriptionInfo{
		IsSubscribed:           subscription.IsActive && !subscription.UnsubscribedAt.Valid,
		SubscriptionDate:       subscription.SubscriptionDate,
		ReceiveEventUpdates:    subscription.ReceiveEventUpdates,
		ReceiveNewsletters:     subscription.ReceiveNewsletters,
		ReceivePromotions:      subscription.ReceivePromotions,
		NotificationPreference: subscription.NotificationTypes,
	}
}

func (s *subscriptionService) validatePaginationParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

// Main service methods
func (s *subscriptionService) SubscribeToOrganization(UserID, organizerID string) (*organizer_dto.SubscriptionInfo, error) {
	// Validate UUIDs first
	if err := s.validateUUID(UserID); err != nil {
		return nil, errors.New("invalid user ID format")
	}
	if err := s.validateUUID(organizerID); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	var result *organizer_dto.SubscriptionInfo
	var organizer organizers.Organizer
	var subscriber members.User

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Check organizer exists and accepts subscriptions with lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("CreatedByUser").
			Where("id = ? AND is_accepting_subscriptions = ? AND deleted_at IS NULL", organizerID, true).
			First(&organizer).Error; err != nil {
			return errors.New("organizer not found or not accepting subscriptions")
		}

		// Fetch subscriber
		if err := tx.Where("id = ? AND deleted_at IS NULL", UserID).First(&subscriber).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("subscriber not found")
			}
			return err
		}

		// Try to create new subscription
		subscription := organizers.OrganizationSubscription{
			OrganizerID:         organizerID,
			SubscriberID:        UserID,
			ReceiveEventUpdates: true,
			ReceiveNewsletters:  true,
			ReceivePromotions:   false,
			NotificationTypes:   "all",
			IsActive:            true,
			SubscriptionDate:    time.Now(),
		}

		createErr := tx.Create(&subscription).Error
		if createErr == nil {
			// Update subscriber count
			if err := tx.Model(&organizer).
				Update("subscriber_count", gorm.Expr("subscriber_count + 1")).Error; err != nil {
				return err
			}
			result = s.toSubscriptionInfo(&subscription)
			return nil
		}

		// If create failed due to duplicate, handle existing subscription
		var existingSubscription organizers.OrganizationSubscription
		if err := tx.Where("organizer_id = ? AND subscriber_id = ?", organizerID, UserID).
			First(&existingSubscription).Error; err != nil {
			return createErr
		}

		if !existingSubscription.UnsubscribedAt.Valid && existingSubscription.IsActive {
			return errors.New("already subscribed to this organizer")
		}

		// Reactivate existing subscription
		updates := map[string]interface{}{
			"unsubscribed_at":       gorm.DeletedAt{},
			"is_active":             true,
			"receive_event_updates": true,
			"receive_newsletters":   true,
			"receive_promotions":    false,
			"notification_types":    "all",
			"last_updated":          time.Now(),
		}

		if err := tx.Model(&existingSubscription).Updates(updates).Error; err != nil {
			return err
		}

		// Only increment count if this was a reactivation
		if existingSubscription.UnsubscribedAt.Valid {
			if err := tx.Model(&organizer).
				Update("subscriber_count", gorm.Expr("subscriber_count + 1")).Error; err != nil {
				return err
			}
		}

		result = s.toSubscriptionInfo(&existingSubscription)
		return nil
	})

	if err == nil && result != nil {
		// Send notifications after successful subscription
		metadata := map[string]interface{}{
			"organizer_id":        organizerID,
			"organizer_name":      organizer.Name,
			"subscriber_username": subscriber.Username,
		}

		// Notify subscriber
		notifyErr := s.notificationService.TriggerNotification(
			"organizers",
			"subscription",
			"Subscribed to "+organizer.Name,
			fmt.Sprintf("You have successfully subscribed to %s. You'll receive updates based on your preferences.", organizer.Name),
			UserID,
			organizerID,
			[]string{UserID},
			metadata,
		)
		if notifyErr != nil {
			// log.Printf("Failed to send subscriber notification: %v", notifyErr)
		}

		// Notify organizer owner
		if organizer.CreatedBy != "" {
			metadata["action"] = "new_subscriber"
			notifyErr = s.notificationService.TriggerNotification(
				"organizers",
				"new_subscriber",
				"New Subscriber to "+organizer.Name,
				fmt.Sprintf("%s has subscribed to your organization, %s.", subscriber.Username, organizer.Name),
				UserID,
				organizerID,
				[]string{organizer.CreatedBy},
				metadata,
			)
			if notifyErr != nil {
				// log.Printf("Failed to send owner notification: %v", notifyErr)
			}
		}
	}

	return result, err
}

func (s *subscriptionService) UnsubscribeFromOrganization(UserID, organizerID string) error {
	// Validate UUIDs first
	if err := s.validateUUID(UserID); err != nil {
		return errors.New("invalid user ID format")
	}
	if err := s.validateUUID(organizerID); err != nil {
		return errors.New("invalid organizer ID format")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// First check if organizer exists (no lock needed for read)
		var organizer organizers.Organizer
		if err := tx.Where("id = ? AND deleted_at IS NULL", organizerID).
			First(&organizer).Error; err != nil {
			return errors.New("organizer not found")
		}

		// Update subscription if active
		result := tx.Model(&organizers.OrganizationSubscription{}).
			Where("organizer_id = ? AND subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL",
				organizerID, UserID, true).
			Updates(map[string]interface{}{
				"unsubscribed_at": gorm.DeletedAt{Time: time.Now(), Valid: true},
				"is_active":       false,
				"last_updated":    time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("active subscription not found")
		}

		// Update subscriber count
		if err := tx.Model(&organizer).
			Update("subscriber_count", gorm.Expr("subscriber_count - 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *subscriptionService) UpdateSubscriptionPreferences(UserID, organizerID string, preferences organizer_dto.SubscriptionInfo) (*organizer_dto.SubscriptionInfo, error) {
	if err := s.validateUUID(UserID); err != nil {
		return nil, errors.New("invalid user ID format")
	}
	if err := s.validateUUID(organizerID); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	if err := s.checkPermission(UserID, "update_subscription_preferences"); err != nil {
		return nil, err
	}

	subscription, err := s.getActiveSubscription(organizerID, UserID)
	if err != nil {
		return nil, err
	}

	subscription.ReceiveEventUpdates = preferences.ReceiveEventUpdates
	subscription.ReceiveNewsletters = preferences.ReceiveNewsletters
	subscription.ReceivePromotions = preferences.ReceivePromotions
	subscription.NotificationTypes = preferences.NotificationPreference

	if err := s.db.Save(subscription).Error; err != nil {
		return nil, err
	}

	return s.toSubscriptionInfo(subscription), nil
}

func (s *subscriptionService) BanSubscriber(UserID, subscriberID string, reason string) error {
	if err := s.validateUUID(UserID); err != nil {
		return errors.New("invalid user ID format")
	}
	if err := s.validateUUID(subscriberID); err != nil {
		return errors.New("invalid subscriber ID format")
	}

	if err := s.checkPermission(UserID, "ban_subscriber"); err != nil {
		return err
	}

	organizerID, err := s.getOrganizerIDForUser(UserID)
	if err != nil {
		return err
	}

	var subscription organizers.OrganizationSubscription
	if err := s.db.Where("organizer_id = ? AND subscriber_id = ?", organizerID, subscriberID).First(&subscription).Error; err != nil {
		return errors.New("subscription not found")
	}

	subscription.IsBlocked = true
	subscription.BlockedReason = reason
	subscription.LastUpdated = time.Now()
	subscription.IsActive = false

	if err := s.db.Save(&subscription).Error; err != nil {
		return err
	}

	return nil
}

func (s *subscriptionService) GetSubscriptionsForOrganizer(UserID string, page, pageSize int) ([]organizer_dto.SubscriberInfo, int64, error) {
	if err := s.validateUUID(UserID); err != nil {
		return nil, 0, errors.New("invalid user ID format")
	}

	// if err := s.checkPermission(UserID, "view_subscribers"); err != nil {
	// 	return nil, 0, err
	// }

	organizerID, err := s.getOrganizerIDForUser(UserID)
	if err != nil {
		return nil, 0, err
	}

	page, pageSize = s.validatePaginationParams(page, pageSize)
	offset := (page - 1) * pageSize

	var total int64
	var subscriptions []organizers.OrganizationSubscription

	if err := s.db.Model(&organizers.OrganizationSubscription{}).
		Where("organizer_id = ? AND is_active = ? AND unsubscribed_at IS NULL", organizerID, true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.
		Preload("Subscriber").
		Where("organizer_id = ? AND is_active = ? AND unsubscribed_at IS NULL", organizerID, true).
		Offset(offset).
		Limit(pageSize).
		Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	subscribers := make([]organizer_dto.SubscriberInfo, 0, len(subscriptions))
	for _, sub := range subscriptions {
		if sub.Subscriber.ID != "" {
			subscribers = append(subscribers, organizer_dto.SubscriberInfo{
				UserID:       sub.SubscriberID,
				Email:        sub.Subscriber.Email,
				Username:     sub.Subscriber.Username,
				FirstName:    sub.Subscriber.FirstName,
				LastName:     sub.Subscriber.LastName,
				AvatarURL:    sub.Subscriber.AvatarURL,
				SubscribedAt: sub.SubscriptionDate,
				IsBanned:     sub.IsBlocked,
			})
		}
	}

	return subscribers, total, nil
}

func (s *subscriptionService) GetSubscriptionsForUser(UserID string, page, pageSize int) ([]organizer_dto.OrganizerSubscriptionInfo, int64, error) {
	if err := s.validateUUID(UserID); err != nil {
		return nil, 0, errors.New("invalid user ID format")
	}

	page, pageSize = s.validatePaginationParams(page, pageSize)
	offset := (page - 1) * pageSize

	var total int64
	var subscriptions []organizers.OrganizationSubscription

	if err := s.db.Model(&organizers.OrganizationSubscription{}).
		Where("subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL", UserID, true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.
		Preload("Organizer").
		Where("subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL", UserID, true).
		Offset(offset).
		Limit(pageSize).
		Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	organizerSubs := make([]organizer_dto.OrganizerSubscriptionInfo, 0, len(subscriptions))
	for _, sub := range subscriptions {
		if sub.Organizer.ID != "" {
			organizerSubs = append(organizerSubs, organizer_dto.OrganizerSubscriptionInfo{
				OrganizerID:  sub.OrganizerID,
				Name:         sub.Organizer.Name,
				ImageURL:     sub.Organizer.ImageURL,
				SubscribedAt: sub.SubscriptionDate,
				Preferences:  *s.toSubscriptionInfo(&sub),
			})
		}
	}

	return organizerSubs, total, nil
}
