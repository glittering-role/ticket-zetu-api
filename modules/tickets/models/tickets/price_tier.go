// models/price_tier.go
package tickets

import (
	"errors"
	organizers "ticket-zetu-api/modules/organizers/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceTierStatus string

const (
	PriceTierActive   PriceTierStatus = "active"
	PriceTierInactive PriceTierStatus = "inactive"
	PriceTierArchived PriceTierStatus = "archived"
)

type PriceTier struct {
	ID            string          `gorm:"type:char(36);primaryKey" json:"id"`
	OrganizerID   string          `gorm:"type:char(36);not null;index" json:"organizer_id"`
	Name          string          `gorm:"size:50;not null" json:"name"`
	Description   string          `gorm:"type:text" json:"description"`
	BasePrice     float64         `gorm:"type:numeric(10,2);not null;check:base_price >= 0" json:"base_price"`
	Status        PriceTierStatus `gorm:"type:varchar(20);default:'active';check:status IN ('active','inactive','archived')" json:"status"`
	IsDefault     bool            `gorm:"default:false" json:"is_default"`
	EffectiveFrom time.Time       `gorm:"type:timestamp" json:"effective_from"`
	EffectiveTo   *time.Time      `gorm:"type:timestamp;null" json:"effective_to"`
	MinTickets    int             `gorm:"default:0" json:"min_tickets"`
	MaxTickets    *int            `gorm:"null" json:"max_tickets"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	Version       int             `gorm:"default:1" json:"version"`

	// Relationships
	Organizer organizers.Organizer `gorm:"foreignKey:OrganizerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organizer"`
}

func (pt *PriceTier) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	if pt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if pt.EffectiveTo != nil && pt.EffectiveFrom.After(*pt.EffectiveTo) {
		return errors.New("effective_from cannot be after effective_to")
	}
	if pt.MaxTickets != nil && *pt.MaxTickets < pt.MinTickets {
		return errors.New("max_tickets cannot be less than min_tickets")
	}
	return nil
}

func (pt *PriceTier) BeforeUpdate(tx *gorm.DB) (err error) {
	return pt.BeforeCreate(tx)
}

func (PriceTier) TableName() string {
	return "price_tiers"
}
