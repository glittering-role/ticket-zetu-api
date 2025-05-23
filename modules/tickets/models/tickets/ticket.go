package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketValid    TicketStatus = "valid"
	TicketUsed     TicketStatus = "used"
	TicketCanceled TicketStatus = "canceled"
	TicketRefunded TicketStatus = "refunded"
	TicketPending  TicketStatus = "pending"
)

type Ticket struct {
	ID               string         `gorm:"type:char(36);primaryKey" json:"id"`
	TicketNumber     string         `gorm:"type:char(36);unique;not null" json:"ticket_number"`
	EventID          string         `gorm:"type:char(36);not null;index" json:"event_id"`
	UserID           string         `gorm:"type:char(36);not null;index" json:"user_id"`
	TicketTypeID     string         `gorm:"type:char(36);not null;index" json:"ticket_type_id"`
	SeatNumber       string         `gorm:"size:10" json:"seat_number"`
	SeatSection      string         `gorm:"size:50" json:"seat_section"`
	PurchaseTime     time.Time      `gorm:"autoCreateTime" json:"purchase_time"`
	Status           TicketStatus   `gorm:"size:20;not null;default:'valid'" json:"status"`
	QRCodeHash       string         `gorm:"type:text;not null" json:"qr_code_hash"`
	PaymentReference string         `gorm:"size:255;not null" json:"payment_reference"`
	PaymentMethod    string         `gorm:"size:50" json:"payment_method"`
	ActualPrice      float64        `gorm:"type:numeric(10,2);not null" json:"actual_price"`
	DiscountCode     string         `gorm:"size:50" json:"discount_code"`
	CheckedInAt      time.Time      `gorm:"" json:"checked_in_at"`
	CheckedInBy      string         `gorm:"type:char(36)" json:"checked_in_by"`
	Notes            string         `gorm:"type:text" json:"notes"`
	IsTransferable   bool           `gorm:"default:true" json:"is_transferable"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version          int            `gorm:"default:1" json:"version"`

	// Relationships
	Event      events.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"event"`
	User       members.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user"`
	TicketType TicketType   `gorm:"foreignKey:TicketTypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"ticket_type"`
	Discount   DiscountCode `gorm:"foreignKey:DiscountCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"discount"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.TicketNumber == "" {
		t.TicketNumber = uuid.New().String()
	}
	if t.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if t.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if t.TicketTypeID == "" {
		return errors.New("ticket_type_id cannot be empty")
	}
	if t.QRCodeHash == "" {
		return errors.New("qr_code_hash cannot be empty")
	}
	if t.PaymentReference == "" {
		return errors.New("payment_reference cannot be empty")
	}
	if t.ActualPrice < 0 {
		return errors.New("actual_price cannot be negative")
	}
	if t.Status != TicketValid && t.Status != TicketUsed && t.Status != TicketCanceled && t.Status != TicketRefunded && t.Status != TicketPending {
		return errors.New("status must be one of 'valid', 'used', 'canceled', 'refunded', 'pending'")
	}
	return nil
}

func (t *Ticket) BeforeUpdate(tx *gorm.DB) (err error) {
	if t.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if t.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if t.TicketTypeID == "" {
		return errors.New("ticket_type_id cannot be empty")
	}
	if t.QRCodeHash == "" {
		return errors.New("qr_code_hash cannot be empty")
	}
	if t.PaymentReference == "" {
		return errors.New("payment_reference cannot be empty")
	}
	if t.ActualPrice < 0 {
		return errors.New("actual_price cannot be negative")
	}
	if t.Status != TicketValid && t.Status != TicketUsed && t.Status != TicketCanceled && t.Status != TicketRefunded && t.Status != TicketPending {
		return errors.New("status must be one of 'valid', 'used', 'canceled', 'refunded', 'pending'")
	}
	return nil
}

func (Ticket) TableName() string {
	return "tickets"
}
