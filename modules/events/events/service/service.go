package service

import (
	"context"
	"errors"
	"strings"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/users/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventResponse controls the JSON output for events
type EventResponse struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	SubcategoryID  string                 `json:"subcategory_id"`
	Subcategory    categories.Subcategory `json:"subcategory,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Venue          events.Venue           `json:"venue"`
	TotalSeats     int                    `json:"total_seats"`
	AvailableSeats int                    `json:"available_seats"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	PriceTierID    string                 `json:"price_tier_id"`
	BasePrice      float64                `json:"base_price"`
	IsFeatured     bool                   `json:"is_featured"`
	Status         string                 `json:"status"`
	EventImages    []events.EventImage    `json:"event_images,omitempty"`
}

type EventService interface {
	CreateEvent(userID, title, subcategoryID, description, venueID string, totalSeats int, basePrice float64, startTime, endTime time.Time, isFeatured bool) (*events.Event, error)
	UpdateEvent(userID, id, title, description, venueID, subcategoryID, priceTierID string, totalSeats int, basePrice float64, startTime, endTime time.Time, isFeatured bool, status string, imageURLs []string) (*events.Event, error)
	DeleteEvent(userID, id string) error
	GetEvent(userID, id, fields string) (*EventResponse, error) // Corrected syntax
	GetEvents(userID, fields string) ([]EventResponse, error)   // Corrected signature
	AddEventImage(userID, eventID, imageURL string, isPrimary bool) (*events.EventImage, error)
	DeleteEventImage(userID, eventID, imageID string) error
	HasPermission(userID, permission string) (bool, error)
}

type eventService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
	cloudinary           *cloudinary.CloudinaryService
}

func NewEventService(db *gorm.DB, authService authorization.PermissionService, cloudinary *cloudinary.CloudinaryService) EventService {
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
	"price_tier_id":   true,
	"base_price":      true,
	"is_featured":     true,
	"status":          true,
	// Metadata fields only if explicitly requested
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
	"version":    true,
}

func (s *eventService) CreateEvent(userID, title string, subcategoryID string, description, venueID string, totalSeats int, basePrice float64, startTime, endTime time.Time, isFeatured bool) (*events.Event, error) {
	_, err := s.HasPermission(userID, "create:events")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:events permission")
	// }

	// Start a transaction
	var createdEventID string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get organizer
		organizer, err := s.getUserOrganizer(userID)
		if err != nil {
			return err
		}

		// Check organizer status
		if organizer.Status != "active" {
			return errors.New("organizer is not active")
		}

		// Validate venue ID format
		if _, err := uuid.Parse(venueID); err != nil {
			return errors.New("invalid venue ID format")
		}

		// Validate subcategory ID format
		if _, err := uuid.Parse(subcategoryID); err != nil {
			return errors.New("invalid subcategory ID format")
		}

		// Validate subcategory exists
		var subcategory categories.Subcategory
		if err := tx.Where("id = ? AND deleted_at IS NULL", subcategoryID).First(&subcategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("subcategory not found")
			}
			return err
		}

		// Validate venue exists
		var venue events.Venue
		if err := tx.Where("id = ? AND deleted_at IS NULL", venueID).First(&venue).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("venue not found")
			}
			return err
		}

		// Create event
		event := events.Event{
			ID:             uuid.New().String(),
			Title:          title,
			SubcategoryID:  subcategoryID,
			Description:    description,
			VenueID:        venueID,
			TotalSeats:     totalSeats,
			AvailableSeats: totalSeats,
			StartTime:      startTime,
			EndTime:        endTime,
			OrganizerID:    organizer.ID,
			BasePrice:      basePrice,
			IsFeatured:     isFeatured,
			Status:         events.EventActive,
			Version:        1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		// Save the event ID for fetching after transaction
		createdEventID = event.ID
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch the created event to return
	var event events.Event
	if err := s.db.Where("id = ?", createdEventID).First(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *eventService) UpdateEvent(userID, id, title, description, venueID string, subcategoryID string, priceTierID string, totalSeats int, basePrice float64, startTime, endTime time.Time, isFeatured bool, status string, imageURLs []string) (*events.Event, error) {
	hasPerm, err := s.HasPermission(userID, "update:events")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:events permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	if _, err := uuid.Parse(venueID); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	if _, err := uuid.Parse(priceTierID); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	if _, err := uuid.Parse(subcategoryID); err != nil {
		return nil, errors.New("invalid subcategory ID format")
	}

	// Validate subcategory exists
	var subcategory categories.Subcategory
	if err := s.db.Where("id = ? AND deleted_at IS NULL", subcategoryID).First(&subcategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subcategory not found")
		}
		return nil, err
	}

	// Validate venue exists
	var venue events.Venue
	if err := s.db.Where("id = ? AND deleted_at IS NULL", venueID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	// Validate price tier exists
	var priceTier tickets.PriceTier
	if err := s.db.Where("id = ? AND deleted_at IS NULL", priceTierID).First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("price tier not found")
		}
		return nil, err
	}

	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	event.Title = title
	event.SubcategoryID = subcategoryID
	event.Description = description
	event.VenueID = venueID
	event.TotalSeats = totalSeats
	event.AvailableSeats = totalSeats // Assume full reset; adjust if bookings exist
	event.StartTime = startTime
	event.EndTime = endTime
	event.PriceTierID = priceTierID
	event.BasePrice = basePrice
	event.IsFeatured = isFeatured
	event.Status = events.EventStatus(status)
	event.Version++
	event.UpdatedAt = time.Now()

	if err := s.db.Save(&event).Error; err != nil {
		return nil, err
	}

	if len(imageURLs) > 0 {
		var existingImages []events.EventImage
		if err := s.db.Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&existingImages).Error; err != nil {
			return nil, err
		}
		for _, img := range existingImages {
			if err := s.cloudinary.DeleteFile(context.Background(), img.ImageURL); err != nil {
				return nil, err
			}
		}
		if err := s.db.Where("event_id = ?", event.ID).Delete(&events.EventImage{}).Error; err != nil {
			return nil, err
		}
		for i, url := range imageURLs {
			eventImage := &events.EventImage{
				ID:        uuid.New().String(),
				EventID:   event.ID,
				ImageURL:  url,
				IsPrimary: i == 0,
				CreatedAt: time.Now(),
			}
			if err := s.db.Create(eventImage).Error; err != nil {
				return nil, err
			}
		}
	}

	return &event, nil
}

func (s *eventService) DeleteEvent(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:events")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:events permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid event ID format")
	}

	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("event not found")
		}
		return err
	}

	if event.Status == events.EventActive {
		return errors.New("cannot delete an active event")
	}

	var eventImages []events.EventImage
	if err := s.db.Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&eventImages).Error; err != nil {
		return err
	}
	for _, img := range eventImages {
		if err := s.cloudinary.DeleteFile(context.Background(), img.ImageURL); err != nil {
			return err
		}
	}

	if err := s.db.Delete(&event).Error; err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetEvent(userID, id, fields string) (*EventResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:events")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:events permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	var event events.Event
	query := s.db.Preload("Venue").Preload("EventImages").
		Preload("Subcategory.Category"). // Preload Subcategory and its Category
		Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID)
	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validEventFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select("id", "title", "subcategory_id", "description", "venue_id", "total_seats", "available_seats", "start_time", "end_time", "price_tier_id", "base_price", "is_featured", "status")
		}
	} else {
		query = query.Select("id", "title", "subcategory_id", "description", "venue_id", "total_seats", "available_seats", "start_time", "end_time", "price_tier_id", "base_price", "is_featured", "status")
	}
	if err := query.First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	response := &EventResponse{
		ID:             event.ID,
		Title:          event.Title,
		SubcategoryID:  event.SubcategoryID,
		Subcategory:    event.Subcategory,
		Description:    event.Description,
		Venue:          event.Venue,
		TotalSeats:     event.TotalSeats,
		AvailableSeats: event.AvailableSeats,
		StartTime:      event.StartTime,
		EndTime:        event.EndTime,
		PriceTierID:    event.PriceTierID,
		BasePrice:      event.BasePrice,
		IsFeatured:     event.IsFeatured,
		Status:         string(event.Status),
		EventImages:    event.EventImages,
	}

	return response, nil
}

func (s *eventService) GetEvents(userID, fields string) ([]EventResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:events")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:events permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var events []events.Event
	query := s.db.Preload("Venue").Preload("EventImages").
		Preload("Subcategory.Category"). // Preload Subcategory and its Category
		Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID)
	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validEventFields[field] && field != "created_at" && field != "updated_at" && field != "deleted_at" && field != "version" {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		} else {
			query = query.Select("id", "title", "subcategory_id", "description", "venue_id", "total_seats", "available_seats", "start_time", "end_time", "price_tier_id", "base_price", "is_featured", "status")
		}
	} else {
		query = query.Select("id", "title", "subcategory_id", "description", "venue_id", "total_seats", "available_seats", "start_time", "end_time", "price_tier_id", "base_price", "is_featured", "status")
	}
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = EventResponse{
			ID:             event.ID,
			Title:          event.Title,
			SubcategoryID:  event.SubcategoryID,
			Subcategory:    event.Subcategory,
			Description:    event.Description,
			Venue:          event.Venue,
			TotalSeats:     event.TotalSeats,
			AvailableSeats: event.AvailableSeats,
			StartTime:      event.StartTime,
			EndTime:        event.EndTime,
			PriceTierID:    event.PriceTierID,
			BasePrice:      event.BasePrice,
			IsFeatured:     event.IsFeatured,
			Status:         string(event.Status),
			EventImages:    event.EventImages,
		}
	}

	return responses, nil
}

func (s *eventService) AddEventImage(userID, eventID, imageURL string, isPrimary bool) (*events.EventImage, error) {
	hasPerm, err := s.HasPermission(userID, "create:event_images")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:event_images permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(eventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", eventID, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	if isPrimary {
		if err := s.db.Model(&events.EventImage{}).Where("event_id = ? AND is_primary = ?", eventID, true).Update("is_primary", false).Error; err != nil {
			return nil, err
		}
	}

	eventImage := events.EventImage{
		ID:        uuid.New().String(),
		EventID:   eventID,
		ImageURL:  imageURL,
		IsPrimary: isPrimary,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&eventImage).Error; err != nil {
		return nil, err
	}

	return &eventImage, nil
}

func (s *eventService) DeleteEventImage(userID, eventID, imageID string) error {
	hasPerm, err := s.HasPermission(userID, "delete:event_images")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:event_images permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event ID format")
	}

	if _, err := uuid.Parse(imageID); err != nil {
		return errors.New("invalid image ID format")
	}

	var event events.Event
	dbResult := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", eventID, organizer.ID).First(&event)
	if dbResult.Error != nil {
		if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
			return errors.New("event not found")
		}
		return dbResult.Error
	}

	var eventImage events.EventImage
	if err := s.db.Where("id = ? AND event_id = ? AND deleted_at IS NULL", imageID, eventID).First(&eventImage); err != nil {
		if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
			return errors.New("event image not found")
		}
		return dbResult.Error
	}

	if err := s.cloudinary.DeleteFile(context.Background(), eventImage.ImageURL); err != nil {
		return err
	}

	if err := s.db.Delete(&eventImage).Error; err != nil {
		return err
	}

	return nil
}
