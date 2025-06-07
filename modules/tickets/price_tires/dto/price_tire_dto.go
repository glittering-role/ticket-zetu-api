package dto

import "time"

type CreatePriceTierRequest struct {
	OrganizerID        string     `json:"organizer_id" binding:"required,uuid4"`
	Name               string     `json:"name" binding:"required,min=2,max=50"`
	Description        string     `json:"description,omitempty" binding:"omitempty,max=1000"`
	BasePrice          float64    `json:"base_price" binding:"required,gte=0"`
	PercentageIncrease float64    `json:"percentage_increase" binding:"required,gte=0"`
	IsDefault          bool       `json:"is_default" binding:"omitempty"`
	EffectiveFrom      time.Time  `json:"effective_from" binding:"required"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty" binding:"omitempty,gtfield=EffectiveFrom"`
	MinTickets         int        `json:"min_tickets" binding:"gte=0"`
	MaxTickets         *int       `json:"max_tickets,omitempty" binding:"omitempty,gte=0"`
}

type UpdatePriceTierRequest struct {
	Name               *string    `json:"name,omitempty" binding:"omitempty,min=2,max=50"`
	Description        *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	BasePrice          *float64   `json:"base_price,omitempty" binding:"omitempty,gte=0"`
	PercentageIncrease *float64   `json:"percentage_increase,omitempty" binding:"omitempty,gte=0"`
	IsDefault          *bool      `json:"is_default,omitempty"`
	EffectiveFrom      *time.Time `json:"effective_from,omitempty"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty" binding:"omitempty,gtfield=EffectiveFrom"`
	MinTickets         *int       `json:"min_tickets,omitempty" binding:"omitempty,gte=0"`
	MaxTickets         *int       `json:"max_tickets,omitempty" binding:"omitempty,gte=0"`
	Status             *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive archived"`
}

type GetPriceTierResponse struct {
	ID                 string     `json:"id"`
	OrganizerID        string     `json:"organizer_id"`
	Name               string     `json:"name"`
	Description        string     `json:"description,omitempty"`
	BasePrice          float64    `json:"base_price"`
	PercentageIncrease float64    `json:"percentage_increase"`
	Status             string     `json:"status"`
	IsDefault          bool       `json:"is_default"`
	EffectiveFrom      time.Time  `json:"effective_from"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty"`
	MinTickets         int        `json:"min_tickets"`
	MaxTickets         *int       `json:"max_tickets,omitempty"`
	Version            int        `json:"version"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
