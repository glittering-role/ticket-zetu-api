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

type CreatePriceTierData struct {
	OrganizerID        string     `json:"organizer_id" example:"d5f9a7e3-3e4b-4c4d-8a44-1bdfcdb930f1" validate:"required,uuid"`
	Name               string     `json:"name" example:"Early Bird" validate:"required,min=2,max=50"`
	Description        string     `json:"description,omitempty" example:"Discounted price for early purchasers" validate:"omitempty,max=1000"`
	BasePrice          float64    `json:"base_price" example:"50.00" validate:"required,gte=0"`
	PercentageIncrease float64    `json:"percentage_increase" example:"10.0" validate:"required,gte=0"`
	IsDefault          bool       `json:"is_default" example:"false"`
	EffectiveFrom      time.Time  `json:"effective_from" example:"2025-07-01T00:00:00Z" validate:"required"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty" example:"2025-07-15T23:59:59Z" validate:"omitempty,gtfield=EffectiveFrom"`
	MinTickets         int        `json:"min_tickets" example:"0" validate:"gte=0"`
	MaxTickets         *int       `json:"max_tickets,omitempty" example:"100" validate:"omitempty,gte=0"`
}
