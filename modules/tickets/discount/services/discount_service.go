package discount_service

import (
	"errors"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/users/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountResponse struct {
	ID            string                 `json:"id"`
	Code          string                 `json:"code"`
	EventID       string                 `json:"event_id"`
	Event         events.Event           `json:"event,omitempty"`
	DiscountType  tickets.DiscountType   `json:"discount_type"`
	DiscountValue float64                `json:"discount_value"`
	ValidFrom     time.Time              `json:"valid_from"`
	ValidUntil    time.Time              `json:"valid_until"`
	MaxUses       int                    `json:"max_uses"`
	CurrentUses   int                    `json:"current_uses"`
	IsActive      bool                   `json:"is_active"`
	Source        tickets.DiscountSource `json:"source"`
	PromoterID    string                 `json:"promoter_id,omitempty"`
	MinOrderValue float64                `json:"min_order_value"`
	IsSingleUse   bool                   `json:"is_single_use"`
	OrganizerID   string                 `json:"organizer_id"`
	Organizer     organizers.Organizer   `json:"organizer,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

type DiscountService interface {
	CreateDiscount(userID string, discount *tickets.DiscountCode) (*DiscountResponse, error)
	GetDiscount(userID, id string) (*DiscountResponse, error)
	GetDiscounts(userID string) ([]DiscountResponse, error)
	UpdateDiscount(userID, id string, discount *tickets.DiscountCode) (*DiscountResponse, error)
	CancelDiscount(userID, id string) error
	ValidateDiscountCode(code string, eventID string, orderValue float64) (*tickets.DiscountCode, error)
}

type discountService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
	logHandler           *handler.LogHandler
}

func NewDiscountService(db *gorm.DB, authService authorization.PermissionService, logHandler *handler.LogHandler) DiscountService {
	return &discountService{
		db:                   db,
		authorizationService: authService,
		logHandler:           logHandler,
	}
}

func (s *discountService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *discountService) getUserOrganizer(userID string) (*organizers.Organizer, error) {
	var organizer organizers.Organizer
	if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}
	return &organizer, nil
}

func (s *discountService) CreateDiscount(userID string, discount *tickets.DiscountCode) (*DiscountResponse, error) {
	hasPerm, err := s.HasPermission(userID, "create:discounts")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:discounts permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	discount.OrganizerID = organizer.ID

	if discount.EventID != "" {
		var event events.Event
		if err := s.db.Where("id = ? AND organizer_id = ?", discount.EventID, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("event not found or not owned by organizer")
			}
			return nil, err
		}
	}

	if err := s.db.Create(discount).Error; err != nil {
		return nil, err
	}

	return s.mapToDiscountResponse(discount), nil
}

func (s *discountService) GetDiscount(userID, id string) (*DiscountResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:discounts")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:discounts permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var discount tickets.DiscountCode
	if err := s.db.Preload("Event").Preload("Organizer").
		Where("id = ? AND organizer_id = ?", id, organizer.ID).
		First(&discount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("discount not found")
		}
		return nil, err
	}

	return s.mapToDiscountResponse(&discount), nil
}

func (s *discountService) GetDiscounts(userID string) ([]DiscountResponse, error) {
	hasPerm, err := s.HasPermission(userID, "read:discounts")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:discounts permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var discounts []tickets.DiscountCode
	if err := s.db.Preload("Event").Preload("Organizer").
		Where("organizer_id = ?", organizer.ID).
		Find(&discounts).Error; err != nil {
		return nil, err
	}

	responses := make([]DiscountResponse, len(discounts))
	for i, discount := range discounts {
		responses[i] = *s.mapToDiscountResponse(&discount)
	}

	return responses, nil
}

func (s *discountService) UpdateDiscount(userID, id string, discount *tickets.DiscountCode) (*DiscountResponse, error) {
	hasPerm, err := s.HasPermission(userID, "update:discounts")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:discounts permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	var existingDiscount tickets.DiscountCode
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&existingDiscount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("discount not found")
		}
		return nil, err
	}

	// Prevent updating certain fields
	discount.ID = existingDiscount.ID
	discount.OrganizerID = existingDiscount.OrganizerID
	discount.Source = existingDiscount.Source
	discount.PromoterID = existingDiscount.PromoterID
	discount.CurrentUses = existingDiscount.CurrentUses

	if err := s.db.Save(discount).Error; err != nil {
		return nil, err
	}

	return s.mapToDiscountResponse(discount), nil
}

func (s *discountService) CancelDiscount(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "update:discounts")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks update:discounts permission")
	}

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return err
	}

	var discount tickets.DiscountCode
	if err := s.db.Where("id = ? AND organizer_id = ?", id, organizer.ID).First(&discount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("discount not found")
		}
		return err
	}

	discount.IsActive = false
	if err := s.db.Save(&discount).Error; err != nil {
		return err
	}

	return nil
}

func (s *discountService) ValidateDiscountCode(code string, eventID string, orderValue float64) (*tickets.DiscountCode, error) {
	var discount tickets.DiscountCode
	if err := s.db.Where("code = ? AND is_active = true", code).First(&discount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired discount code")
		}
		return nil, err
	}

	now := time.Now()
	if now.Before(discount.ValidFrom) {
		return nil, errors.New("discount code is not yet valid")
	}
	if now.After(discount.ValidUntil) {
		return nil, errors.New("discount code has expired")
	}
	if discount.MaxUses > 0 && discount.CurrentUses >= discount.MaxUses {
		return nil, errors.New("discount code has reached maximum uses")
	}
	if discount.EventID != "" && discount.EventID != eventID {
		return nil, errors.New("discount code is not valid for this event")
	}
	if discount.MinOrderValue > 0 && orderValue < discount.MinOrderValue {
		return nil, errors.New("order value is below the minimum required for this discount")
	}

	return &discount, nil
}

func (s *discountService) mapToDiscountResponse(discount *tickets.DiscountCode) *DiscountResponse {
	return &DiscountResponse{
		ID:            discount.ID,
		Code:          discount.Code,
		EventID:       discount.EventID,
		Event:         discount.Event,
		DiscountType:  discount.DiscountType,
		DiscountValue: discount.DiscountValue,
		ValidFrom:     discount.ValidFrom,
		ValidUntil:    discount.ValidUntil,
		MaxUses:       discount.MaxUses,
		CurrentUses:   discount.CurrentUses,
		IsActive:      discount.IsActive,
		Source:        discount.Source,
		PromoterID:    discount.PromoterID,
		MinOrderValue: discount.MinOrderValue,
		IsSingleUse:   discount.IsSingleUse,
		OrganizerID:   discount.OrganizerID,
		Organizer:     discount.Organizer,
		CreatedAt:     discount.CreatedAt,
	}
}
