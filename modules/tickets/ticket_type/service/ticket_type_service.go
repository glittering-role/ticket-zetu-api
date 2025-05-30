package ticket_type_service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/users/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTypeResponse struct {
	ID                string       `json:"id"`
	EventID           string       `json:"event_id"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	PriceModifier     float64      `json:"price_modifier"`
	Benefits          string       `json:"benefits"`
	MaxTicketsPerUser int          `json:"max_tickets_per_user"`
	Status            string       `json:"status"`
	IsDefault         bool         `json:"is_default"`
	SalesStart        time.Time    `json:"sales_start"`
	SalesEnd          *time.Time   `json:"sales_end"`
	QuantityAvailable *int         `json:"quantity_available"`
	MinTicketsPerUser int          `json:"min_tickets_per_user"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Event             events.Event `json:"event,omitempty"`
}

type TicketTypeService interface {
	CreateTicketType(userID, eventID, name, description string, priceModifier float64, benefits string, maxTicketsPerUser int, status string, isDefault bool, salesStart time.Time, salesEnd *time.Time, quantityAvailable *int, minTicketsPerUser int) (*tickets.TicketType, error)
	UpdateTicketType(userID, id, name, description string, priceModifier float64, benefits string, maxTicketsPerUser int, status string, isDefault *bool, salesStart time.Time, salesEnd *time.Time, quantityAvailable *int, minTicketsPerUser int) (*tickets.TicketType, error)
	DeleteTicketType(userID, id string) error
	GetTicketType(userID, id, fields string) (*TicketTypeResponse, error)
	GetTicketTypes(userID, eventID, fields string) ([]TicketTypeResponse, error)
	HasPermission(userID, permission string) (bool, error)
}

type ticketTypeService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
}

func NewTicketTypeService(db *gorm.DB, authService authorization.PermissionService) TicketTypeService {
	return &ticketTypeService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *ticketTypeService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *ticketTypeService) getUserOrganizer(userID string) (*organizers.Organizer, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}
	return &organizer, nil
}

var validTicketTypeFields = map[string]bool{
	"id":                   true,
	"event_id":             true,
	"name":                 true,
	"description":          true,
	"price_modifier":       true,
	"benefits":             true,
	"max_tickets_per_user": true,
	"status":               true,
	"is_default":           true,
	"sales_start":          true,
	"sales_end":            true,
	"quantity_available":   true,
	"min_tickets_per_user": true,
	"created_at":           true,
	"updated_at":           true,
}

func (s *ticketTypeService) CreateTicketType(userID, eventID, name, description string, priceModifier float64, benefits string, maxTicketsPerUser int, status string, isDefault bool, salesStart time.Time, salesEnd *time.Time, quantityAvailable *int, minTicketsPerUser int) (*tickets.TicketType, error) {
	hasPerm, err := s.HasPermission(userID, "create:ticket_types")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:ticket_types permission")
	}

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Verify the event belongs to the organizer
	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	// Set default status if not provided
	if status == "" {
		status = string(tickets.TicketTypeActive)
	}

	// If this is being set as default, unset any existing default for this event
	if isDefault {
		if err := s.db.Model(&tickets.TicketType{}).
			Where("event_id = ? AND is_default = ?", eventID, true).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	ticketType := &tickets.TicketType{
		ID:                uuid.New().String(),
		EventID:           eventID,
		OrganizerID:       organizer.ID,
		Name:              name,
		Description:       description,
		PriceModifier:     priceModifier,
		Benefits:          benefits,
		MaxTicketsPerUser: maxTicketsPerUser,
		Status:            tickets.TicketTypeStatus(status),
		IsDefault:         isDefault,
		SalesStart:        salesStart,
		SalesEnd:          salesEnd,
		QuantityAvailable: quantityAvailable,
		MinTicketsPerUser: minTicketsPerUser,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.db.Create(ticketType).Error; err != nil {
		return nil, err
	}

	return ticketType, nil
}

func (s *ticketTypeService) UpdateTicketType(userID, id, name, description string, priceModifier float64, benefits string, maxTicketsPerUser int, status string, isDefault *bool, salesStart time.Time, salesEnd *time.Time, quantityAvailable *int, minTicketsPerUser int) (*tickets.TicketType, error) {
	hasPerm, err := s.HasPermission(userID, "update:ticket_types")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:ticket_types permission")
	}

	// Validate ticket type ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid ticket type ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var ticketType tickets.TicketType
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket type not found")
		}
		return nil, err
	}

	// If this is being set as default, unset any existing default for this event
	if isDefault != nil && *isDefault {
		if err := s.db.Model(&tickets.TicketType{}).
			Where("event_id = ? AND is_default = ? AND id != ?", ticketType.EventID, true, id).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	// Update fields
	if name != "" {
		ticketType.Name = name
	}
	if description != "" {
		ticketType.Description = description
	}
	if priceModifier >= 0 {
		ticketType.PriceModifier = priceModifier
	}
	if benefits != "" {
		ticketType.Benefits = benefits
	}
	if maxTicketsPerUser > 0 {
		ticketType.MaxTicketsPerUser = maxTicketsPerUser
	}
	if status != "" {
		ticketType.Status = tickets.TicketTypeStatus(status)
	}
	if isDefault != nil {
		ticketType.IsDefault = *isDefault
	}
	if !salesStart.IsZero() {
		ticketType.SalesStart = salesStart
	}
	if salesEnd != nil {
		ticketType.SalesEnd = salesEnd
	}
	if quantityAvailable != nil {
		ticketType.QuantityAvailable = quantityAvailable
	}
	if minTicketsPerUser > 0 {
		ticketType.MinTicketsPerUser = minTicketsPerUser
	}

	ticketType.UpdatedAt = time.Now()
	ticketType.Version++

	if err := s.db.Save(&ticketType).Error; err != nil {
		return nil, err
	}

	return &ticketType, nil
}

func (s *ticketTypeService) DeleteTicketType(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:ticket_types")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:ticket_types permission")
	}

	// Validate ticket type ID
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid ticket type ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	var ticketType tickets.TicketType
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ticket type not found")
		}
		return err
	}

	// Check if it's a default ticket type
	if ticketType.IsDefault {
		return errors.New("cannot delete default ticket type")
	}

	// Check if the ticket type is in use
	var count int64
	if err := s.db.Model(&tickets.Ticket{}).Where("ticket_type_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("ticket type is in use")
	}

	if err := s.db.Delete(&ticketType).Error; err != nil {
		return err
	}

	return nil
}

func (s *ticketTypeService) GetTicketType(userID, id, fields string) (*TicketTypeResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:ticket_types")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:ticket_types permission")
	}

	// Validate ticket type ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid ticket type ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var ticketType tickets.TicketType
	query := s.db.Preload("Event").
		Where("id = ? AND organizer_id = ?", id, organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validTicketTypeFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket type not found")
		}
		return nil, err
	}

	response := &TicketTypeResponse{
		ID:                ticketType.ID,
		EventID:           ticketType.EventID,
		Name:              ticketType.Name,
		Description:       ticketType.Description,
		PriceModifier:     ticketType.PriceModifier,
		Benefits:          ticketType.Benefits,
		MaxTicketsPerUser: ticketType.MaxTicketsPerUser,
		Status:            string(ticketType.Status),
		IsDefault:         ticketType.IsDefault,
		SalesStart:        ticketType.SalesStart,
		SalesEnd:          ticketType.SalesEnd,
		QuantityAvailable: ticketType.QuantityAvailable,
		MinTicketsPerUser: ticketType.MinTicketsPerUser,
		CreatedAt:         ticketType.CreatedAt,
		UpdatedAt:         ticketType.UpdatedAt,
		Event:             ticketType.Event,
	}

	return response, nil
}

func (s *ticketTypeService) GetTicketTypes(userID, eventID, fields string) ([]TicketTypeResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:ticket_types")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:ticket_types permission")
	}

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Verify the event belongs to the organizer
	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	var ticketTypes []tickets.TicketType
	query := s.db.Preload("Event").
		Where("event_id = ?", eventID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validTicketTypeFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.Find(&ticketTypes).Error; err != nil {
		return nil, err
	}

	responses := make([]TicketTypeResponse, len(ticketTypes))
	for i, tt := range ticketTypes {
		responses[i] = TicketTypeResponse{
			ID:                tt.ID,
			EventID:           tt.EventID,
			Name:              tt.Name,
			Description:       tt.Description,
			PriceModifier:     tt.PriceModifier,
			Benefits:          tt.Benefits,
			MaxTicketsPerUser: tt.MaxTicketsPerUser,
			Status:            string(tt.Status),
			IsDefault:         tt.IsDefault,
			SalesStart:        tt.SalesStart,
			SalesEnd:          tt.SalesEnd,
			QuantityAvailable: tt.QuantityAvailable,
			MinTicketsPerUser: tt.MinTicketsPerUser,
			CreatedAt:         tt.CreatedAt,
			UpdatedAt:         tt.UpdatedAt,
			Event:             tt.Event,
		}
	}

	return responses, nil
}
