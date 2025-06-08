package dto

import (
	"time"

	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
)

type CreateEvent struct {
	Title         string                 `json:"title" validate:"required"`
	Description   string                 `json:"description,omitempty"`
	SubcategoryID string                 `json:"subcategory_id" validate:"required"`
	Subcategory   categories.Subcategory `json:"subcategory,omitempty"`
	VenueID       string                 `json:"venue_id" validate:"required"`
	Venue         events.Venue           `json:"venue,omitempty"`

	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	Timezone  string    `json:"timezone,omitempty"`
	Language  string    `json:"language,omitempty"`

	EventType      string `json:"event_type" validate:"oneof=online offline hybrid"`
	MinAge         int    `json:"min_age"`
	TotalSeats     int    `json:"total_seats" validate:"required"`
	AvailableSeats int    `json:"available_seats"`

	IsFree     bool   `json:"is_free"`
	HasTickets bool   `json:"has_tickets"`
	IsFeatured bool   `json:"is_featured"`
	Status     string `json:"status,omitempty"`

	Tags []string `json:"tags,omitempty"`
}

type UpdateEvent struct {
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	SubcategoryID *string `json:"subcategory_id,omitempty"`
	VenueID       *string `json:"venue_id,omitempty"`

	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Timezone  *string    `json:"timezone,omitempty"`
	Language  *string    `json:"language,omitempty"`

	EventType      *string `json:"event_type,omitempty"`
	MinAge         *int    `json:"min_age,omitempty"`
	TotalSeats     *int    `json:"total_seats,omitempty"`
	AvailableSeats *int    `json:"available_seats,omitempty"`

	IsFree     *bool   `json:"is_free,omitempty"`
	HasTickets *bool   `json:"has_tickets,omitempty"`
	IsFeatured *bool   `json:"is_featured,omitempty"`
	Status     *string `json:"status,omitempty"`

	Tags *[]string `json:"tags,omitempty"`
}

type EventResponse struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description,omitempty"`
	SubcategoryID string                 `json:"subcategory_id"`
	Subcategory   categories.Subcategory `json:"subcategory"`
	VenueID       string                 `json:"venue_id"`
	Venue         events.Venue           `json:"venue"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timezone  string    `json:"timezone"`
	Language  string    `json:"language"`

	EventType      string `json:"event_type"`
	MinAge         int    `json:"min_age"`
	TotalSeats     int    `json:"total_seats"`
	AvailableSeats int    `json:"available_seats"`

	IsFree     bool   `json:"is_free"`
	HasTickets bool   `json:"has_tickets"`
	IsFeatured bool   `json:"is_featured"`
	Status     string `json:"status"`

	Tags        []string            `json:"tags,omitempty"`
	EventImages []events.EventImage `json:"event_images,omitempty"`
	PublishedAt *time.Time          `json:"published_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
