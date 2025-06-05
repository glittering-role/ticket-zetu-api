package service

import (
	"errors"
	"strings"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VenueService interface {
	CreateVenue(userID, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64) (*events.Venue, error)
	UpdateVenue(userID, id, name, description, address, city, state, country string, capacity int, contactInfo string, latitude, longitude float64, status string, imageURLs []string) (*events.Venue, error)
	DeleteVenue(userID, id string) error
	GetVenue(userID, id, fields string) (*venue_dto.VenueResponse, error)
	GetVenues(userID, fields string) ([]venue_dto.VenueResponse, error)
	AddVenueImage(userID, venueID, imageURL string, isPrimary bool) (*events.VenueImage, error)
	DeleteVenueImage(userID, venueID, imageID string) error
	HasPermission(userID, permission string) (bool, error)
	GetAllVenues(fields string) ([]venue_dto.VenueResponse, error)
}

type venueService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	cloudinary           *cloudinary.CloudinaryService
}

func NewVenueService(db *gorm.DB, authService authorization_service.PermissionService, cloudinary *cloudinary.CloudinaryService) VenueService {
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

func (s *venueService) mapVenueToResponse(venue *events.Venue) *venue_dto.VenueResponse {
	return &venue_dto.VenueResponse{
		ID:          venue.ID,
		Name:        venue.Name,
		Description: venue.Description,
		Address:     venue.Address,
		City:        venue.City,
		State:       venue.State,
		Country:     venue.Country,
		Capacity:    venue.Capacity,
		ContactInfo: venue.ContactInfo,
		Latitude:    venue.Latitude,
		Longitude:   venue.Longitude,
		Status:      venue.Status,
		CreatedAt:   venue.CreatedAt,
		VenueImages: venue.VenueImages,
	}
}

// Valid fields for the venues table
var validVenueFields = map[string]bool{
	"id":           true,
	"name":         true,
	"description":  true,
	"address":      true,
	"city":         true,
	"state":        true,
	"country":      true,
	"capacity":     true,
	"contact_info": true,
	"latitude":     true,
	"longitude":    true,
	"status":       true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"version":      true,
}

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
