package service

import (
	"errors"
	"fmt"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"gorm.io/gorm"
)

func (s *eventService) CreateEvent(createDto dto.CreateEvent, userID string) (*dto.EventResponse, error) {
	// Check permissions
	// hasPerm, err := s.HasPermission(userID, "create:events")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:events permission")
	// }

	// Generate slug
	slug, err := s.generateSlug(createDto.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to generate slug: %v", err)
	}

	var event *events.Event

	// Start transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get organizer
		organizer, err := s.getUserOrganizer(userID)
		if err != nil {
			return err
		}

		// Check organizer status
		if organizer.Status != "active" {
			return errors.New("organizer is not active")
		}
		if organizer.IsBanned {
			return errors.New("organizer is banned")
		}
		if organizer.IsFlagged {
			return errors.New("organizer is flagged for review")
		}

		// Validate subcategory
		var subcategory categories.Subcategory
		if err := tx.Where("id = ? AND deleted_at IS NULL", createDto.SubcategoryID).First(&subcategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("subcategory not found")
			}
			return err
		}

		// Validate venue
		var venue events.Venue
		if err := tx.Where("id = ? AND deleted_at IS NULL", createDto.VenueID).First(&venue).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("venue not found")
			}
			return err
		}

		// Build event
		event = &events.Event{
			Title:          createDto.Title,
			Slug:           slug,
			Description:    createDto.Description,
			SubcategoryID:  createDto.SubcategoryID,
			VenueID:        createDto.VenueID,
			TotalSeats:     createDto.TotalSeats,
			AvailableSeats: createDto.TotalSeats,
			StartTime:      createDto.StartTime,
			EndTime:        createDto.EndTime,
			Timezone:       createDto.Timezone,
			Language:       createDto.Language,
			EventType:      events.EventType(createDto.EventType),
			MinAge:         createDto.MinAge,
			IsFree:         createDto.IsFree,
			HasTickets:     createDto.HasTickets,
			IsFeatured:     createDto.IsFeatured,
			Status:         "draft",
			OrganizerID:    organizer.ID,
			Tags:           createDto.Tags,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Create the event
		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.toDto(event)
}
