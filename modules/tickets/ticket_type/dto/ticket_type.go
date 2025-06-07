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
