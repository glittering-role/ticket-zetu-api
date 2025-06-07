package ticket_type_service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/tickets/ticket_type/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Add to your ticketTypeService implementation
func (s *ticketTypeService) GetAllTicketTypesForOrganization(userID, fields string) ([]dto.TicketTypeResponse, error) {
	_, err := s.HasPermission(userID, "read:ticket_types")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:ticket_types permission")
	// }

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var ticketTypes []tickets.TicketType
	query := s.db.Preload("Event").
		Where("organizer_id = ?", organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validTicketTypeFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.Find(&ticketTypes).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.TicketTypeResponse, 0, len(ticketTypes))
	for _, tt := range ticketTypes {
		resp := s.toDTO(&tt)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}

func (s *ticketTypeService) GetTicketType(userID, id, fields string) (*dto.TicketTypeResponse, error) {
	_, err := s.HasPermission(userID, "read:ticket_types")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:ticket_types permission")
	// }

	// Validate ticket type ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid ticket type ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var ticketType tickets.TicketType
	query := s.db.Preload("Event").
		Where("id = ? AND organizer_id = ?", id, organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validTicketTypeFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket type not found")
		}
		return nil, err
	}

	return s.toDTO(&ticketType), nil
}

func (s *ticketTypeService) GetTicketTypes(userID, eventID, fields string) ([]dto.TicketTypeResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:ticket_types")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:ticket_types permission")
	}

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Verify the event belongs to the organizer
	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	var ticketTypes []tickets.TicketType
	query := s.db.Preload("Event").
		Where("event_id = ?", eventID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validTicketTypeFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.Find(&ticketTypes).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.TicketTypeResponse, 0, len(ticketTypes))
	for _, tt := range ticketTypes {
		resp := s.toDTO(&tt)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}
