package price_tier_service

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

type PriceTierResponse struct {
	ID                 string               `json:"id"`
	OrganizerID        string               `json:"organizer_id"`
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	PercentageIncrease float64              `json:"percentage_increase"`
	Status             string               `json:"status"`
	IsDefault          bool                 `json:"is_default"`
	EffectiveFrom      time.Time            `json:"effective_from"`
	EffectiveTo        *time.Time           `json:"effective_to"`
	MinTickets         int                  `json:"min_tickets"`
	MaxTickets         *int                 `json:"max_tickets"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	Organizer          organizers.Organizer `json:"organizer,omitempty"`
}

type PriceTierService interface {
	CreatePriceTier(userID, name, description string, percentageIncrease float64, status string, isDefault bool, effectiveFrom time.Time, effectiveTo *time.Time, minTickets int, maxTickets *int) (*tickets.PriceTier, error)
	UpdatePriceTier(userID, id, name, description string, percentageIncrease float64, status string, isDefault *bool, effectiveFrom time.Time, effectiveTo *time.Time, minTickets int, maxTickets *int) (*tickets.PriceTier, error)
	DeletePriceTier(userID, id string) error
	GetPriceTier(userID, id, fields string) (*PriceTierResponse, error)
	GetPriceTiers(userID, fields string) ([]PriceTierResponse, error) // Removed organizerID parameter
	HasPermission(userID, permission string) (bool, error)
	GetAllPriceTiers(userID, fields string, page, limit int) ([]PriceTierResponse, error)
}

type priceTierService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
}

func NewPriceTierService(db *gorm.DB, authService authorization.PermissionService) PriceTierService {
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

func (s *priceTierService) CreatePriceTier(userID, name, description string, percentageIncrease float64, status string, isDefault bool, effectiveFrom time.Time, effectiveTo *time.Time, minTickets int, maxTickets *int) (*tickets.PriceTier, error) {
	// Check permissions
	_, err := s.HasPermission(userID, "create:price_tiers")
	// if err != nil || !hasPerm {
	// 	return nil, errors.New("user lacks create:price_tiers permission")
	// }

	// Get user's organizer automatically
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	// Check organizer status
	if organizer.Status != "active" {
		return nil, errors.New("organizer is not active")
	}

	// Set default status if not provided
	if status == "" {
		status = string(tickets.PriceTierActive)
	}

	// Create new price tier
	priceTier := &tickets.PriceTier{
		ID:                 uuid.New().String(),
		OrganizerID:        organizer.ID,
		Name:               name,
		Description:        description,
		PercentageIncrease: percentageIncrease,
		Status:             tickets.PriceTierStatus(status),
		IsDefault:          isDefault,
		EffectiveFrom:      effectiveFrom,
		EffectiveTo:        effectiveTo,
		MinTickets:         minTickets,
		MaxTickets:         maxTickets,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.db.Create(priceTier).Error; err != nil {
		return nil, err
	}

	return priceTier, nil
}

func (s *priceTierService) UpdatePriceTier(userID, id, name, description string, percentageIncrease float64, status string, isDefault *bool, effectiveFrom time.Time, effectiveTo *time.Time, minTickets int, maxTickets *int) (*tickets.PriceTier, error) {
	hasPerm, err := s.HasPermission(userID, "update:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks update:price_tiers permission")
	}

	// Validate price tier ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	// Get user's organizer automatically
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

	// Update fields
	if name != "" {
		priceTier.Name = name
	}
	if description != "" {
		priceTier.Description = description
	}
	if percentageIncrease >= 0 {
		priceTier.PercentageIncrease = percentageIncrease
	}
	if status != "" {
		priceTier.Status = tickets.PriceTierStatus(status)
	}
	if isDefault != nil {
		priceTier.IsDefault = *isDefault
	}
	if !effectiveFrom.IsZero() {
		priceTier.EffectiveFrom = effectiveFrom
	}
	if effectiveTo != nil {
		priceTier.EffectiveTo = effectiveTo
	}
	if minTickets >= 0 {
		priceTier.MinTickets = minTickets
	}
	if maxTickets != nil {
		priceTier.MaxTickets = maxTickets
	}

	priceTier.UpdatedAt = time.Now()

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

	// Get user's organizer automatically
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

func (s *priceTierService) GetPriceTier(userID, id, fields string) (*PriceTierResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks read:price_tiers permission")
	}

	// Validate price tier ID
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid price tier ID format")
	}

	// Get user's organizer automatically
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var priceTier tickets.PriceTier
	query := s.db.Preload("Organizer").
		Where("id = ? AND organizer_id = ?", id, organizer.ID)

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

	response := &PriceTierResponse{
		ID:                 priceTier.ID,
		OrganizerID:        priceTier.OrganizerID,
		Name:               priceTier.Name,
		Description:        priceTier.Description,
		PercentageIncrease: priceTier.PercentageIncrease,
		Status:             string(priceTier.Status),
		IsDefault:          priceTier.IsDefault,
		EffectiveFrom:      priceTier.EffectiveFrom,
		EffectiveTo:        priceTier.EffectiveTo,
		MinTickets:         priceTier.MinTickets,
		MaxTickets:         priceTier.MaxTickets,
		CreatedAt:          priceTier.CreatedAt,
		UpdatedAt:          priceTier.UpdatedAt,
		Organizer:          priceTier.Organizer,
	}

	return response, nil
}

func (s *priceTierService) GetPriceTiers(userID, fields string) ([]PriceTierResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks read:price_tiers permission")
	}

	// Get user's organizer automatically
	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var priceTiers []tickets.PriceTier
	query := s.db.Preload("Organizer").
		Where("organizer_id = ?", organizer.ID)

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

	responses := make([]PriceTierResponse, len(priceTiers))
	for i, pt := range priceTiers {
		responses[i] = PriceTierResponse{
			ID:                 pt.ID,
			OrganizerID:        pt.OrganizerID,
			Name:               pt.Name,
			Description:        pt.Description,
			PercentageIncrease: pt.PercentageIncrease,
			Status:             string(pt.Status),
			IsDefault:          pt.IsDefault,
			EffectiveFrom:      pt.EffectiveFrom,
			EffectiveTo:        pt.EffectiveTo,
			MinTickets:         pt.MinTickets,
			MaxTickets:         pt.MaxTickets,
			CreatedAt:          pt.CreatedAt,
			UpdatedAt:          pt.UpdatedAt,
			Organizer:          pt.Organizer,
		}
	}

	return responses, nil
}

// GetAllPriceTiers returns all price tiers for all organizers, with pagination.
func (s *priceTierService) GetAllPriceTiers(userID, fields string, page, limit int) ([]PriceTierResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:price_tiers")
	if err != nil || !hasPerm {
		return nil, errors.New("user lacks read:price_tiers permission")
	}

	var priceTiers []tickets.PriceTier
	query := s.db.Preload("Organizer")

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

	responses := make([]PriceTierResponse, len(priceTiers))
	for i, pt := range priceTiers {
		responses[i] = PriceTierResponse{
			ID:                 pt.ID,
			OrganizerID:        pt.OrganizerID,
			Name:               pt.Name,
			Description:        pt.Description,
			PercentageIncrease: pt.PercentageIncrease,
			Status:             string(pt.Status),
			IsDefault:          pt.IsDefault,
			EffectiveFrom:      pt.EffectiveFrom,
			EffectiveTo:        pt.EffectiveTo,
			MinTickets:         pt.MinTickets,
			MaxTickets:         pt.MaxTickets,
			CreatedAt:          pt.CreatedAt,
			UpdatedAt:          pt.UpdatedAt,
			Organizer:          pt.Organizer,
		}
	}

	return responses, nil
}
