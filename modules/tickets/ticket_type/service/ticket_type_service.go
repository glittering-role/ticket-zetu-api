package ticket_type_service

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/tickets/ticket_type/dto"
	"ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTypeService interface {
	CreateTicketType(userID string, input dto.CreateTicketTypeInput) (*dto.TicketTypeResponse, error)
	UpdateTicketType(userID string, id string, nput dto.UpdateTicketTypeInput) (*dto.TicketTypeResponse, error)
	DeleteTicketType(userID, id string) error
	GetTicketType(userID, id, fields string) (*dto.TicketTypeResponse, error)
	GetTicketTypes(userID, eventID, fields string) ([]dto.TicketTypeResponse, error)
	HasPermission(userID, permission string) (bool, error)
	GetAllTicketTypesForOrganization(userID, fields string) ([]dto.TicketTypeResponse, error)
	AssociatePriceTier(userID, ticketTypeID string, input dto.AssociatePriceTierInput) (*dto.PriceTierResponse, error)
	DisassociatePriceTier(userID, ticketTypeID, priceTierID string) error
}

type ticketTypeService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
}

func NewTicketTypeService(db *gorm.DB, authService authorization_service.PermissionService) TicketTypeService {
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

func (s *ticketTypeService) toDTO(ticketType *tickets.TicketType) *dto.TicketTypeResponse {
	var priceTiers []tickets.PriceTier
	if err := s.db.Model(ticketType).Association("PriceTiers").Find(&priceTiers); err != nil {
		// Log error if needed, but don't fail the response
	}

	priceTierResponses := make([]dto.PriceTierResponse, len(priceTiers))
	for i, pt := range priceTiers {
		priceTierResponses[i] = *s.toPriceTierDTO(&pt)
	}

	return &dto.TicketTypeResponse{
		ID:                ticketType.ID,
		EventID:           ticketType.EventID,
		Name:              ticketType.Name,
		Description:       ticketType.Description,
		PriceModifier:     ticketType.PriceModifier,
		Benefits:          ticketType.Benefits,
		MaxTicketsPerUser: ticketType.MaxTicketsPerUser,
		MinTicketsPerUser: ticketType.MinTicketsPerUser,
		Status:            string(ticketType.Status),
		IsDefault:         ticketType.IsDefault,
		SalesStart:        ticketType.SalesStart,
		SalesEnd:          ticketType.SalesEnd,
		QuantityAvailable: ticketType.QuantityAvailable,
		CreatedAt:         ticketType.CreatedAt,
		UpdatedAt:         ticketType.UpdatedAt,
		PriceTiers:        priceTierResponses,
	}
}

func (s *ticketTypeService) toPriceTierDTO(priceTier *tickets.PriceTier) *dto.PriceTierResponse {
	return &dto.PriceTierResponse{
		ID:            priceTier.ID,
		OrganizerID:   priceTier.OrganizerID,
		Name:          priceTier.Name,
		Description:   priceTier.Description,
		BasePrice:     priceTier.BasePrice,
		Status:        string(priceTier.Status),
		IsDefault:     priceTier.IsDefault,
		EffectiveFrom: priceTier.EffectiveFrom,
		EffectiveTo:   priceTier.EffectiveTo,
		MinTickets:    priceTier.MinTickets,
		MaxTickets:    priceTier.MaxTickets,
		CreatedAt:     priceTier.CreatedAt,
		UpdatedAt:     priceTier.UpdatedAt,
	}
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

func (s *ticketTypeService) CreateTicketType(userID string, input dto.CreateTicketTypeInput) (*dto.TicketTypeResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "create:ticket_types")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:ticket_types permission")
	// }

	// Validate event ID
	if _, err := uuid.Parse(input.EventID); err != nil {
		return nil, errors.New("invalid event ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Verify the event belongs to the organizer
	var event events.Event
	if err := s.db.Where("id = ? AND organizer_id = ?", input.EventID, organizer.ID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	// If this is being set as default, unset any existing default for this event
	if input.IsDefault {
		if err := s.db.Model(&tickets.TicketType{}).
			Where("event_id = ? AND is_default = ?", input.EventID, true).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	ticketType := &tickets.TicketType{
		ID:                uuid.New().String(),
		EventID:           input.EventID,
		OrganizerID:       organizer.ID,
		Name:              input.Name,
		Description:       input.Description,
		PriceModifier:     input.PriceModifier,
		Benefits:          input.Benefits,
		MaxTicketsPerUser: input.MaxTicketsPerUser,
		Status:            tickets.TicketTypeStatus(input.Status),
		IsDefault:         input.IsDefault,
		SalesStart:        input.SalesStart,
		SalesEnd:          input.SalesEnd,
		QuantityAvailable: input.QuantityAvailable,
		MinTicketsPerUser: input.MinTicketsPerUser,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.db.Create(ticketType).Error; err != nil {
		return nil, err
	}

	return s.toDTO(ticketType), nil
}

func (s *ticketTypeService) UpdateTicketType(userID string, id string, input dto.UpdateTicketTypeInput) (*dto.TicketTypeResponse, error) {
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
	if input.IsDefault {
		if err := s.db.Model(&tickets.TicketType{}).
			Where("event_id = ? AND is_default = ? AND id != ?", ticketType.EventID, true, input.ID).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	// Update fields
	ticketType.Name = input.Name
	ticketType.Description = input.Description
	ticketType.PriceModifier = input.PriceModifier
	ticketType.Benefits = input.Benefits
	ticketType.MaxTicketsPerUser = input.MaxTicketsPerUser
	ticketType.Status = tickets.TicketTypeStatus(input.Status)
	ticketType.IsDefault = input.IsDefault
	ticketType.SalesStart = input.SalesStart
	ticketType.SalesEnd = input.SalesEnd
	ticketType.QuantityAvailable = input.QuantityAvailable
	ticketType.MinTicketsPerUser = input.MinTicketsPerUser

	ticketType.UpdatedAt = time.Now()
	ticketType.Version++

	if err := s.db.Save(&ticketType).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&ticketType), nil
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
