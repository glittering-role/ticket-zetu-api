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
	"ticket-zetu-api/modules/tickets/models/tickets"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// SearchFilter defines the parameters for searching and filtering events
type SearchFilter struct {
	Query     string     // Search term for title or description
	StartDate *time.Time // Filter by events starting on or after this date
	EndDate   *time.Time // Filter by events ending on or before this date
	EventType string     // Filter by event type (online, offline, hybrid)
	IsFree    *bool      // Filter by free or paid events
	Status    string     // Filter by event status
	MinPrice  *float64   // Filter by minimum ticket price
	MaxPrice  *float64   // Filter by maximum ticket price
	Page      int        // Page number (1-based)
	PageSize  int        // Number of items per page (default: 20)
}

// PaginatedResponse wraps the paginated event results with metadata
type PaginatedResponse struct {
	Events      []dto.MinimalEventResponse `json:"events"`
	TotalItems  int64                      `json:"total_items"`
	CurrentPage int                        `json:"current_page"`
	TotalPages  int                        `json:"total_pages"`
}

type EventService interface {
	CreateEvent(createDto dto.CreateEvent, userID string) (*dto.EventResponse, error)
	UpdateEvent(updateDto dto.UpdateEvent, userID, id string) (*dto.EventResponse, error)
	DeleteEvent(userID, id string) error
	GetEvent(userID, id string) (*dto.EventResponse, error)
	GetEvents(userID string, filter SearchFilter) (*PaginatedResponse, error) // Updated signature
	SearchEvents(userID string, filter SearchFilter) (*PaginatedResponse, error)
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

func (s *eventService) generateSlug(title string) (string, error) {
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
	// Load event images only
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
		// TicketTypes omitted from MinimalEventResponse to improve performance
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

	// Load ticket types and their price tiers for full response
	var ticketTypes []tickets.TicketType
	if err := s.db.Preload("PriceTiers").Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&ticketTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch ticket types: %v", err)
	}

	// Convert ticket types to DTO
	ticketTypeResponses := make([]dto.TicketTypeResponse, len(ticketTypes))
	for i, tt := range ticketTypes {
		priceTierResponses := make([]dto.PriceTierResponse, len(tt.PriceTiers))
		for j, pt := range tt.PriceTiers {
			priceTierResponses[j] = dto.PriceTierResponse{
				ID:            pt.ID,
				Name:          pt.Name,
				Description:   pt.Description,
				BasePrice:     pt.BasePrice,
				Status:        pt.Status,
				IsDefault:     pt.IsDefault,
				EffectiveFrom: pt.EffectiveFrom,
				EffectiveTo:   pt.EffectiveTo,
				MinTickets:    pt.MinTickets,
				MaxTickets:    pt.MaxTickets,
				CreatedAt:     pt.CreatedAt,
				UpdatedAt:     pt.UpdatedAt,
			}
		}
		ticketTypeResponses[i] = dto.TicketTypeResponse{
			ID:                tt.ID,
			Name:              tt.Name,
			Description:       tt.Description,
			PriceModifier:     tt.PriceModifier,
			Benefits:          tt.Benefits,
			MinTicketsPerUser: tt.MinTicketsPerUser,
			MaxTicketsPerUser: tt.MaxTicketsPerUser,
			QuantityAvailable: tt.QuantityAvailable,
			Status:            tt.Status,
			IsDefault:         tt.IsDefault,
			SalesStart:        tt.SalesStart,
			SalesEnd:          tt.SalesEnd,
			PriceTiers:        priceTierResponses,
			CreatedAt:         tt.CreatedAt,
			UpdatedAt:         tt.UpdatedAt,
		}
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
		TicketTypes:    ticketTypeResponses,
	}

	return &struct {
		Full    dto.EventResponse
		Minimal dto.MinimalEventResponse
	}{Full: full}, nil
}
