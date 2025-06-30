package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStock struct {
	ID           string `gorm:"type:char(36);primaryKey" json:"id"`
	TicketTypeID string `gorm:"type:char(36);not null;uniqueIndex" json:"ticket_type_id"`
	EventID      string `gorm:"type:char(36);not null;index" json:"event_id"`

	// Stock tracking
	TotalStock     int `gorm:"not null;default:0;check:total_stock >= 0" json:"total_stock"`
	AvailableStock int `gorm:"not null;default:0;check:available_stock >= 0 AND available_stock <= total_stock" json:"available_stock"`
	ReservedStock  int `gorm:"not null;default:0;check:reserved_stock >= 0 AND (reserved_stock + available_stock + held_stock + resale_stock) <= total_stock" json:"reserved_stock"`
	HeldStock      int `gorm:"not null;default:0;check:held_stock >= 0" json:"held_stock"`
	ResaleStock    int `gorm:"not null;default:0;check:resale_stock >= 0" json:"resale_stock"`

	// Holding configuration
	HoldDuration time.Duration `gorm:"-" json:"hold_duration"`
	HoldSeconds  int64         `gorm:"default:900" json:"hold_seconds"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	Version      int           `gorm:"default:1" json:"version"`

	// Relationships
	TicketType *TicketType  `gorm:"foreignKey:TicketTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"ticket_type"`
	Event      events.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"event"`
}

func (ts *TicketStock) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == "" {
		ts.ID = uuid.New().String()
	}

	// Convert duration to seconds for DB storage
	if ts.HoldDuration != 0 {
		ts.HoldSeconds = int64(ts.HoldDuration.Seconds())
	} else if ts.HoldSeconds == 0 {
		ts.HoldSeconds = 900 // Default 15 minutes
	}

	// Validate required fields
	if ts.TicketTypeID == "" {
		return errors.New("ticket_type_id cannot be empty")
	}
	if ts.EventID == "" {
		return errors.New("event_id cannot be empty")
	}

	// Validate stock values
	if ts.TotalStock < 0 {
		return errors.New("total_stock cannot be negative")
	}
	if ts.AvailableStock > ts.TotalStock {
		return errors.New("available_stock cannot exceed total_stock")
	}

	// Ensure stock sums are valid
	totalTracked := ts.AvailableStock + ts.ReservedStock + ts.HeldStock + ts.ResaleStock
	if totalTracked > ts.TotalStock {
		return errors.New("sum of stock allocations cannot exceed total_stock")
	}

	return nil
}

func (ts *TicketStock) AfterFind(tx *gorm.DB) error {
	ts.HoldDuration = time.Duration(ts.HoldSeconds) * time.Second
	return nil
}

func (ts *TicketStock) BeforeUpdate(tx *gorm.DB) error {
	return ts.BeforeCreate(tx)
}

func (TicketStock) TableName() string {
	return "ticket_stocks"
}
