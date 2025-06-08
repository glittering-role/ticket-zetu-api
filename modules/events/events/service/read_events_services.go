package service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *eventService) GetEvent(userID, id, fields string) (*dto.EventResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "read:events")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:events permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	var event events.Event
	query := s.db.Preload("Venue").Preload("EventImages").
		Preload("Subcategory.Category").
		Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validEventFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select(defaultEventFields)
		}
	} else {
		query = query.Select(defaultEventFields)
	}

	if err := query.First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return s.toDto(&event)
}

func (s *eventService) GetEvents(userID, fields string) ([]dto.EventResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "read:events")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:events permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var events []events.Event
	query := s.db.Preload("Venue").Preload("EventImages").
		Preload("Subcategory.Category").
		Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validEventFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select(defaultEventFields)
		}
	} else {
		query = query.Select(defaultEventFields)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.EventResponse, len(events))
	for i, event := range events {
		dto, err := s.toDto(&event)
		if err != nil {
			return nil, err
		}
		responses[i] = *dto
	}

	return responses, nil
}

var defaultEventFields = []string{
	"id", "title", "subcategory_id", "description", "venue_id",
	"total_seats", "available_seats", "start_time", "end_time",
	"price_tier_id", "base_price", "is_featured", "status",
}
