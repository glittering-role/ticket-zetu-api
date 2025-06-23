package seats

import (
	"errors"
	//	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Seat struct {
	ID          string         `gorm:"type:char(36);primaryKey" json:"id"`
	VenueID     string         `gorm:"not null;index" json:"venue_id"`
	SeatNumber  string         `gorm:"type:varchar(10);not null" json:"seat_number"`
	SeatSection string         `gorm:"type:varchar(50)" json:"seat_section"`
	Status      string         `gorm:"type:varchar(20);not null;default:'available';check:status IN ('available','held','booked')" json:"status"`
	PriceTierID string         `gorm:"index" json:"price_tier_id,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version     int            `gorm:"default:1" json:"version"`

	PriceTier tickets.PriceTier `gorm:"foreignKey:PriceTierID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"price_tier,omitempty"`
	Venue     events.Venue      `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"venue"`
}

func (s *Seat) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.VenueID == "" {
		return errors.New("venue_id cannot be empty")
	}
	if s.SeatNumber == "" {
		return errors.New("seat_number cannot be empty")
	}
	return nil
}

func (s *Seat) BeforeUpdate(tx *gorm.DB) error {
	return s.BeforeCreate(tx)
}

func (Seat) TableName() string {
	return "seats"
}
