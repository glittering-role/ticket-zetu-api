package dto

import (
	"ticket-zetu-api/modules/tickets/models/tickets"
	"time"
)

// CreateDiscountCodeInput defines the input structure for creating a discount code
type CreateDiscountCodeInput struct {
	Code          string                 `json:"code" binding:"required" example:"SUMMER2025"`
	EventID       string                 `json:"event_id" example:"3c6d4e3b-8c1d-4a2a-bd97-abcdef123456"`
	DiscountType  tickets.DiscountType   `json:"discount_type" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	DiscountValue float64                `json:"discount_value" binding:"required,gt=0" example:"15.5"`
	ValidFrom     time.Time              `json:"valid_from" binding:"required" example:"2025-06-20T00:00:00Z"`
	ValidUntil    time.Time              `json:"valid_until" binding:"required,gtfield=ValidFrom" example:"2025-07-20T23:59:59Z"`
	MaxUses       int                    `json:"max_uses" binding:"gte=0" example:"100"`
	CurrentUses   int                    `json:"current_uses" binding:"gte=0" example:"0"`
	IsActive      bool                   `json:"is_active" binding:"required" example:"true"`
	Source        tickets.DiscountSource `json:"source" binding:"required,oneof=organizer promo" example:"organizer"`
	PromoterID    string                 `json:"promoter_id,omitempty" example:"22bb86ce-89af-4e2d-9a47-00aa12345678"`
	MinOrderValue float64                `json:"min_order_value" binding:"gte=0" example:"500.00"`
	IsSingleUse   bool                   `json:"is_single_use" binding:"required" example:"false"`
}

// UpdateDiscountCodeInput defines the input structure for updating a discount code
type UpdateDiscountCodeInput struct {
	Code          string                 `json:"code" binding:"required"`
	EventID       string                 `json:"event_id"`
	DiscountType  tickets.DiscountType   `json:"discount_type" binding:"required,oneof=percentage fixed_amount"`
	DiscountValue float64                `json:"discount_value" binding:"required,gt=0"`
	ValidFrom     time.Time              `json:"valid_from" binding:"required"`
	ValidUntil    time.Time              `json:"valid_until" binding:"required,gtfield=ValidFrom"`
	MaxUses       int                    `json:"max_uses" binding:"gte=0"`
	CurrentUses   int                    `json:"current_uses" binding:"gte=0"`
	IsActive      bool                   `json:"is_active" binding:"required"`
	Source        tickets.DiscountSource `json:"source" binding:"required,oneof=organizer promo"`
	PromoterID    string                 `json:"promoter_id,omitempty"`
	MinOrderValue float64                `json:"min_order_value" binding:"gte=0"`
	IsSingleUse   bool                   `json:"is_single_use" binding:"required"`
}

// DiscountResponse defines the response structure for a single discount code
type DiscountResponse struct {
	ID            string                 `json:"id"`
	OrganizerID   string                 `json:"organizer_id"`
	Code          string                 `json:"code"`
	EventID       string                 `json:"event_id"`
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
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
}

// GetDiscountsOutput defines the response structure for fetching all discount codes
type GetDiscountsOutput struct {
	Discounts []DiscountResponse `json:"discounts"`
}
