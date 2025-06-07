package price_tier_service

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/tickets/price_tires/dto"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceTierService interface {
	CreatePriceTier(userID string, input dto.CreatePriceTierRequest) (*tickets.PriceTier, error)
	UpdatePriceTier(userID, id string, input dto.UpdatePriceTierRequest) (*tickets.PriceTier, error)
	DeletePriceTier(userID, id string) error
	GetPriceTier(userID, id, fields string) (*dto.GetPriceTierResponse, error)
	GetPriceTiers(userID, fields string) ([]dto.GetPriceTierResponse, error)
	GetAllPriceTiers(userID, fields string, page, limit int) ([]dto.GetPriceTierResponse, error)
	HasPermission(userID, permission string) (bool, error)
}

type priceTierService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
}

func NewPriceTierService(db *gorm.DB, authService authorization_service.PermissionService) PriceTierService {
	return &priceTierService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *priceTierService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *priceTierService) toDTO(pt *tickets.PriceTier) *dto.GetPriceTierResponse {
	if pt == nil {
		return nil
	}

	return &dto.GetPriceTierResponse{
		ID:            pt.ID,
		OrganizerID:   pt.OrganizerID,
		Name:          pt.Name,
		Description:   pt.Description,
		BasePrice:     pt.BasePrice,
		Status:        string(pt.Status),
		IsDefault:     pt.IsDefault,
		EffectiveFrom: pt.EffectiveFrom,
		EffectiveTo:   pt.EffectiveTo,
		MinTickets:    pt.MinTickets,
		MaxTickets:    pt.MaxTickets,
		Version:       pt.Version,
		CreatedAt:     pt.CreatedAt,
		UpdatedAt:     pt.UpdatedAt,
	}
}

func (s *priceTierService) getUserOrganizer(userID string) (*organizers.Organizer, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}
	return &organizer, nil
}

var validPriceTierFields = map[string]bool{
	"id":                  true,
	"organizer_id":        true,
	"name":                true,
	"description":         true,
	"base_price":          true,
	"percentage_increase": true,
	"status":              true,
	"is_default":          true,
	"effective_from":      true,
	"effective_to":        true,
	"min_tickets":         true,
	"max_tickets":         true,
	"created_at":          true,
	"updated_at":          true,
}

func (s *priceTierService) CreatePriceTier(userID string, input dto.CreatePriceTierRequest) (*tickets.PriceTier, error) {
	hasPerm, err := s.HasPermission(userID, "create:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks create:price_tiers permission")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Check organizer status
	if organizer.Status != "active" {
		return nil, errors.New("organizer is not active")
	}

	// If this is being set as default, unset any existing default for this organizer
	if input.IsDefault {
		if err := s.db.Model(&tickets.PriceTier{}).
			Where("organizer_id = ? AND is_default = ?", organizer.ID, true).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	priceTier := &tickets.PriceTier{
		ID:            uuid.New().String(),
		OrganizerID:   organizer.ID,
		Name:          input.Name,
		Description:   input.Description,
		BasePrice:     input.BasePrice,
		Status:        tickets.PriceTierStatus("active"),
		IsDefault:     input.IsDefault,
		EffectiveFrom: input.EffectiveFrom,
		EffectiveTo:   input.EffectiveTo,
		MinTickets:    input.MinTickets,
		MaxTickets:    input.MaxTickets,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(priceTier).Error; err != nil {
		return nil, err
	}

	return priceTier, nil
}

func (s *priceTierService) UpdatePriceTier(userID, id string, input dto.UpdatePriceTierRequest) (*tickets.PriceTier, error) {
	hasPerm, err := s.HasPermission(userID, "update:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks update:price_tiers permission")
	}

	// Validate price tier ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var priceTier tickets.PriceTier
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("price tier not found")
		}
		return nil, err
	}

	// If this is being set as default, unset any existing default for this organizer
	if input.IsDefault != nil && *input.IsDefault {
		if err := s.db.Model(&tickets.PriceTier{}).
			Where("organizer_id = ? AND is_default = ? AND id != ?", organizer.ID, true, id).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	// Update fields
	if input.Name != nil {
		priceTier.Name = *input.Name
	}
	if input.Description != nil {
		priceTier.Description = *input.Description
	}
	if input.BasePrice != nil {
		priceTier.BasePrice = *input.BasePrice
	}

	if input.Status != nil {
		priceTier.Status = tickets.PriceTierStatus(*input.Status)
	}
	if input.IsDefault != nil {
		priceTier.IsDefault = *input.IsDefault
	}
	if input.EffectiveFrom != nil {
		priceTier.EffectiveFrom = *input.EffectiveFrom
	}
	if input.EffectiveTo != nil {
		priceTier.EffectiveTo = input.EffectiveTo
	}
	if input.MinTickets != nil {
		priceTier.MinTickets = *input.MinTickets
	}
	if input.MaxTickets != nil {
		priceTier.MaxTickets = input.MaxTickets
	}

	priceTier.UpdatedAt = time.Now()
	priceTier.Version++

	if err := s.db.Save(&priceTier).Error; err != nil {
		return nil, err
	}

	return &priceTier, nil
}

func (s *priceTierService) DeletePriceTier(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:price_tiers")
	if err != nil || !hasPerm {
		return errors.New("user lacks delete:price_tiers permission")
	}

	// Validate price tier ID
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid price tier ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	var priceTier tickets.PriceTier
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("price tier not found")
		}
		return err
	}

	// Check if it's a default price tier
	if priceTier.IsDefault {
		return errors.New("cannot delete default price tier")
	}

	// Check if the price tier is in use by any events
	var count int64
	if err := s.db.Model(&events.Event{}).Where("price_tier_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("price tier is in use by events")
	}

	if err := s.db.Delete(&priceTier).Error; err != nil {
		return err
	}

	return nil
}

func (s *priceTierService) GetPriceTier(userID, id, fields string) (*dto.GetPriceTierResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks read:price_tiers permission")
	}

	// Validate price tier ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var priceTier tickets.PriceTier
	query := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validPriceTierFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.First(&priceTier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("price tier not found")
		}
		return nil, err
	}

	return s.toDTO(&priceTier), nil
}

func (s *priceTierService) GetPriceTiers(userID, fields string) ([]dto.GetPriceTierResponse, error) {
	_, err := s.HasPermission(userID, "read:price_tiers")
	// if err != nil || !hasPerm {
	// 	return nil, errors.New("user lacks read:price_tiers permission")
	// }

	// Get user's organizer
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var priceTiers []tickets.PriceTier
	query := s.db.Where("organizer_id = ?", organizer.ID)

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validPriceTierFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if err := query.Find(&priceTiers).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.GetPriceTierResponse, 0, len(priceTiers))
	for _, pt := range priceTiers {
		resp := s.toDTO(&pt)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}

func (s *priceTierService) GetAllPriceTiers(userID, fields string, page, limit int) ([]dto.GetPriceTierResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "read:price_tiers")
	// if err != nil || !hasPerm {
	// 	return nil, errors.New("user lacks read:price_tiers permission")
	// }

	var priceTiers []tickets.PriceTier
	query := s.db

	if fields != "" {
		selectedFields := []string{}
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if validPriceTierFields[field] {
				selectedFields = append(selectedFields, field)
			}
		}
		if len(selectedFields) > 0 {
			query = query.Select(selectedFields)
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	if err := query.Find(&priceTiers).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.GetPriceTierResponse, 0, len(priceTiers))
	for _, pt := range priceTiers {
		resp := s.toDTO(&pt)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}
