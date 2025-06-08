package service

import (
	"context"
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *eventService) AddEventImage(userID, eventID, imageURL string, isPrimary bool) (*events.EventImage, error) {
	// Check permissions
	hasPerm, err := s.HasPermission(userID, "create:event_images")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:event_images permission")
	}

	// Get organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	var eventImage *events.EventImage

	// Start transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Confirm event ownership
		var event events.Event
		if err := tx.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", eventID, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("event not found")
			}
			return err
		}

		// Check image count
		var imageCount int64
		if err := tx.Model(&events.EventImage{}).Where("event_id = ? AND deleted_at IS NULL", eventID).Count(&imageCount).Error; err != nil {
			return err
		}
		if imageCount >= 5 {
			return errors.New("maximum 5 images allowed per event")
		}

		// Unset current primary image if applicable
		if isPrimary {
			if err := tx.Model(&events.EventImage{}).Where("event_id = ? AND is_primary = true AND deleted_at IS NULL", eventID).Update("is_primary", false).Error; err != nil {
				return err
			}
		}

		// Create new event image
		eventImage = &events.EventImage{
			EventID:   eventID,
			ImageURL:  imageURL,
			IsPrimary: isPrimary,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		}

		if err := tx.Create(eventImage).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return eventImage, nil
}

func (s *eventService) DeleteEventImage(userID, eventID, imageID string) error {
	// Check permissions
	hasPerm, err := s.HasPermission(userID, "delete:event_images")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:event_images permission")
	}

	// Get organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	// Validate IDs
	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event ID format")
	}
	if _, err := uuid.Parse(imageID); err != nil {
		return errors.New("invalid image ID format")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {

		var event events.Event
		if err := tx.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", eventID, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("event not found")
			}
			return err
		}

		var eventImage events.EventImage
		if err := tx.Where("id = ? AND event_id = ? AND deleted_at IS NULL", imageID, eventID).First(&eventImage).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("event image not found")
			}
			return err
		}

		// Delete from Cloudinary
		if err := s.cloudinary.DeleteFile(context.Background(), eventImage.ImageURL); err != nil {
			return err
		}

		// Soft delete from database
		if err := tx.Delete(&eventImage).Error; err != nil {
			return err
		}

		return nil
	})
}
