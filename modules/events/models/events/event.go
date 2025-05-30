package events

import (
	"errors"
	"ticket-zetu-api/modules/events/models/categories"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventActive   EventStatus = "active"
	EventInactive EventStatus = "inactive"
)

type Event struct {
	ID             string                 `gorm:"type:char(36);primaryKey" json:"id"`
	Title          string                 `gorm:"size:255;not null;index" json:"title"`
	SubcategoryID  string                 `gorm:"not null;index" json:"subcategory_id"`
	Subcategory    categories.Subcategory `gorm:"foreignKey:SubcategoryID"`
	Description    string                 `gorm:"type:text" json:"description"`
	VenueID        string                 `gorm:"type:char(36);index" json:"venue_id"`
	TotalSeats     int                    `gorm:"not null" json:"total_seats"`
	AvailableSeats int                    `gorm:"not null" json:"available_seats"`
	StartTime      time.Time              `gorm:"not null;index" json:"start_time"`
	EndTime        time.Time              `gorm:"not null" json:"end_time"`
	OrganizerID    string                 `gorm:"type:char(36);not null;index" json:"-"`
	PriceTierID    string                 `gorm:"type:char(36);index" json:"price_tier_id"`
	BasePrice      float64                `gorm:"type:numeric(10,2);not null;check:base_price >= 0" json:"base_price"`
	IsFeatured     bool                   `gorm:"default:false;index" json:"is_featured"`
	Status         EventStatus            `gorm:"size:20;not null;default:'active';index" json:"status"`
	CreatedAt      time.Time              `gorm:"autoCreateTime" json:"-"`
	UpdatedAt      time.Time              `gorm:"autoUpdateTime" json:"-"`
	DeletedAt      gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`
	Version        int                    `gorm:"default:1" json:"-"`

	// Relationships
	Venue       Venue        `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"venue"`
	EventImages []EventImage `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"event_images"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.BasePrice < 0 {
		return errors.New("base_price cannot be negative")
	}
	if e.TotalSeats <= 0 {
		return errors.New("total_seats must be positive")
	}
	if e.AvailableSeats < 0 || e.AvailableSeats > e.TotalSeats {
		return errors.New("available_seats must be non-negative and not exceed total_seats")
	}
	if _, err := uuid.Parse(e.OrganizerID); err != nil {
		return errors.New("invalid organizer_id format")
	}
	return nil
}

func (e *Event) BeforeUpdate(tx *gorm.DB) (err error) {
	if e.BasePrice < 0 {
		return errors.New("base_price cannot be negative")
	}
	if e.TotalSeats <= 0 {
		return errors.New("total_seats must be positive")
	}
	if e.AvailableSeats < 0 || e.AvailableSeats > e.TotalSeats {
		return errors.New("available_seats must be non-negative and not exceed total_seats")
	}
	if _, err := uuid.Parse(e.OrganizerID); err != nil {
		return errors.New("invalid organizer_id format")
	}

	return nil
}

type EventImage struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	EventID   string         `gorm:"type:char(36);not null;index" json:"event_id"`
	ImageURL  string         `gorm:"size:255;not null" json:"image_url"`
	IsPrimary bool           `gorm:"default:false;index" json:"is_primary"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int            `gorm:"default:1" json:"version"`

	// Relationship
	Event Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (ei *EventImage) BeforeCreate(tx *gorm.DB) (err error) {
	if ei.ID == "" {
		ei.ID = uuid.New().String()
	}
	return nil
}

func (EventImage) TableName() string {
	return "event_images"
}

func (Event) TableName() string {
	return "events"
}
