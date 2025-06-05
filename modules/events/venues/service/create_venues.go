package service

import (
	"ticket-zetu-api/modules/events/models/events"

	"time"
)

func (s *venueService) CreateVenue(userID, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64) (*events.Venue, error) {
	_, err := s.HasPermission(userID, "create:venues")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:venues permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	venue := events.Venue{
		Name:        name,
		Description: description,
		Address:     address,
		City:        city,
		State:       state,
		Country:     country,
		Capacity:    capacity,
		ContactInfo: contactInfo,
		Latitude:    latitude,
		Longitude:   longitude,
		OrganizerID: organizer.ID,
		Status:      "active",
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&venue).Error; err != nil {
		return nil, err
	}

	return &venue, nil
}
