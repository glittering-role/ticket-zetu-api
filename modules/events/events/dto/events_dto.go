package dto

import (
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"time"
)

// Seat represents an individual seat in a venue
type Seat struct {
	SeatID      string `json:"seat_id" example:"A1"`
	Row         string `json:"row" example:"A"`
	Section     string `json:"section" example:"Main"`
	SeatNumber  int    `json:"seat_number" example:"1"`
	IsAvailable bool   `json:"is_available" example:"true"`
}

// ReservedSeat represents a seat reserved for an event
type ReservedSeat struct {
	ReservationID string     `json:"reservation_id" example:"uuid-1234"`
	SeatID        string     `json:"seat_id" example:"A1"`
	EventID       string     `json:"event_id" example:"uuid-5678"`
	UserID        string     `json:"user_id" example:"uuid-9012"`
	Status        string     `json:"status" example:"reserved"`
	ReservedAt    time.Time  `json:"reserved_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

// TicketTypeResponse contains essential fields for a ticket type
type TicketTypeResponse struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	PriceModifier     float64                  `json:"price_modifier"`
	Benefits          string                   `json:"benefits,omitempty"`
	MinTicketsPerUser int                      `json:"min_tickets_per_user"`
	MaxTicketsPerUser int                      `json:"max_tickets_per_user"`
	QuantityAvailable *int                     `json:"quantity_available,omitempty"`
	Status            tickets.TicketTypeStatus `json:"status"`
	IsDefault         bool                     `json:"is_default"`
	SalesStart        time.Time                `json:"sales_start"`
	SalesEnd          *time.Time               `json:"sales_end,omitempty"`
	PriceTiers        []PriceTierResponse      `json:"price_tiers,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

// PriceTierResponse contains essential fields for a price tier
type PriceTierResponse struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description,omitempty"`
	BasePrice     float64                 `json:"base_price"`
	Status        tickets.PriceTierStatus `json:"status"`
	IsDefault     bool                    `json:"is_default"`
	EffectiveFrom time.Time               `json:"effective_from"`
	EffectiveTo   *time.Time              `json:"effective_to,omitempty"`
	MinTickets    int                     `json:"min_tickets"`
	MaxTickets    *int                    `json:"max_tickets,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

type CreateEvent struct {
	Title          string                 `json:"title" validate:"required"`
	Description    string                 `json:"description,omitempty"`
	SubcategoryID  string                 `json:"subcategory_id" validate:"required"`
	Subcategory    categories.Subcategory `json:"subcategory,omitempty"`
	VenueID        string                 `json:"venue_id" validate:"required"`
	Venue          events.Venue           `json:"venue,omitempty"`
	StartTime      time.Time              `json:"start_time" validate:"required"`
	EndTime        time.Time              `json:"end_time" validate:"required"`
	Timezone       string                 `json:"timezone,omitempty"`
	Language       string                 `json:"language,omitempty"`
	EventType      string                 `json:"event_type" validate:"oneof=online offline hybrid"`
	MinAge         int                    `json:"min_age"`
	TotalSeats     int                    `json:"total_seats" validate:"required"`
	AvailableSeats int                    `json:"available_seats"`
	IsFree         bool                   `json:"is_free"`
	HasTickets     bool                   `json:"has_tickets"`
	IsFeatured     bool                   `json:"is_featured"`
	Status         string                 `json:"status,omitempty"`
	TicketTypes    []TicketTypeResponse   `json:"ticket_types,omitempty" validate:"dive"`
}

type UpdateEvent struct {
	Title          *string               `json:"title,omitempty"`
	Description    *string               `json:"description,omitempty"`
	SubcategoryID  *string               `json:"subcategory_id,omitempty"`
	VenueID        *string               `json:"venue_id,omitempty"`
	StartTime      *time.Time            `json:"start_time,omitempty"`
	EndTime        *time.Time            `json:"end_time,omitempty"`
	Timezone       *string               `json:"timezone,omitempty"`
	Language       *string               `json:"language,omitempty"`
	EventType      *string               `json:"event_type,omitempty"`
	MinAge         *int                  `json:"min_age,omitempty"`
	TotalSeats     *int                  `json:"total_seats,omitempty"`
	AvailableSeats *int                  `json:"available_seats,omitempty"`
	IsFree         *bool                 `json:"is_free,omitempty"`
	HasTickets     *bool                 `json:"has_tickets,omitempty"`
	IsFeatured     *bool                 `json:"is_featured,omitempty"`
	Status         *string               `json:"status,omitempty"`
	TicketTypes    *[]TicketTypeResponse `json:"ticket_types,omitempty" validate:"dive"`
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
type VenueResponse struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	Description           string              `json:"description,omitempty"`
	Address               string              `json:"address"`
	City                  string              `json:"city"`
	State                 string              `json:"state,omitempty"`
	PostalCode            string              `json:"postal_code,omitempty"`
	Country               string              `json:"country"`
	Capacity              int                 `json:"capacity"`
	VenueType             string              `json:"venue_type"`
	Layout                string              `json:"layout,omitempty"`
	AccessibilityFeatures string              `json:"accessibility_features,omitempty"`
	Facilities            string              `json:"facilities,omitempty"`
	ContactInfo           string              `json:"contact_info,omitempty"`
	Timezone              string              `json:"timezone,omitempty"`
	Latitude              float64             `json:"latitude"`
	Longitude             float64             `json:"longitude"`
	Status                string              `json:"status"`
	OrganizerID           string              `json:"organizer_id"`
	CreatedAt             time.Time           `json:"created_at"`
	VenueImages           []events.VenueImage `json:"venue_images,omitempty"`
	Seats                 []Seat              `json:"seats,omitempty"`
}

// EventResponse for single event retrieval with full details
type EventResponse struct {
	ID             string               `json:"id"`
	Title          string               `json:"title"`
	Slug           string               `json:"slug"`
	Description    string               `json:"description,omitempty"`
	SubcategoryID  string               `json:"subcategory_id"`
	Subcategory    SubcategoryResponse  `json:"subcategory"`
	VenueID        string               `json:"venue_id"`
	Venue          VenueResponse        `json:"venue"`
	Upvotes        int                  `json:"upvotes"`
	Downvotes      int                  `json:"downvotes"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	Timezone       string               `json:"timezone"`
	Language       string               `json:"language"`
	EventType      string               `json:"event_type"`
	MinAge         int                  `json:"min_age"`
	TotalSeats     int                  `json:"total_seats"`
	AvailableSeats int                  `json:"available_seats"`
	IsFree         bool                 `json:"is_free"`
	HasTickets     bool                 `json:"has_tickets"`
	IsFeatured     bool                 `json:"is_featured"`
	Status         string               `json:"status"`
	EventImages    []events.EventImage  `json:"event_images,omitempty"`
	PublishedAt    *time.Time           `json:"published_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	TicketTypes    []TicketTypeResponse `json:"ticket_types,omitempty"`
	ReservedSeats  []ReservedSeat       `json:"reserved_seats,omitempty"`
}

// MinimalEventResponse for listing multiple events
type MinimalEventResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Slug        string               `json:"slug"`
	StartTime   time.Time            `json:"start_time"`
	EndTime     time.Time            `json:"end_time"`
	Timezone    string               `json:"timezone"`
	EventType   string               `json:"event_type"`
	IsFree      bool                 `json:"is_free"`
	HasTickets  bool                 `json:"has_tickets"`
	Upvotes     int                  `json:"upvotes"`
	Downvotes   int                  `json:"downvotes"`
	IsFeatured  bool                 `json:"is_featured"`
	Status      string               `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	EventImages []events.EventImage  `json:"event_images,omitempty"`
	TicketTypes []TicketTypeResponse `json:"ticket_types,omitempty"`
}
