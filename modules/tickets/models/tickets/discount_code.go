package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountPercentage  DiscountType = "percentage"
	DiscountFixedAmount DiscountType = "fixed_amount"
)

type DiscountCode struct {
	ID            string         `gorm:"type:char(36);primaryKey" json:"id"`
	Code          string         `gorm:"size:50;not null;unique" json:"code"`
	EventID       string         `gorm:"type:char(36);index" json:"event_id"`
	DiscountType  DiscountType   `gorm:"size:20;not null" json:"discount_type"`
	DiscountValue float64        `gorm:"type:numeric(10,2);not null" json:"discount_value"`
	ValidFrom     time.Time      `gorm:"autoCreateTime" json:"valid_from"`
	ValidUntil    time.Time      `gorm:"" json:"valid_until"`
	MaxUses       int            `gorm:"default:0" json:"max_uses"`
	CurrentUses   int            `gorm:"default:0" json:"current_uses"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version       int            `gorm:"default:1" json:"version"`

	// Relationships
	Event   events.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"event"`
	Tickets []Ticket     `gorm:"foreignKey:DiscountCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"tickets"`
}

func (dc *DiscountCode) BeforeCreate(tx *gorm.DB) (err error) {
	if dc.ID == "" {
		dc.ID = uuid.New().String()
	}
	if dc.Code == "" {
		return errors.New("code cannot be empty")
	}
	if dc.DiscountType != DiscountPercentage && dc.DiscountType != DiscountFixedAmount {
		return errors.New("discount_type must be 'percentage' or 'fixed_amount'")
	}
	if dc.DiscountValue < 0 {
		return errors.New("discount_value cannot be negative")
	}
	if dc.MaxUses < 0 {
		return errors.New("max_uses cannot be negative")
	}
	if dc.CurrentUses < 0 {
		return errors.New("current_uses cannot be negative")
	}
	if dc.CurrentUses > dc.MaxUses && dc.MaxUses != 0 {
		return errors.New("current_uses cannot exceed max_uses")
	}
	return nil
}

func (dc *DiscountCode) BeforeUpdate(tx *gorm.DB) (err error) {
	if dc.Code == "" {
		return errors.New("code cannot be empty")
	}
	if dc.DiscountType != DiscountPercentage && dc.DiscountType != DiscountFixedAmount {
		return errors.New("discount_type must be 'percentage' or 'fixed_amount'")
	}
	if dc.DiscountValue < 0 {
		return errors.New("discount_value cannot be negative")
	}
	if dc.MaxUses < 0 {
		return errors.New("max_uses cannot be negative")
	}
	if dc.CurrentUses < 0 {
		return errors.New("current_uses cannot be negative")
	}
	if dc.CurrentUses > dc.MaxUses && dc.MaxUses != 0 {
		return errors.New("current_uses cannot exceed max_uses")
	}
	return nil
}

func (DiscountCode) TableName() string {
	return "discount_codes"
}
