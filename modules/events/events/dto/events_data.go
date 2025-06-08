package dto

import "time"

type CreateEventInput struct {
	Title         string `json:"title" example:"Summer Music Festival" validate:"required,min=3,max=100"`
	Description   string `json:"description,omitempty" example:"An open-air festival with live performances, food, and fun."`
	SubcategoryID string `json:"subcategory_id" example:"c7e1249f-fc03-eb9d-ed90-8c236bd1996d" validate:"required,uuid"`
	VenueID       string `json:"venue_id" example:"a1d3c4e6-89ab-44ce-8e65-123456789abc" validate:"required,uuid"`

	StartTime time.Time `json:"start_time" example:"2025-08-15T18:00:00Z" validate:"required"`
	EndTime   time.Time `json:"end_time" example:"2025-08-15T23:00:00Z" validate:"required,gtfield=StartTime"`
	Timezone  string    `json:"timezone,omitempty" example:"Africa/Nairobi"` // No need to validate oneof if internal doesn't enforce it
	Language  string    `json:"language,omitempty" example:"en"`

	EventType      string `json:"event_type" example:"offline" validate:"required,oneof=online offline hybrid"`
	MinAge         int    `json:"min_age,omitempty" example:"18"`
	TotalSeats     int    `json:"total_seats" example:"500" validate:"required,gte=1"`
	AvailableSeats int    `json:"available_seats,omitempty" example:"480" validate:"gte=0,ltefield=TotalSeats"`

	IsFree     bool   `json:"is_free" example:"false"`
	HasTickets bool   `json:"has_tickets" example:"true"`
	IsFeatured bool   `json:"is_featured" example:"true"`
	Status     string `json:"status,omitempty" example:"active"` // optional in both

	Tags []string `json:"tags,omitempty" example:"[\"music\", \"festival\", \"live\"]"` // match to slice instead of comma-separated string
}
