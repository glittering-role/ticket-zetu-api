package venue_dto

import (
	"ticket-zetu-api/modules/events/models/events"
	"time"
)

type VenueResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Address     string              `json:"address"`
	City        string              `json:"city"`
	State       string              `json:"state,omitempty"`
	Country     string              `json:"country"`
	Capacity    int                 `json:"capacity"`
	ContactInfo string              `json:"contact_info,omitempty"`
	Latitude    float64             `json:"latitude"`
	Longitude   float64             `json:"longitude"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	VenueImages []events.VenueImage `json:"venue_images,omitempty"`
}

type CreateVenueDto struct {
	Name        string  `form:"name" validate:"required,min=2,max=255" example:"Nairobi Arena"`
	Description string  `form:"description" validate:"max=1000" example:"A multi-purpose indoor venue for concerts, sports, and conferences."`
	Address     string  `form:"address" validate:"required" example:"123 Arena Road"`
	City        string  `form:"city" validate:"required,min=2,max=100" example:"Nairobi"`
	State       string  `form:"state" validate:"max=100" example:"Nairobi County"`
	Country     string  `form:"country" validate:"required,min=2,max=100" example:"Kenya"`
	Capacity    int     `form:"capacity" validate:"gte=0" example:"5000"`
	ContactInfo string  `form:"contact_info" validate:"max=255" example:"+254712345678"`
	Latitude    float64 `form:"latitude" example:"-1.2921"`
	Longitude   float64 `form:"longitude" example:"36.8219"`
	Status      string  `form:"status" validate:"oneof=active inactive suspended" example:"active"`
}

type UpdateVenueDto struct {
	Name        string  `form:"name" validate:"required,min=2,max=255"`
	Description string  `form:"description" validate:"max=1000"`
	Address     string  `form:"address" validate:"required"`
	City        string  `form:"city" validate:"required,min=2,max=100"`
	State       string  `form:"state" validate:"max=100"`
	Country     string  `form:"country" validate:"required,min=2,max=100"`
	Capacity    int     `form:"capacity" validate:"gte=0"`
	ContactInfo string  `form:"contact_info" validate:"max=255"`
	Latitude    float64 `form:"latitude"`
	Longitude   float64 `form:"longitude"`
	Status      string  `form:"status" validate:"oneof=active inactive suspended"`
}
