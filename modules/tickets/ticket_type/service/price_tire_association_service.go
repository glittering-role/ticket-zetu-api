package ticket_type_service

import (
	"errors"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/tickets/ticket_type/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *ticketTypeService) AssociatePriceTier(userID, ticketTypeID string, input dto.AssociatePriceTierInput) (*dto.PriceTierResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "create:price_tiers")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:price_tiers permission")
	// }

	// Validate IDs
	if _, err := uuid.Parse(ticketTypeID); err != nil {
		return nil, errors.New("invalid ticket type ID format")
	}
	if _, err := uuid.Parse(input.PriceTierID); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Verify the ticket type belongs to the organizer
	var ticketType tickets.TicketType
	if err := s.db.Where("id = ? AND organizer_id = ?", ticketTypeID, organizer.ID).First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket type not found")
		}
		return nil, err
	}

	// Verify the price tier exists and belongs to the same organizer
	var priceTier tickets.PriceTier
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", input.PriceTierID, organizer.ID).First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("price tier not found or not accessible")
		}
		return nil, err
	}

	// Validate price tier status
	if priceTier.Status != tickets.PriceTierActive {
		return nil, errors.New("price tier must be active")
	}

	// Validate effective dates
	now := time.Now()
	if priceTier.EffectiveFrom.After(now) {
		return nil, errors.New("price tier is not yet effective")
	}
	if priceTier.EffectiveTo != nil && priceTier.EffectiveTo.Before(now) {
		return nil, errors.New("price tier has expired")
	}

	// Check ticket limits compatibility
	if priceTier.MaxTickets != nil && *priceTier.MaxTickets < ticketType.MaxTicketsPerUser {
		return nil, errors.New("price tier max_tickets is less than ticket type max_tickets_per_user")
	}
	if priceTier.MinTickets > ticketType.MinTicketsPerUser {
		return nil, errors.New("price tier min_tickets is greater than ticket type min_tickets_per_user")
	}

	// Check if the association already exists
	var count int64
	if err := s.db.Table("ticket_type_price_tiers").
		Where("ticket_type_id = ? AND price_tier_id = ?", ticketTypeID, input.PriceTierID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("price tier already associated with ticket type")
	}

	// Associate price tier with ticket type
	if err := s.db.Model(&ticketType).Association("PriceTiers").Append(&priceTier); err != nil {
		return nil, err
	}

	return s.toPriceTierDTO(&priceTier), nil
}

func (s *ticketTypeService) DisassociatePriceTier(userID, ticketTypeID, priceTierID string) error {
	hasPerm, err := s.HasPermission(userID, "delete:price_tiers")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:price_tiers permission")
	}

	// Validate IDs
	if _, err := uuid.Parse(ticketTypeID); err != nil {
		return errors.New("invalid ticket type ID format")
	}
	if _, err := uuid.Parse(priceTierID); err != nil {
		return errors.New("invalid price tier ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	// Verify the ticket type belongs to the organizer
	var ticketType tickets.TicketType
	if err := s.db.Where("id = ? AND organizer_id = ?", ticketTypeID, organizer.ID).First(&ticketType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ticket type not found")
		}
		return err
	}

	// Verify the price tier exists and belongs to the same organizer
	var priceTier tickets.PriceTier
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", priceTierID, organizer.ID).First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("price tier not found or not accessible")
		}
		return err
	}

	// Check if the association exists
	var count int64
	if err := s.db.Table("ticket_type_price_tiers").
		Where("ticket_type_id = ? AND price_tier_id = ?", ticketTypeID, priceTierID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("price tier not associated with ticket type")
	}

	// Remove the association
	if err := s.db.Model(&ticketType).Association("PriceTiers").Delete(&priceTier); err != nil {
		return err
	}

	return nil
}
