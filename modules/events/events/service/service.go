package service

import (
	"errors"
	"fmt"
	"strings"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/events/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventService interface {
	CreateEvent(createDto dto.CreateEvent, userID string) (*dto.EventResponse, error)
	UpdateEvent(updateDto dto.UpdateEvent, userID, id string) (*dto.EventResponse, error)
	DeleteEvent(userID, id string) error
	GetEvent(userID, id, fields string) (*dto.EventResponse, error)
	GetEvents(userID, fields string) ([]dto.EventResponse, error)
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
	baseSlug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	baseSlug = strings.ReplaceAll(baseSlug, "'", "")
	baseSlug = strings.ReplaceAll(baseSlug, `"`, "")

	slug := baseSlug
	counter := 1

	for {
		var count int64
		err := s.db.Model(&events.Event{}).Where("slug = ?", slug).Count(&count).Error
		if err != nil {
			return "", err
		}

		if count == 0 {
			break
		}

		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	return slug, nil
}

// toDto converts an Event model to EventResponse DTO
func (s *eventService) toDto(event *events.Event) (*dto.EventResponse, error) {
	// Load related data if needed
	var subcategory categories.Subcategory
	if err := s.db.First(&subcategory, "id = ?", event.SubcategoryID).Error; err != nil {
		return nil, err
	}

	var venue events.Venue
	if err := s.db.First(&venue, "id = ?", event.VenueID).Error; err != nil {
		return nil, err
	}

	var eventImages []events.EventImage
	if err := s.db.Where("event_id = ?", event.ID).Find(&eventImages).Error; err != nil {
		return nil, err
	}

	return &dto.EventResponse{
		ID:             event.ID,
		Title:          event.Title,
		Slug:           event.Slug,
		Description:    event.Description,
		SubcategoryID:  event.SubcategoryID,
		Subcategory:    subcategory,
		VenueID:        event.VenueID,
		Venue:          venue,
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
		Tags:           event.Tags,
		EventImages:    eventImages,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
	}, nil
}

// Valid fields for the events table
var validEventFields = map[string]bool{
	"id":              true,
	"title":           true,
	"subcategory_id":  true,
	"description":     true,
	"venue_id":        true,
	"total_seats":     true,
	"available_seats": true,
	"start_time":      true,
	"end_time":        true,
	"base_price":      true,
	"is_featured":     true,
	"status":          true,
	"created_at":      true,
	"updated_at":      true,
	"deleted_at":      true,
	"version":         true,
}
