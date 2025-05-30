// In tickets/models/discount_codes.go
package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountPercentage  DiscountType = "percentage"
	DiscountFixedAmount DiscountType = "fixed_amount"
)

type DiscountSource string

const (
	DiscountSourceOrganizer DiscountSource = "organizer"
	DiscountSourcePromo     DiscountSource = "promo"
)

type DiscountCode struct {
	ID          string `gorm:"type:char(36);primaryKey" json:"id"`
	OrganizerID string `gorm:"type:char(36);not null;index" json:"organizer_id"`

	Code          string         `gorm:"size:50;not null;unique" json:"code"`
	EventID       string         `gorm:"type:char(36);index" json:"event_id"`
	DiscountType  DiscountType   `gorm:"size:20;not null" json:"discount_type"`
	DiscountValue float64        `gorm:"type:numeric(10,2);not null" json:"discount_value"`
	ValidFrom     time.Time      `gorm:"autoCreateTime" json:"valid_from"`
	ValidUntil    time.Time      `gorm:"" json:"valid_until"`
	MaxUses       int            `gorm:"default:0" json:"max_uses"`
	CurrentUses   int            `gorm:"default:0" json:"current_uses"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	Source        DiscountSource `gorm:"size:20;not null;default:'organizer'" json:"source"`
	PromoterID    string         `gorm:"type:char(36);index" json:"promoter_id,omitempty"`
	MinOrderValue float64        `gorm:"type:numeric(10,2);default:0" json:"min_order_value"`
	IsSingleUse   bool           `gorm:"default:false" json:"is_single_use"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version       int            `gorm:"default:1" json:"version"`

	// Relationships
	Event     events.Event         `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"event"`
	Tickets   []Ticket             `gorm:"foreignKey:DiscountCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"tickets"`
	Organizer organizers.Organizer `gorm:"foreignKey:OrganizerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organizer"`
	Promoter  members.User         `gorm:"foreignKey:PromoterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"promoter,omitempty"`
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
	if dc.ValidUntil.Before(dc.ValidFrom) {
		return errors.New("valid_until must be after valid_from")
	}
	if dc.Source == DiscountSourcePromo && dc.PromoterID == "" {
		return errors.New("promoter_id is required for promo source discounts")
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
	if dc.ValidUntil.Before(dc.ValidFrom) {
		return errors.New("valid_until must be after valid_from")
	}
	return nil
}

func (DiscountCode) TableName() string {
	return "discount_codes"
}
