package service

import (
	"errors"

	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/models/events"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"
	organizers "ticket-zetu-api/modules/organizers/models"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VenueService interface {
	CreateVenue(userID string, dto venue_dto.CreateVenueDto) (*venue_dto.CreateVenueDto, error)
	UpdateVenue(userID, id string, dto venue_dto.UpdateVenueDto) (*venue_dto.UpdateVenueDto, error)
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
