package service

import (
	"context"
	"errors"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VenueService interface {
	CreateVenue(userID, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64) (*events.Venue, error)
	UpdateVenue(userID, id, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64, status string, imageURLs []string) (*events.Venue, error)
	DeleteVenue(userID, id string) error
	GetVenue(userID, id string) (*events.Venue, error)
	GetVenues(userID string) ([]events.Venue, error)
	AddVenueImage(userID, venueID, imageURL string, isPrimary bool) (*events.VenueImage, error)
	DeleteVenueImage(userID, venueID, imageID string) error
	HasPermission(userID, permission string) (bool, error)
}

type venueService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
	cloudinary           *cloudinary.CloudinaryService
}

func NewVenueService(db *gorm.DB, authService authorization.PermissionService, cloudinary *cloudinary.CloudinaryService) VenueService {
	return &venueService{
		db:                   db,
		authorizationService: authService,
		cloudinary:           cloudinary,
	}
}

func (s *venueService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *venueService) getUserOrganizer(userID string) (*organizers.Organizer, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}
	return &organizer, nil
}

func (s *venueService) CreateVenue(userID, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64) (*events.Venue, error) {
	_, err := s.HasPermission(userID, "create:venues")
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
	}

	if err := s.db.Create(&venue).Error; err != nil {
		return nil, err
	}

	return &venue, nil
}

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

	if err := s.db.Save(&venue).Error; err != nil {
		return nil, err
	}

	// Update venue images: delete existing and add new ones if provided
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

	// Delete associated images from Cloudinary
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

func (s *venueService) GetVenue(userID, id string) (*events.Venue, error) {
	hasPerm, err := s.HasPermission(userID, "read:venues")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:venues permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	var venue events.Venue
	if err := s.db.Preload("VenueImages").Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	return &venue, nil
}

func (s *venueService) GetVenues(userID string) ([]events.Venue, error) {
	hasPerm, err := s.HasPermission(userID, "read:venues")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:venues permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var venues []events.Venue
	if err := s.db.Preload("VenueImages").Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID).Find(&venues).Error; err != nil {
		return nil, err
	}

	return venues, nil
}

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

	// If setting as primary, unset other primary images
	if isPrimary {
		if err := s.db.Model(&events.VenueImage{}).Where("venue_id = ? AND is_primary = ?", venueID, true).Update("is_primary", false).Error; err != nil {
			return nil, err
		}
	}

	venueImage := events.VenueImage{
		VenueID:   venueID,
		ImageURL:  imageURL,
		IsPrimary: isPrimary,
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
