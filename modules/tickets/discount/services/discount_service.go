package discount_service

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/tickets/discount/dto"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountService interface {
	CreateDiscount(userID string, discount *dto.CreateDiscountCodeInput) (*dto.DiscountResponse, error)
	GetDiscount(userID, id string) (*dto.DiscountResponse, error)
	GetDiscounts(userID string) (*dto.GetDiscountsOutput, error)
	UpdateDiscount(userID, id string, discount *dto.UpdateDiscountCodeInput) (*dto.DiscountResponse, error)
	CancelDiscount(userID, id string) error
	ValidateDiscountCode(code string, eventID string, orderValue float64) (*tickets.DiscountCode, error)
}

type discountService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
}

func NewDiscountService(db *gorm.DB, authService authorization_service.PermissionService) DiscountService {
	return &discountService{
		db:                   db,
		authorizationService: authService,
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

func (s *discountService) CreateDiscount(userID string, input *dto.CreateDiscountCodeInput) (*dto.DiscountResponse, error) {
	// hasPerm, err := s.HasPermission(userID, "create:discounts")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:discounts permission")
	// }

	organizer, err := s.getUserOrganizer(userID)
	if err != nil {
		return nil, err
	}

	if input.EventID != "" {
		if _, err := uuid.Parse(input.EventID); err != nil {
			return nil, errors.New("invalid event ID format")
		}
		var event events.Event
		if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", input.EventID, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("event not found or not owned by organizer")
			}
			return nil, err
		}
	}

	if input.PromoterID != "" {
		if _, err := uuid.Parse(input.PromoterID); err != nil {
			return nil, errors.New("invalid promoter ID format")
		}
	}

	discount := &tickets.DiscountCode{
		OrganizerID:   organizer.ID,
		Code:          input.Code,
		EventID:       input.EventID,
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
		ValidFrom:     input.ValidFrom.UTC(),
		ValidUntil:    input.ValidUntil.UTC(),
		MaxUses:       input.MaxUses,
		CurrentUses:   0,
		IsActive:      input.IsActive,
		Source:        input.Source,
		PromoterID:    input.PromoterID,
		MinOrderValue: input.MinOrderValue,
		IsSingleUse:   input.IsSingleUse,
	}

	if err := s.db.Create(discount).Error; err != nil {
		return nil, err
	}

	return s.mapToDiscountResponse(discount), nil
}

func (s *discountService) GetDiscount(userID, id string) (*dto.DiscountResponse, error) {
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

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid discount ID format")
	}

	var discount tickets.DiscountCode
	if err := s.db.Preload("Event").Preload("Organizer").
		Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).
		First(&discount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("discount not found")
		}
		return nil, err
	}

	return s.mapToDiscountResponse(&discount), nil
}

func (s *discountService) GetDiscounts(userID string) (*dto.GetDiscountsOutput, error) {
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
		Where("organizer_id = ? AND deleted_at IS NULL", organizer.ID).
		Find(&discounts).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.DiscountResponse, len(discounts))
	for i, discount := range discounts {
		responses[i] = *s.mapToDiscountResponse(&discount)
	}

	return &dto.GetDiscountsOutput{Discounts: responses}, nil
}

func (s *discountService) UpdateDiscount(userID, id string, input *dto.UpdateDiscountCodeInput) (*dto.DiscountResponse, error) {
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

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid discount ID format")
	}

	var existingDiscount tickets.DiscountCode
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&existingDiscount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("discount not found")
		}
		return nil, err
	}

	if input.EventID != "" {
		if _, err := uuid.Parse(input.EventID); err != nil {
			return nil, errors.New("invalid event ID format")
		}
		var event events.Event
		if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", input.EventID, organizer.ID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("event not found or not owned by organizer")
			}
			return nil, err
		}
	}

	if input.PromoterID != "" {
		if _, err := uuid.Parse(input.PromoterID); err != nil {
			return nil, errors.New("invalid promoter ID format")
		}
	}

	discount := &tickets.DiscountCode{
		ID:            existingDiscount.ID,
		OrganizerID:   existingDiscount.OrganizerID,
		Code:          input.Code,
		EventID:       input.EventID,
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
		ValidFrom:     input.ValidFrom.UTC(),
		ValidUntil:    input.ValidUntil.UTC(),
		MaxUses:       input.MaxUses,
		CurrentUses:   existingDiscount.CurrentUses,
		IsActive:      input.IsActive,
		Source:        existingDiscount.Source,
		PromoterID:    existingDiscount.PromoterID,
		MinOrderValue: input.MinOrderValue,
		IsSingleUse:   input.IsSingleUse,
	}

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

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid discount ID format")
	}

	var discount tickets.DiscountCode
	if err := s.db.Where("id = ? AND organizer_id = ? AND deleted_at IS NULL", id, organizer.ID).First(&discount).Error; err != nil {
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
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var discount tickets.DiscountCode
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("code = ? AND is_active = true AND deleted_at IS NULL", code).
		First(&discount).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired discount code")
		}
		return nil, err
	}

	now := time.Now().UTC()
	if now.Before(discount.ValidFrom) {
		tx.Rollback()
		return nil, errors.New("discount code is not yet valid")
	}
	if now.After(discount.ValidUntil) {
		tx.Rollback()
		return nil, errors.New("discount code has expired")
	}
	if discount.MaxUses > 0 && discount.CurrentUses >= discount.MaxUses {
		tx.Rollback()
		return nil, errors.New("discount code has reached maximum uses")
	}
	if discount.EventID != "" && discount.EventID != eventID {
		tx.Rollback()
		return nil, errors.New("discount code is not valid for this event")
	}
	if discount.MinOrderValue > 0 && orderValue < discount.MinOrderValue {
		tx.Rollback()
		return nil, errors.New("order value is below the minimum required for this discount")
	}

	discount.CurrentUses++
	if err := tx.Save(&discount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &discount, nil
}

func (s *discountService) mapToDiscountResponse(discount *tickets.DiscountCode) *dto.DiscountResponse {
	return &dto.DiscountResponse{
		ID:            discount.ID,
		OrganizerID:   discount.OrganizerID,
		Code:          discount.Code,
		EventID:       discount.EventID,
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
		CreatedAt:     discount.CreatedAt,
		UpdatedAt:     discount.UpdatedAt,
	}
}
