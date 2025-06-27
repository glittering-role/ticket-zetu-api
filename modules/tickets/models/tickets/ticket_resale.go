package tickets

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ResaleStatus string

const (
	ResaleListed    ResaleStatus = "listed"
	ResalePending   ResaleStatus = "pending"
	ResaleCompleted ResaleStatus = "completed"
	ResaleCanceled  ResaleStatus = "canceled"
	ResaleExpired   ResaleStatus = "expired"
)

type TicketResale struct {
	ID             string  `gorm:"type:char(36);primaryKey" json:"id"`
	TicketID       string  `gorm:"type:char(36);not null;index" json:"ticket_id"`
	OriginalUserID string  `gorm:"type:char(36);not null;index" json:"original_user_id"`
	NewUserID      *string `gorm:"type:char(36);index" json:"new_user_id"`

	// Pricing
	OriginalPrice float64 `gorm:"type:numeric(10,2);not null" json:"original_price"`
	ResalePrice   float64 `gorm:"type:numeric(10,2);not null" json:"resale_price"`
	PlatformFee   float64 `gorm:"type:numeric(10,2);not null" json:"platform_fee"`

	// Time controls
	ListedAt  time.Time  `gorm:"autoCreateTime" json:"listed_at"`
	SoldAt    *time.Time `gorm:"index" json:"sold_at"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`

	// Status
	Status ResaleStatus `gorm:"type:varchar(20);default:'listed'" json:"status"`

	// Anti-flipping rules
	MinHoldDuration time.Duration `gorm:"-" json:"min_hold_duration"`
	MinHoldDays     int           `gorm:"default:1" json:"min_hold_days"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Ticket Ticket `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"ticket"`
}

func (tr *TicketResale) BeforeCreate() error {
	if tr.ID == "" {
		tr.ID = uuid.New().String()
	}
	if tr.TicketID == "" {
		return errors.New("ticket_id cannot be empty")
	}
	if tr.OriginalUserID == "" {
		return errors.New("original_user_id cannot be empty")
	}
	if tr.ResalePrice <= 0 {
		return errors.New("resale_price must be greater than zero")
	}
	if tr.PlatformFee < 0 {
		return errors.New("platform_fee cannot be negative")
	}
	if tr.ExpiresAt.IsZero() {
		return errors.New("expires_at cannot be empty")
	}
	return nil
}

func (tr *TicketResale) BeforeUpdate() error {
	return tr.BeforeCreate()
}

func (TicketResale) TableName() string {
	return "ticket_resales"
}
