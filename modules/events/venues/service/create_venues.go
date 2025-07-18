package service

import (
	//"errors"
	"encoding/json"
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"
)

func (s *venueService) CreateVenue(userID string, dto venue_dto.CreateVenueDto) (*venue_dto.CreateVenueDto, error) {
	// Check permissions
	// hasPerm, err := s.HasPermission(userID, "create:venues")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:venues permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Validate JSON fields
	if dto.Layout != "" {
		if !json.Valid([]byte(dto.Layout)) {
			return nil, errors.New("layout must be valid JSON")
		}
	}
	if dto.AccessibilityFeatures != "" {
		if !json.Valid([]byte(dto.AccessibilityFeatures)) {
			return nil, errors.New("accessibility_features must be valid JSON")
		}
	}
	if dto.Facilities != "" {
		if !json.Valid([]byte(dto.Facilities)) {
			return nil, errors.New("facilities must be valid JSON")
		}
	}

	// Validate and normalize JSON fields
	if dto.Layout != "" {
		if !json.Valid([]byte(dto.Layout)) {
			return nil, errors.New("layout must be valid JSON")
		}
	} else {
		dto.Layout = "{}"
	}

	if dto.AccessibilityFeatures != "" {
		if !json.Valid([]byte(dto.AccessibilityFeatures)) {
			return nil, errors.New("accessibility_features must be valid JSON")
		}
	} else {
		dto.AccessibilityFeatures = "[]"
	}

	if dto.Facilities != "" {
		if !json.Valid([]byte(dto.Facilities)) {
			return nil, errors.New("facilities must be valid JSON")
		}
	} else {
		dto.Facilities = "[]"
	}

	// Create the venue
	venue := events.Venue{
		OrganizerID:           organizer.ID,
		Name:                  dto.Name,
		Description:           dto.Description,
		Address:               dto.Address,
		City:                  dto.City,
		State:                 dto.State,
		PostalCode:            dto.PostalCode,
		Country:               dto.Country,
		Capacity:              dto.Capacity,
		VenueType:             events.VenueType(dto.VenueType),
		Layout:                dto.Layout,
		AccessibilityFeatures: dto.AccessibilityFeatures,
		Facilities:            dto.Facilities,
		ContactInfo:           dto.ContactInfo,
		Timezone:              dto.Timezone,
		Latitude:              dto.Latitude,
		Longitude:             dto.Longitude,
		Status:                events.VenueStatus(dto.Status),
	}

	if err := s.db.Create(&venue).Error; err != nil {
		return nil, err
	}

	return &dto, nil
}
