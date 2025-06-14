package service

import (
	"errors"
	"fmt"
	"strings"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *eventService) GetEvent(userID, id string) (*dto.EventResponse, error) {
	// Validate permissions
	// hasPerm, err := s.HasPermission(userID, "read:events")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to check permissions: %v", err)
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:events permission")
	// }

	// Validate event ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	// Get organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Fetch event with necessary associations
	var event events.Event
	query := s.db.
		Preload("Venue.VenueImages", "deleted_at IS NULL").
		Preload("Subcategory", "deleted_at IS NULL").
		Preload("EventImages", "deleted_at IS NULL").
		Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID)

	if err := query.First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, fmt.Errorf("failed to fetch event: %v", err)
	}

	// Convert to DTO
	dtoResult, err := s.toDto(&event, true)
	if err != nil {
		return nil, err
	}
	return &dtoResult.Full, nil
}

func (s *eventService) GetEvents(userID string, filter SearchFilter) (*PaginatedResponse, error) {
	// Validate permissions
	// hasPerm, err := s.HasPermission(userID, "read:events")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to check permissions: %v", err)
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:events permission")
	// }

	// Get organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Validate pagination parameters
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20 // Default to 20, cap at 100 to prevent abuse
	}

	// Fetch events with pagination
	var eventList []events.Event
	var totalItems int64

	// Build query
	query := s.db.Model(&events.Event{}).
		Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)

	// Count total items
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count events: %v", err)
	}

	// Calculate total pages
	totalPages := int((totalItems + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	// Fetch paginated events
	if err := query.
		Order("created_at DESC"). // Order by creation date
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&eventList).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events: %v", err)
	}

	// Convert to DTO
	responses := make([]dto.MinimalEventResponse, len(eventList))
	for i, event := range eventList {
		response, err := s.toDto(&event, false)
		if err != nil {
			return nil, err
		}
		responses[i] = response.Minimal
	}

	return &PaginatedResponse{
		Events:      responses,
		TotalItems:  totalItems,
		CurrentPage: filter.Page,
		TotalPages:  totalPages,
	}, nil
}

func (s *eventService) SearchEvents(userID string, filter SearchFilter) (*PaginatedResponse, error) {
	// Validate permissions
	// hasPerm, err := s.HasPermission(userID, "read:events")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to check permissions: %v", err)
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:events permission")
	// }

	// Get organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Validate pagination parameters
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20 // Default to 20, cap at 100 to prevent abuse
	}

	// Build query
	query := s.db.Model(&events.Event{}).
		Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)

	// Apply search query
	if filter.Query != "" {
		searchTerm := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Apply filters
	if filter.StartDate != nil {
		query = query.Where("start_time >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("end_time <= ?", *filter.EndDate)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.IsFree != nil {
		query = query.Where("is_free = ?", *filter.IsFree)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Apply price range filter with LEFT JOIN to include events without tickets
	if filter.MinPrice != nil || filter.MaxPrice != nil {
		query = query.Joins("LEFT JOIN ticket_types ON ticket_types.event_id = events.id AND ticket_types.deleted_at IS NULL").
			Joins("LEFT JOIN ticket_type_price_tiers ON ticket_type_price_tiers.ticket_type_id = ticket_types.id").
			Joins("LEFT JOIN price_tiers ON price_tiers.id = ticket_type_price_tiers.price_tier_id AND price_tiers.deleted_at IS NULL")
		if filter.MinPrice != nil {
			query = query.Where("price_tiers.base_price >= ? OR price_tiers.base_price IS NULL", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			query = query.Where("price_tiers.base_price <= ? OR price_tiers.base_price IS NULL", *filter.MaxPrice)
		}
	}

	// Count total items
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count events: %v", err)
	}

	// Calculate total pages
	totalPages := int((totalItems + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	// Fetch paginated events
	var events []events.Event
	if err := query.
		Order("created_at DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events: %v", err)
	}

	// Convert to DTO
	responses := make([]dto.MinimalEventResponse, len(events))
	for i, event := range events {
		response, err := s.toDto(&event, false)
		if err != nil {
			return nil, err
		}
		responses[i] = response.Minimal
	}

	return &PaginatedResponse{
		Events:      responses,
		TotalItems:  totalItems,
		CurrentPage: filter.Page,
		TotalPages:  totalPages,
	}, nil
}
