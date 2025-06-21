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
	notification_service "ticket-zetu-api/modules/notifications/service"
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
	Query     string
	StartDate *time.Time
	EndDate   *time.Time
	EventType string
	IsFree    *bool
	Status    string
	MinPrice  *float64
	MaxPrice  *float64
	Page      int
	PageSize  int
}

// PaginatedResponse wraps the paginated event results with metadata
type PaginatedResponse struct {
	Events      []dto.MinimalEventResponse `json:"events"`
	TotalItems  int64                      `json:"total_items"`
	CurrentPage int                        `json:"current_page"`
	TotalPages  int                        `json:"total_pages"`
}

type eventService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	cloudinary           *cloudinary.CloudinaryService
	notificationService  notification_service.NotificationService
	contentFilter        *ContentFilter
}

type EventService interface {
	CreateEvent(createDto dto.CreateEvent, userID string) (*dto.EventResponse, error)
	UpdateEvent(updateDto dto.UpdateEvent, userID, id string) (*dto.EventResponse, error)
	DeleteEvent(userID, id string) error
	GetEvent(userID, id string) (*dto.EventResponse, error)
	GetEvents(userID string, filter SearchFilter) (*PaginatedResponse, error)
	SearchEvents(userID string, filter SearchFilter) (*PaginatedResponse, error)
	AddEventImage(userID, eventID, imageURL string, isPrimary bool) (*events.EventImage, error)
	DeleteEventImage(userID, eventID, imageID string) error
	HasPermission(userID, permission string) (bool, error)

	ToggleFavorite(userID, eventID string) error
	ToggleUpvote(userID, eventID string) (string, error)
	ToggleDownvote(userID, eventID string) (string, error)
	AddComment(userID, eventID, content string) (*events.Comment, error)
	AddReply(userID, eventID, commentID, content string) (*events.Comment, error)
	EditComment(userID, commentID, newContent string) (*events.Comment, error)
	DeleteComment(userID, commentID string) error
	GetUserFavorites(userID string) ([]events.Favorite, error)
	GetUserComments(userID string) ([]events.Comment, error)
}

func NewEventService(db *gorm.DB, authService authorization_service.PermissionService, cloudinary *cloudinary.CloudinaryService, notificationService notification_service.NotificationService) EventService {
	return &eventService{
		db:                   db,
		authorizationService: authService,
		cloudinary:           cloudinary,
		notificationService:  notificationService,
		contentFilter:        NewContentFilter(),
	}
}

// sendNotification sends a notification with the provided details
func (s *eventService) sendNotification(notificationType, title, message, senderID, eventID string, recipientIDs []string, metadata map[string]interface{}) error {
	notificationErr := s.notificationService.TriggerNotification(
		"events",
		notificationType,
		title,
		message,
		senderID,
		eventID,
		recipientIDs,
		metadata,
	)
	if notificationErr != nil {
		fmt.Printf("Failed to send %s notification: %v\n", notificationType, notificationErr)
	}
	return notificationErr
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

	// Load vote counts
	var upvotes, downvotes int64
	if err := s.db.Model(&events.Vote{}).
		Where("event_id = ? AND type = ?", event.ID, events.VoteTypeUp).
		Count(&upvotes).Error; err != nil {
		return nil, fmt.Errorf("failed to count upvotes: %v", err)
	}
	if err := s.db.Model(&events.Vote{}).
		Where("event_id = ? AND type = ?", event.ID, events.VoteTypeDown).
		Count(&downvotes).Error; err != nil {
		return nil, fmt.Errorf("failed to count downvotes: %v", err)
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
		Upvotes:     int(upvotes),
		Downvotes:   int(downvotes),
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
		Upvotes:        int(upvotes),
		Downvotes:      int(downvotes),
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
