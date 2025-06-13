package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type EventService interface {
	CreateEvent(createDto dto.CreateEvent, userID string) (*dto.EventResponse, error)
	UpdateEvent(updateDto dto.UpdateEvent, userID, id string) (*dto.EventResponse, error)
	DeleteEvent(userID, id string) error
	GetEvent(userID, id string) (*dto.EventResponse, error)
	GetEvents(userID string) ([]dto.MinimalEventResponse, error)
	AddEventImage(userID, eventID, imageURL string, isPrimary bool) (*events.EventImage, error)
	DeleteEventImage(userID, eventID, imageID string) error
	HasPermission(userID, permission string) (bool, error)
}

type eventService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	cloudinary           *cloudinary.CloudinaryService
}

func NewEventService(db *gorm.DB, authService authorization_service.PermissionService, cloudinary *cloudinary.CloudinaryService) EventService {
	return &eventService{
		db:                   db,
		authorizationService: authService,
		cloudinary:           cloudinary,
	}
}

func (s *eventService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *eventService) getUserOrganizer(userID string) (*organizers.Organizer, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}
	return &organizer, nil
}

// generateSlug creates a unique slug for the event
func (s *eventService) generateSlug(title string) (string, error) {
	// Generate base slug
	baseSlug := slug.Make(title)
	if baseSlug == "" {
		return "", fmt.Errorf("invalid title for slug generation")
	}

	var count int64
	err := s.db.Model(&events.Event{}).
		Where("slug = ? AND deleted_at IS NULL", baseSlug).
		Count(&count).Error
	if err != nil {
		return "", fmt.Errorf("failed to check slug: %v", err)
	}

	if count == 0 {
		return baseSlug, nil
	}

	var existingSlugs []string
	err = s.db.Model(&events.Event{}).
		Where("slug LIKE ? AND deleted_at IS NULL", baseSlug+"-%").
		Pluck("slug", &existingSlugs).Error
	if err != nil {
		return "", fmt.Errorf("failed to fetch existing slugs: %v", err)
	}

	maxSuffix := 0
	for _, existingSlug := range existingSlugs {
		suffix := strings.TrimPrefix(existingSlug, baseSlug+"-")
		if num, err := strconv.Atoi(suffix); err == nil && num > maxSuffix {
			maxSuffix = num
		}
	}

	return fmt.Sprintf("%s-%d", baseSlug, maxSuffix+1), nil
}

// toDto converts an Event model to either EventResponse or MinimalEventResponse based on fullDetails
func (s *eventService) toDto(event *events.Event, fullDetails bool) (*struct {
	Full    dto.EventResponse
	Minimal dto.MinimalEventResponse
}, error) {
	// Load event images before using in minimal and full responses
	var eventImages []events.EventImage
	if err := s.db.Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&eventImages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch event images: %v", err)
	}

	// Common fields for both responses
	minimal := dto.MinimalEventResponse{
		ID:          event.ID,
		Title:       event.Title,
		Slug:        event.Slug,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
		Timezone:    event.Timezone,
		EventType:   string(event.EventType),
		IsFree:      event.IsFree,
		HasTickets:  event.HasTickets,
		IsFeatured:  event.IsFeatured,
		Status:      string(event.Status),
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
		EventImages: eventImages,
	}

	if !fullDetails {
		return &struct {
			Full    dto.EventResponse
			Minimal dto.MinimalEventResponse
		}{Minimal: minimal}, nil
	}

	// Load related data for full response
	var subcategory categories.Subcategory
	if err := s.db.First(&subcategory, "id = ? AND deleted_at IS NULL", event.SubcategoryID).Error; err != nil {
		return nil, fmt.Errorf("subcategory not found: %v", err)
	}

	var venue events.Venue
	if err := s.db.Preload("VenueImages").First(&venue, "id = ? AND deleted_at IS NULL", event.VenueID).Error; err != nil {
		return nil, fmt.Errorf("venue not found: %v", err)
	}

	// Build full response
	full := dto.EventResponse{
		ID:            event.ID,
		Title:         event.Title,
		Slug:          event.Slug,
		Description:   event.Description,
		SubcategoryID: event.SubcategoryID,
		Subcategory: dto.SubcategoryResponse{
			ID:         subcategory.ID,
			Name:       subcategory.Name,
			CategoryID: subcategory.CategoryID,
			IsActive:   subcategory.IsActive,
			ImageURL:   subcategory.ImageURL,
		},
		VenueID: event.VenueID,
		Venue: dto.VenueResponse{
			ID:          venue.ID,
			Name:        venue.Name,
			Address:     venue.Address,
			City:        venue.City,
			Country:     venue.Country,
			Capacity:    venue.Capacity,
			Status:      venue.Status,
			VenueImages: venue.VenueImages,
		},
		StartTime:      event.StartTime,
		EndTime:        event.EndTime,
		Timezone:       event.Timezone,
		Language:       event.Language,
		EventType:      string(event.EventType),
		MinAge:         event.MinAge,
		TotalSeats:     event.TotalSeats,
		AvailableSeats: event.AvailableSeats,
		IsFree:         event.IsFree,
		HasTickets:     event.HasTickets,
		IsFeatured:     event.IsFeatured,
		Status:         string(event.Status),
		EventImages:    eventImages,
		PublishedAt:    event.PublishedAt,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
	}

	return &struct {
		Full    dto.EventResponse
		Minimal dto.MinimalEventResponse
	}{Full: full}, nil
}
