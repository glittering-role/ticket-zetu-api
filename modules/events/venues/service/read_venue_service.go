package service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *venueService) GetVenue(userID, id, fields string) (*venue_dto.VenueResponse, error) {
	_, err := s.HasPermission(userID, "read:venues")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:venues permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	var venue events.Venue
	query := s.db.Preload("VenueImages").Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID)
	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validVenueFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
		}
	} else {
		query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
	}
	if err := query.First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	return s.mapVenueToResponse(&venue), nil
}

func (s *venueService) GetVenues(userID, fields string) ([]venue_dto.VenueResponse, error) {
	_, err := s.HasPermission(userID, "read:venues")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:venues permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var venues []events.Venue
	query := s.db.Preload("VenueImages").Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)
	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validVenueFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
		}
	} else {
		query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
	}
	if err := query.Find(&venues).Error; err != nil {
		return nil, err
	}

	responses := make([]venue_dto.VenueResponse, len(venues))
	for i, venue := range venues {
		responses[i] = *s.mapVenueToResponse(&venue)
	}

	return responses, nil
}

func (s *venueService) GetAllVenues(fields string) ([]venue_dto.VenueResponse, error) {
	var venues []events.Venue
	query := s.db.Preload("VenueImages").Where("deleted_at IS NULL")
	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validVenueFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
		}
	} else {
		query = query.Select("id", "name", "description", "address", "city", "state", "country", "capacity", "contact_info", "latitude", "longitude", "status", "created_at")
	}
	if err := query.Find(&venues).Error; err != nil {
		return nil, err
	}

	responses := make([]venue_dto.VenueResponse, len(venues))
	for i, venue := range venues {
		responses[i] = *s.mapVenueToResponse(&venue)
	}

	return responses, nil
}
