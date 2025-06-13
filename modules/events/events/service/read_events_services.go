package service

import (
	"errors"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *eventService) GetEvent(userID, id string) (*dto.EventResponse, error) {
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

	if err := query.First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	dtoResult, err := s.toDto(&event, true)
	if err != nil {
		return nil, err
	}
	return &dtoResult.Full, nil
}

func (s *eventService) GetEvents(userID string) ([]dto.MinimalEventResponse, error) {
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
	query := s.db.Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.MinimalEventResponse, len(events))
	for i, event := range events {
		response, err := s.toDto(&event, false)
		if err != nil {
			return nil, err
		}
		responses[i] = response.Minimal
	}

	return responses, nil
}
