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
