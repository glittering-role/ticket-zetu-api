package dto

import (
	"ticket-zetu-api/modules/events/models/events"
	"time"
)

type TicketTypeResponse struct {
	ID                string       `json:"id"`
	EventID           string       `json:"event_id"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	PriceModifier     float64      `json:"price_modifier"`
	Benefits          string       `json:"benefits"`
	MaxTicketsPerUser int          `json:"max_tickets_per_user"`
	Status            string       `json:"status"`
	IsDefault         bool         `json:"is_default"`
	SalesStart        time.Time    `json:"sales_start"`
	SalesEnd          *time.Time   `json:"sales_end"`
	QuantityAvailable *int         `json:"quantity_available"`
	MinTicketsPerUser int          `json:"min_tickets_per_user"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Event             events.Event `json:"event,omitempty"`
}

type CreateTicketTypeInput struct {
	EventID           string     `json:"event_id" binding:"required"`
	Name              string     `json:"name" binding:"required"`
	Description       string     `json:"description"`
	PriceModifier     float64    `json:"price_modifier" binding:"required,gt=0"`
	Benefits          string     `json:"benefits"`
	MaxTicketsPerUser int        `json:"max_tickets_per_user" binding:"required,gte=1"`
	Status            string     `json:"status" binding:"required,oneof=active inactive archived"`
	IsDefault         bool       `json:"is_default"`
	SalesStart        time.Time  `json:"sales_start" binding:"required"`
	SalesEnd          *time.Time `json:"sales_end"`
	QuantityAvailable *int       `json:"quantity_available"`
	MinTicketsPerUser int        `json:"min_tickets_per_user" binding:"required,gte=1"`
}

type UpdateTicketTypeInput struct {
	EventID           string     `json:"event_id" binding:"required"`
	Name              string     `json:"name" binding:"required"`
	Description       string     `json:"description"`
	PriceModifier     float64    `json:"price_modifier" binding:"required,gt=0"`
	Benefits          string     `json:"benefits"`
	MaxTicketsPerUser int        `json:"max_tickets_per_user" binding:"required,gte=1"`
	Status            string     `json:"status" binding:"required,oneof=active inactive archived"`
	IsDefault         bool       `json:"is_default"`
	SalesStart        time.Time  `json:"sales_start" binding:"required"`
	SalesEnd          *time.Time `json:"sales_end"`
	QuantityAvailable *int       `json:"quantity_available"`
	MinTicketsPerUser int        `json:"min_tickets_per_user" binding:"required,gte=1"`
	ID                string     `json:"id" binding:"required"`
}

type CreateTicketTypeInputData struct {
	EventID           string     `json:"event_id" example:"a1d3c4e6-89ab-44ce-8e65-123456789abc" validate:"required,uuid"`
	Name              string     `json:"name" example:"VIP Ticket" validate:"required,min=2,max=100"`
	Description       string     `json:"description,omitempty" example:"Access to VIP lounge and priority entry"`
	PriceModifier     float64    `json:"price_modifier" example:"1.5" validate:"required,gt=0"`
	Benefits          string     `json:"benefits,omitempty" example:"Free drinks, front row seats"`
	MaxTicketsPerUser int        `json:"max_tickets_per_user" example:"5" validate:"required,gte=1"`
	Status            string     `json:"status" example:"active" validate:"required,oneof=active inactive archived"`
	IsDefault         bool       `json:"is_default" example:"false"`
	SalesStart        time.Time  `json:"sales_start" example:"2025-07-01T09:00:00Z" validate:"required"`
	SalesEnd          *time.Time `json:"sales_end,omitempty" example:"2025-08-01T23:59:59Z"`
	QuantityAvailable *int       `json:"quantity_available,omitempty" example:"100"`
	MinTicketsPerUser int        `json:"min_tickets_per_user" example:"1" validate:"required,gte=1"`
}
