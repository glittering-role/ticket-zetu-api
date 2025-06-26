package service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common fields used across venue queries
var defaultVenueFields = []string{
	"id", "name", "description", "address", "city", "state", "postal_code",
	"country", "capacity", "venue_type", "layout", "accessibility_features",
	"facilities", "contact_info", "timezone", "latitude", "longitude",
	"status", "organizer_id", "created_at",
}

// Common query builder for all venue retrieval operations
func (s *venueService) buildVenueQuery(fields string, conditions ...interface{}) *gorm.DB {
	query := s.db.Preload("VenueImages")

	// Apply conditions if any
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	// Handle field selection
	if fields != "" {
		selectedFields := s.filterValidFields(fields)
		if len(selectedFields) > 0 {
			return query.Select(selectedFields)
		}
	}

	return query.Select(defaultVenueFields)
}

// Filter and validate requested fields
func (s *venueService) filterValidFields(fields string) []string {
	selectedFields := []string{}
	for _, field := range strings.Split(fields, ",") {
		field = strings.TrimSpace(field)
		if validVenueFields[field] &&
			field != "created_at" &&
			field != "updated_at" &&
			field != "deleted_at" &&
			field != "version" {
			selectedFields = append(selectedFields, field)
		}
	}
	return selectedFields
}

// Common function to convert venues to DTOs
func (s *venueService) venuesToDTOs(venues []events.Venue) []venue_dto.VenueResponse {
	responses := make([]venue_dto.VenueResponse, len(venues))
	for i, venue := range venues {
		responses[i] = *s.mapVenueToResponse(&venue)
	}
	return responses
}

func (s *venueService) GetVenue(userID, id, fields string) (*venue_dto.VenueResponse, error) {
	// if _, err := s.HasPermission(userID, "read:venues"); err != nil {
	// 	return nil, err
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	var venue events.Venue
	query := s.buildVenueQuery(fields, "id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID)
	if err := query.First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	return s.mapVenueToResponse(&venue), nil
}

func (s *venueService) GetVenues(userID, fields string) ([]venue_dto.VenueResponse, error) {
	// if _, err := s.HasPermission(userID, "read:venues"); err != nil {
	// 	return nil, err
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var venues []events.Venue
	query := s.buildVenueQuery(fields, "organizer_id = ? AND deleted_at IS NULL", organizer.ID)
	if err := query.Find(&venues).Error; err != nil {
		return nil, err
	}

	return s.venuesToDTOs(venues), nil
}

func (s *venueService) GetAllVenues(fields string) ([]venue_dto.VenueResponse, error) {
	var venues []events.Venue
	query := s.buildVenueQuery(fields, "deleted_at IS NULL")
	if err := query.Find(&venues).Error; err != nil {
		return nil, err
	}

	return s.venuesToDTOs(venues), nil
}
