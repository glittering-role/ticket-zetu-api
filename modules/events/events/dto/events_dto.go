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
}

// SubcategoryResponse contains essential fields for a subcategory
type SubcategoryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ImageURL   string `json:"image_url"`
	CategoryID string `json:"category_id"`
	IsActive   bool   `json:"is_active"`
}

// VenueResponse contains essential fields for a venue
type VenueImage struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	VenueID   string     `json:"venue_id"`
	ImageURL  string     `json:"image_url"`
	IsPrimary bool       `json:"is_primary"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
type VenueResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Address     string              `json:"address"`
	City        string              `json:"city"`
	Country     string              `json:"country"`
	Capacity    int                 `json:"capacity"`
	Status      string              `json:"status"`
	VenueImages []events.VenueImage `json:"venue_images,omitempty"`
}

// EventResponse for single event retrieval with full details
type EventResponse struct {
	ID             string              `json:"id"`
	Title          string              `json:"title"`
	Slug           string              `json:"slug"`
	Description    string              `json:"description,omitempty"`
	SubcategoryID  string              `json:"subcategory_id"`
	Subcategory    SubcategoryResponse `json:"subcategory"`
	VenueID        string              `json:"venue_id"`
	Venue          VenueResponse       `json:"venue"`
	StartTime      time.Time           `json:"start_time"`
	EndTime        time.Time           `json:"end_time"`
	Timezone       string              `json:"timezone"`
	Language       string              `json:"language"`
	EventType      string              `json:"event_type"`
	MinAge         int                 `json:"min_age"`
	TotalSeats     int                 `json:"total_seats"`
	AvailableSeats int                 `json:"available_seats"`
	IsFree         bool                `json:"is_free"`
	HasTickets     bool                `json:"has_tickets"`
	IsFeatured     bool                `json:"is_featured"`
	Status         string              `json:"status"`
	EventImages    []events.EventImage `json:"event_images,omitempty"`
	PublishedAt    *time.Time          `json:"published_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// MinimalEventResponse for listing multiple events
type MinimalEventResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Slug        string              `json:"slug"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Timezone    string              `json:"timezone"`
	EventType   string              `json:"event_type"`
	IsFree      bool                `json:"is_free"`
	HasTickets  bool                `json:"has_tickets"`
	IsFeatured  bool                `json:"is_featured"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	EventImages []events.EventImage `json:"event_images,omitempty"`
}
