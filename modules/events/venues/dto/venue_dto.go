package venue_dto

import (
	"ticket-zetu-api/modules/events/models/events"
	"time"
)

type Seat struct {
	SeatID      string `json:"seat_id" example:"A1"`
	Row         string `json:"row" example:"A"`
	Section     string `json:"section" example:"Main"`
	SeatNumber  int    `json:"seat_number" example:"1"`
	IsAvailable bool   `json:"is_available" example:"true"`
}

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

type CreateVenueDto struct {
	Name                  string  `form:"name" validate:"required,min=2,max=255" example:"Nairobi Arena"`
	Description           string  `form:"description" validate:"max=1000" example:"A multi-purpose indoor venue for concerts, sports, and conferences."`
	Address               string  `form:"address" validate:"required" example:"123 Arena Road"`
	City                  string  `form:"city" validate:"required,min=2,max=100" example:"Nairobi"`
	State                 string  `form:"state" validate:"max=100" example:"Nairobi County"`
	PostalCode            string  `form:"postal_code" validate:"max=20" example:"00100"`
	Country               string  `form:"country" validate:"required,min=2,max=100" example:"Kenya"`
	Capacity              int     `form:"capacity" validate:"gte=0" example:"5000"`
	VenueType             string  `form:"venue_type" validate:"required" example:"stadium"`
	Layout                string  `form:"layout" validate:"max=1000" example:"{\"seating\": \"tiered\", \"sections\": 4}"`
	AccessibilityFeatures string  `form:"accessibility_features" validate:"max=1000" example:"[\"wheelchair_ramp\", \"elevators\", \"braille_signage\"]"`
	Facilities            string  `form:"facilities" validate:"max=1000" example:"[\"restrooms\", \"parking\", \"concession_stands\"]"`
	ContactInfo           string  `form:"contact_info" validate:"max=255" example:"+254712345678"`
	Timezone              string  `form:"timezone" validate:"max=100" example:"Africa/Nairobi"`
	Latitude              float64 `form:"latitude" example:"-1.2921"`
	Longitude             float64 `form:"longitude" example:"36.8219"`
	Status                string  `form:"status" validate:"oneof=active inactive suspended" example:"active"`
}

type UpdateVenueDto struct {
	Name                  string  `form:"name" validate:"required,min=2,max=255"`
	Description           string  `form:"description" validate:"max=1000"`
	Address               string  `form:"address" validate:"required"`
	City                  string  `form:"city" validate:"required,min=2,max=100"`
	State                 string  `form:"state" validate:"max=100"`
	PostalCode            string  `form:"postal_code" validate:"max=20"`
	Country               string  `form:"country" validate:"required,min=2,max=100"`
	Capacity              int     `form:"capacity" validate:"gte=0"`
	VenueType             string  `form:"venue_type" validate:"required,oneof=stadium hotel park theater other"`
	Layout                string  `form:"layout" validate:"max=1000"`
	AccessibilityFeatures string  `form:"accessibility_features" validate:"max=1000"`
	Facilities            string  `form:"facilities" validate:"max=1000"`
	ContactInfo           string  `form:"contact_info" validate:"max=255"`
	Timezone              string  `form:"timezone" validate:"max=100"`
	Latitude              float64 `form:"latitude"`
	Longitude             float64 `form:"longitude"`
	Status                string  `form:"status" validate:"oneof=active inactive suspended"`
}
