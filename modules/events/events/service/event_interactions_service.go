package service

import (
	"errors"
	"fmt"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"gorm.io/gorm"
)

// ToggleFavorite handles favorite operations
func (s *eventService) ToggleFavorite(userID, eventID string) error {
	if userID == "" || eventID == "" {
		return errors.New("userID and eventID are required")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var existing events.Favorite
		err := tx.Where("user_id = ? AND event_id = ?", userID, eventID).First(&existing).Error

		if err == nil {
			// Favorite exists, so we're removing it
			if err := tx.Delete(&existing).Error; err != nil {
				return fmt.Errorf("failed to remove favorite: %w", err)
			}
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing favorite: %w", err)
		}

		// Favorite doesn't exist, so we're adding it
		favorite := events.Favorite{
			UserID:    userID,
			EventID:   eventID,
			CreatedAt: time.Now(),
		}

		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf("failed to create favorite: %w", err)
		}

		// Fetch event and organizer details
		var event struct {
			Title       string `json:"title"`
			OrganizerID string `json:"organizer_id"`
		}
		if err := tx.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get event details: %w", err)
		}

		var organizer struct {
			CreatedBy string `json:"created_by"`
			Name      string `json:"name"`
		}
		if err := tx.Table("organizers").Where("id = ?", event.OrganizerID).First(&organizer).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get organizer details: %w", err)
		}

		var user struct {
			Username string `json:"username"`
		}
		if err := tx.Table("user_profiles").Where("id = ?", userID).First(&user).Error; err != nil {
			tx.Rollback()
			return nil
		}

		// Send notifications after successful favorite creation
		metadata := map[string]interface{}{
			"event_id":      eventID,
			"event_title":   event.Title,
			"user_id":       userID,
			"user_username": user.Username,
		}

		// Notify user
		s.sendNotification(
			"favorite",
			"Favorited "+event.Title,
			fmt.Sprintf("You have successfully favorited %s.", event.Title),
			userID,
			eventID,
			[]string{userID},
			metadata,
		)

		// Notify organizer owner
		if organizer.CreatedBy != "" && organizer.CreatedBy != userID {
			metadata["action"] = "new_favorite"
			s.sendNotification(
				"new_favorite",
				"New Favorite on "+event.Title,
				fmt.Sprintf("%s has favorited your event, %s.", user.Username, event.Title),
				userID,
				eventID,
				[]string{organizer.CreatedBy},
				metadata,
			)
		}

		return nil
	})
}

// ToggleUpvote handles upvote operations and returns appropriate response
func (s *eventService) ToggleUpvote(userID, eventID string) (string, error) {
	if userID == "" || eventID == "" {
		return "", errors.New("userID and eventID are required")
	}

	var response string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing events.Vote
		err := tx.Where("user_id = ? AND event_id = ?", userID, eventID).First(&existing).Error

		if err == nil {
			// Vote exists, remove it completely
			if err := tx.Delete(&existing).Error; err != nil {
				return fmt.Errorf("failed to remove vote: %w", err)
			}
			response = "upvote removed"
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing vote: %w", err)
		}

		// No vote exists, create new upvote
		vote := events.Vote{
			UserID:    userID,
			EventID:   eventID,
			Type:      events.VoteTypeUp,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&vote).Error; err != nil {
			return fmt.Errorf("failed to create upvote: %w", err)
		}
		response = "upvote added"

		// Fetch event and organizer details for notification
		var event struct {
			Title       string `json:"title"`
			OrganizerID string `json:"organizer_id"`
		}
		if err := tx.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
			return fmt.Errorf("failed to get event details: %w", err)
		}

		var organizer struct {
			CreatedBy string `json:"created_by"`
			Name      string `json:"name"`
		}
		if err := tx.Table("organizers").Where("id = ?", event.OrganizerID).First(&organizer).Error; err != nil {
			return fmt.Errorf("failed to get organizer details: %w", err)
		}

		var voter struct {
			Username string `json:"username"`
		}
		if err := tx.Table("user_profiles").Where("id = ?", userID).First(&voter).Error; err != nil {
			return fmt.Errorf("failed to get voter details: %w", err)
		}

		// Send notifications asynchronously
		go func() {
			notifyTx := s.db.Begin()
			defer func() {
				if r := recover(); r != nil {
					notifyTx.Rollback()
					fmt.Printf("Recovered from panic in notification goroutine: %v\n", r)
				}
			}()

			metadata := map[string]interface{}{
				"event_id":       eventID,
				"event_title":    event.Title,
				"voter_id":       userID,
				"voter_username": voter.Username,
				"vote_type":      "up",
			}

			// Notify voter
			s.sendNotification(
				"up_vote",
				"Upvoted "+event.Title,
				fmt.Sprintf("You have successfully upvoted %s.", event.Title),
				userID,
				eventID,
				[]string{userID},
				metadata,
			)

			// Notify organizer owner
			if organizer.CreatedBy != "" && organizer.CreatedBy != userID {
				metadata["action"] = "new_up_vote"
				s.sendNotification(
					"new_up_vote",
					"New upvote on "+event.Title,
					fmt.Sprintf("%s has upvoted your event, %s.", voter.Username, event.Title),
					userID,
					eventID,
					[]string{organizer.CreatedBy},
					metadata,
				)
			}

			if err := notifyTx.Commit().Error; err != nil {
				fmt.Printf("Failed to commit notification transaction: %v\n", err)
			}
		}()

		return nil
	})

	return response, err
}

// ToggleDownvote handles downvote operations and returns appropriate response
func (s *eventService) ToggleDownvote(userID, eventID string) (string, error) {
	if userID == "" || eventID == "" {
		return "", errors.New("userID and eventID are required")
	}

	var response string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing events.Vote
		err := tx.Where("user_id = ? AND event_id = ?", userID, eventID).First(&existing).Error

		if err == nil {
			// Vote exists, remove it completely
			if err := tx.Delete(&existing).Error; err != nil {
				return fmt.Errorf("failed to remove vote: %w", err)
			}
			response = "downvote removed"
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing vote: %w", err)
		}

		// No vote exists, create new downvote
		vote := events.Vote{
			UserID:    userID,
			EventID:   eventID,
			Type:      events.VoteTypeDown,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&vote).Error; err != nil {
			return fmt.Errorf("failed to create downvote: %w", err)
		}
		response = "downvote added"

		// Fetch event and organizer details for notification
		var event struct {
			Title       string `json:"title"`
			OrganizerID string `json:"organizer_id"`
		}
		if err := tx.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
			return fmt.Errorf("failed to get event details: %w", err)
		}

		var organizer struct {
			CreatedBy string `json:"created_by"`
			Name      string `json:"name"`
		}
		if err := tx.Table("organizers").Where("id = ?", event.OrganizerID).First(&organizer).Error; err != nil {
			return fmt.Errorf("failed to get organizer details: %w", err)
		}

		var voter struct {
			Username string `json:"username"`
		}
		if err := tx.Table("user_profiles").Where("id = ?", userID).First(&voter).Error; err != nil {
			return fmt.Errorf("failed to get voter details: %w", err)
		}

		// Send notifications asynchronously
		go func() {
			notifyTx := s.db.Begin()
			defer func() {
				if r := recover(); r != nil {
					notifyTx.Rollback()
					fmt.Printf("Recovered from panic in notification goroutine: %v\n", r)
				}
			}()

			metadata := map[string]interface{}{
				"event_id":       eventID,
				"event_title":    event.Title,
				"voter_id":       userID,
				"voter_username": voter.Username,
				"vote_type":      "down",
			}

			// Notify voter
			s.sendNotification(
				"down_vote",
				"Downvoted "+event.Title,
				fmt.Sprintf("You have downvoted %s.", event.Title),
				userID,
				eventID,
				[]string{userID},
				metadata,
			)

			// Notify organizer owner
			if organizer.CreatedBy != "" && organizer.CreatedBy != userID {
				metadata["action"] = "new_down_vote"
				s.sendNotification(
					"new_down_vote",
					"New downvote on "+event.Title,
					fmt.Sprintf("%s has downvoted your event, %s.", voter.Username, event.Title),
					userID,
					eventID,
					[]string{organizer.CreatedBy},
					metadata,
				)
			}

			if err := notifyTx.Commit().Error; err != nil {
				fmt.Printf("Failed to commit notification transaction: %v\n", err)
			}
		}()

		return nil
	})

	return response, err
}

// GetUserFavorites handles fetching user favorites
func (s *eventService) GetUserFavorites(userID string) ([]events.Favorite, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var favorites []events.Favorite
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&favorites).Error; err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}

	return favorites, nil
}

// GetUserComments handles fetching user comments
func (s *eventService) GetUserComments(userID string) ([]events.Comment, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var comments []events.Comment
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL AND parent_id IS NULL", userID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return comments, nil
}
