package tickets

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TicketType struct {
	ID                string         `gorm:"type:char(36);primaryKey" json:"id"`
	EventID           string         `gorm:"type:char(36);not null" json:"event_id"`
	Name              string         `gorm:"size:100;not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description"`
	PriceModifier     float64        `gorm:"type:numeric(5,2);not null;default:1.0" json:"price_modifier"`
	Benefits          string         `gorm:"type:text" json:"benefits"`
	MaxTicketsPerUser int            `gorm:"not null;default:10" json:"max_tickets_per_user"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version           int            `gorm:"default:1" json:"version"`
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) (err error) {
	if tt.ID == "" {
		tt.ID = uuid.New().String()
	}
	// Validate required fields
	if tt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if tt.PriceModifier < 0 {
		return errors.New("price_modifier cannot be negative")
	}
	if tt.MaxTicketsPerUser < 1 {
		return errors.New("max_tickets_per_user must be at least 1")
	}
	return nil
}

func (tt *TicketType) BeforeUpdate(tx *gorm.DB) (err error) {
	// Validate required fields
	if tt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if tt.PriceModifier < 0 {
		return errors.New("price_modifier cannot be negative")
	}
	if tt.MaxTicketsPerUser < 1 {
		return errors.New("max_tickets_per_user must be at least 1")
	}
	return nil
}

func (TicketType) TableName() string {
	return "ticket_types"
}
