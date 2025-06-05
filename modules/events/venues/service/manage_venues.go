package service

import (
	"context"
	"errors"

	"ticket-zetu-api/modules/events/models/events"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *venueService) UpdateVenue(userID, id, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64, status string, imageURLs []string) (*events.Venue, error) {
	hasPerm, err := s.HasPermission(userID, "update:venues")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:venues permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	venue.Name = name
	venue.Description = description
	venue.Address = address
	venue.City = city
	venue.State = state
	venue.Country = country
	venue.Capacity = capacity
	venue.ContactInfo = contactInfo
	venue.Latitude = latitude
	venue.Longitude = longitude
	venue.Status = status
	venue.Version++
	venue.UpdatedAt = time.Now()

	if err := s.db.Save(&venue).Error; err != nil {
		return nil, err
	}

	if len(imageURLs) > 0 {
		var existingImages []events.VenueImage
		if err := s.db.Where("venue_id = ? AND deleted_at IS NULL", venue.ID).Find(&existingImages).Error; err != nil {
			return nil, err
		}
		for _, img := range existingImages {
			if err := s.cloudinary.DeleteFile(context.Background(), img.ImageURL); err != nil {
				return nil, err
			}
		}
		if err := s.db.Where("venue_id = ?", venue.ID).Delete(&events.VenueImage{}).Error; err != nil {
			return nil, err
		}
		for i, url := range imageURLs {
			venueImage := events.VenueImage{
				VenueID:   venue.ID,
				ImageURL:  url,
				IsPrimary: i == 0,
				CreatedAt: time.Now(),
			}
			if err := s.db.Create(&venueImage).Error; err != nil {
				return nil, err
			}
		}
	}

	return &venue, nil
}

func (s *venueService) DeleteVenue(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:venues")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:venues permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid venue ID format")
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("venue not found")
		}
		return err
	}

	if venue.Status == "active" {
		return errors.New("cannot delete an active venue")
	}

	var venueImages []events.VenueImage
	if err := s.db.Where("venue_id = ? AND deleted_at IS NULL", venue.ID).Find(&venueImages).Error; err != nil {
		return err
	}
	for _, img := range venueImages {
		if err := s.cloudinary.DeleteFile(context.Background(), img.ImageURL); err != nil {
			return err
		}
	}

	if err := s.db.Delete(&venue).Error; err != nil {
		return err
	}

	return nil
}
