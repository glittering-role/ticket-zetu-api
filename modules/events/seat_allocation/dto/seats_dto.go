package dto

import "ticket-zetu-api/modules/tickets/models/tickets"

type GetSeatDTO struct {
	ID          string             `json:"id"`
	VenueID     string             `json:"venue_id"`
	SeatNumber  string             `json:"seat_number"`
	SeatSection string             `json:"seat_section"`
	Status      string             `json:"status"`
	PriceTierID string             `json:"price_tier_id,omitempty"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	DeletedAt   *string            `json:"deleted_at,omitempty"`
	VenueName   string             `json:"venue_name,omitempty"`
	PriceTier   *tickets.PriceTier `json:"price_tier,omitempty"`
}

type CreateSeatDTO struct {
	VenueID     string `json:"venue_id" binding:"required"`
	SeatNumber  string `json:"seat_number" binding:"required"`
	SeatSection string `json:"seat_section"`
	Status      string `json:"status" binding:"required,oneof=available held booked"`
	PriceTierID string `json:"price_tier_id,omitempty"`
}

type UpdateSeatDTO struct {
	ID          string `json:"id" binding:"required"`
	VenueID     string `json:"venue_id" binding:"required"`
	SeatNumber  string `json:"seat_number" binding:"required"`
	SeatSection string `json:"seat_section"`
	Status      string `json:"status" binding:"required,oneof=available held booked"`
	PriceTierID string `json:"price_tier_id,omitempty"`
}

type ToggleSeatStatusDTO struct {
	ID     string `json:"id" binding:"required"`
	Status string `json:"status" binding:"required,oneof=available held booked"`
}

type SeatFilterDTO struct {
	VenueID     string `json:"venue_id,omitempty"`
	SeatNumber  string `json:"seat_number,omitempty"`
	SeatSection string `json:"seat_section,omitempty"`
	Status      string `json:"status,omitempty" binding:"omitempty,oneof=available held booked"`
	PriceTierID string `json:"price_tier_id,omitempty"`
	Page        int    `json:"page,omitempty" binding:"omitempty,min=1"`
	PageSize    int    `json:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}
