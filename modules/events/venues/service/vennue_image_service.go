package service

import (
	"context"
	"errors"

	"ticket-zetu-api/modules/events/models/events"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *venueService) AddVenueImage(userID, venueID, imageURL string, isPrimary bool) (*events.VenueImage, error) {
	hasPerm, err := s.HasPermission(userID, "create:venue_images")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:venue_images permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(venueID); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", venueID, organizer.ID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	if isPrimary {
		if err := s.db.Model(&events.VenueImage{}).Where("venue_id = ? AND is_primary = ?", venueID, true).Update("is_primary", false).Error; err != nil {
			return nil, err
		}
	}

	venueImage := events.VenueImage{
		VenueID:   venueID,
		ImageURL:  imageURL,
		IsPrimary: isPrimary,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&venueImage).Error; err != nil {
		return nil, err
	}

	return &venueImage, nil
}

func (s *venueService) DeleteVenueImage(userID, venueID, imageID string) error {
	hasPerm, err := s.HasPermission(userID, "delete:venue_images")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:venue_images permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(venueID); err != nil {
		return errors.New("invalid venue ID format")
	}

	if _, err := uuid.Parse(imageID); err != nil {
		return errors.New("invalid image ID format")
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", venueID, organizer.ID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("venue not found")
		}
		return err
	}

	var venueImage events.VenueImage
	if err := s.db.Where("id = ? AND venue_id = ? AND deleted_at IS NULL", imageID, venueID).First(&venueImage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("venue image not found")
		}
		return err
	}

	if err := s.cloudinary.DeleteFile(context.Background(), venueImage.ImageURL); err != nil {
		return err
	}

	if err := s.db.Delete(&venueImage).Error; err != nil {
		return err
	}

	return nil
}
