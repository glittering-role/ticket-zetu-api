package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceTier struct {
	ID                 string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name               string         `gorm:"size:50;not null" json:"name"`
	Description        string         `gorm:"type:text" json:"description"`
	PercentageIncrease float64        `gorm:"type:numeric(5,2);not null" json:"percentage_increase"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version            int            `gorm:"default:1" json:"version"`

	// Relationships
	Events []events.Event `gorm:"foreignKey:PriceTierID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"events"`
}

func (pt *PriceTier) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	if pt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if pt.PercentageIncrease < 0 {
		return errors.New("percentage_increase cannot be negative")
	}
	return nil
}

func (pt *PriceTier) BeforeUpdate(tx *gorm.DB) (err error) {
	if pt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if pt.PercentageIncrease < 0 {
		return errors.New("percentage_increase cannot be negative")
	}
	return nil
}

func (PriceTier) TableName() string {
	return "price_tiers"
}
