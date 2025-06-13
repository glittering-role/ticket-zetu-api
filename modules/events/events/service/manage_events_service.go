package service

import (
	"context"
	"errors"
	"fmt"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *eventService) UpdateEvent(updateDto dto.UpdateEvent, userID, id string) (*dto.EventResponse, error) {
	hasPerm, err := s.HasPermission(userID, "update:events")
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:events permission")
	}

	// Validate IDs
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer: %w", err)
	}

	var event events.Event
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get existing event
		if err := tx.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("event not found or not owned by organizer")
			}
			return fmt.Errorf("failed to fetch event: %w", err)
		}

		// Update fields if provided in DTO
		if updateDto.Title != nil {
			event.Title = *updateDto.Title
			// Update slug if title has changed
			slug, err := s.generateSlug(*updateDto.Title)
			if err != nil {
				return fmt.Errorf("failed to generate slug: %w", err)
			}
			event.Slug = slug
		}
		if updateDto.Description != nil {
			event.Description = *updateDto.Description
		}
		if updateDto.SubcategoryID != nil {
			if _, err := uuid.Parse(*updateDto.SubcategoryID); err != nil {
				return errors.New("invalid subcategory ID format")
			}
			// Validate subcategory exists
			var subcategory categories.Subcategory
			if err := tx.Where("id = ?", *updateDto.SubcategoryID).First(&subcategory).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("subcategory not found")
				}
				return fmt.Errorf("failed to fetch subcategory: %w", err)
			}
			event.SubcategoryID = *updateDto.SubcategoryID
		}
		if updateDto.VenueID != nil {
			if _, err := uuid.Parse(*updateDto.VenueID); err != nil {
				return errors.New("invalid venue ID format")
			}
			// Validate venue exists
			var venue events.Venue
			if err := tx.Where("id = ?", *updateDto.VenueID).First(&venue).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("venue not found")
				}
				return fmt.Errorf("failed to fetch venue: %w", err)
			}
			event.VenueID = *updateDto.VenueID
		}
		if updateDto.TotalSeats != nil {
			event.TotalSeats = *updateDto.TotalSeats
		}
		if updateDto.StartTime != nil {
			event.StartTime = *updateDto.StartTime
		}
		if updateDto.EndTime != nil {
			event.EndTime = *updateDto.EndTime
		}
		if updateDto.Timezone != nil {
			event.Timezone = *updateDto.Timezone
		}
		if updateDto.Language != nil {
			event.Language = *updateDto.Language
		}
		if updateDto.EventType != nil {
			event.EventType = events.EventType(*updateDto.EventType)
		}
		if updateDto.MinAge != nil {
			event.MinAge = *updateDto.MinAge
		}
		if updateDto.IsFree != nil {
			event.IsFree = *updateDto.IsFree
		}
		if updateDto.HasTickets != nil {
			event.HasTickets = *updateDto.HasTickets
		}
		if updateDto.IsFeatured != nil {
			event.IsFeatured = *updateDto.IsFeatured
		}
		if updateDto.Status != nil {
			event.Status = events.EventStatus(*updateDto.Status)
		}

		event.Version++
		event.UpdatedAt = time.Now()

		if err := tx.Save(&event).Error; err != nil {
			return fmt.Errorf("failed to update event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	dtoResult, err := s.toDto(&event, true)
	if err != nil {
		return nil, err
	}
	return &dtoResult.Full, nil

}

func (s *eventService) DeleteEvent(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:events")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:events permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid event ID format")
	}

	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("event not found")
		}
		return err
	}

	// Replace 'EventStatusActive' with the correct constant or string value for active status
	if event.Status == events.EventStatus("active") {
		return errors.New("cannot delete an active event")
	}

	var eventImages []events.EventImage
	if err := s.db.Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&eventImages).Error; err != nil {
		return err
	}
	for _, img := range eventImages {
		if err := s.cloudinary.DeleteFile(context.Background(), img.ImageURL); err != nil {
			return err
		}
	}

	// Soft delete the event
	event.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := s.db.Save(&event).Error; err != nil {
		return err
	}

	return nil
}
