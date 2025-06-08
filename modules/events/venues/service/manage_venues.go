package service

import (
	"context"
	"errors"
	"time"

	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *venueService) UpdateVenue(userID, id string, dto venue_dto.UpdateVenueDto) (*venue_dto.UpdateVenueDto, error) {
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

	// Update venue fields from DTO
	venue.Name = dto.Name
	venue.Description = dto.Description
	venue.Address = dto.Address
	venue.City = dto.City
	venue.State = dto.State
	venue.Country = dto.Country
	venue.Capacity = dto.Capacity
	venue.ContactInfo = dto.ContactInfo
	venue.Latitude = dto.Latitude
	venue.Longitude = dto.Longitude
	venue.Status = dto.Status
	venue.Version++
	venue.UpdatedAt = time.Now()

	if err := s.db.Save(&venue).Error; err != nil {
		return nil, err
	}

	// Return the DTO with updated values
	return &dto, nil
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

	// Soft delete the venue
	venue.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := s.db.Save(&venue).Error; err != nil {
		return err
	}

	return nil
}
