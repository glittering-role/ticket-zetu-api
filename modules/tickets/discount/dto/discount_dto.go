package dto

import (
	"ticket-zetu-api/modules/tickets/models/tickets"
	"time"
)

type CreateDiscountCodeInput struct {
	OrganizerID   string                 `json:"organizer_id" binding:"required"`
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

type UpdateDiscountCodeInput struct {
	ID            string                 `json:"id" binding:"required"`
	OrganizerID   string                 `json:"organizer_id" binding:"required"`
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
