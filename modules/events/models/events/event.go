package events

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"time"
)

type EventStatus string

const (
	EventActive   EventStatus = "active"
	EventInactive EventStatus = "inactive"
)

type Event struct {
	ID             string         `gorm:"type:char(36);primaryKey" json:"id"`
	Title          string         `gorm:"size:255;not null;index" json:"title"`
	CategoryID     int            `gorm:"not null;index" json:"category_id"`
	SubcategoryID  int            `gorm:"not null;index" json:"subcategory_id"`
	Description    string         `gorm:"type:text" json:"description"`
	VenueID        string         `gorm:"type:char(36);index" json:"venue_id"`
	TotalSeats     int            `gorm:"not null" json:"total_seats"`
	AvailableSeats int            `gorm:"not null" json:"available_seats"`
	StartTime      time.Time      `gorm:"not null;index" json:"start_time"`
	EndTime        time.Time      `gorm:"not null" json:"end_time"`
	OrganizerID    int            `gorm:"not null;index" json:"organizer_id"`
	PriceTierID    int            `gorm:"not null;index" json:"price_tier_id"`
	BasePrice      float64        `gorm:"type:numeric(10,2);not null;check:base_price >= 0" json:"base_price"`
	IsFeatured     bool           `gorm:"default:false;index" json:"is_featured"`
	Status         EventStatus    `gorm:"size:20;not null;default:'active';index" json:"status"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version        int            `gorm:"default:1" json:"version"`

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
	return nil
}

func (Event) TableName() string {
	return "events"
}
