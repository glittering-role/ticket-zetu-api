package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTicketLimits struct {
	ID      string `gorm:"type:char(36);primaryKey" json:"id"`
	UserID  string `gorm:"type:char(36);primaryKey" json:"user_id"`
	EventID string `gorm:"type:char(36);primaryKey" json:"event_id"`

	// Limits
	MaxTickets    int `gorm:"not null;default:4" json:"max_tickets"`
	TicketsBought int `gorm:"not null;default:0" json:"tickets_bought"`
	TicketsResold int `gorm:"not null;default:0" json:"tickets_resold"`

	// Dynamic pricing adjustments
	LastPurchaseAt *time.Time `gorm:"index" json:"last_purchase_at"`
	PurchaseCount  int        `gorm:"default:0" json:"purchase_count"`

	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User  members.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Event events.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"event"`
}

func (u *UserTicketLimits) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if u.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if u.MaxTickets <= 0 {
		return errors.New("max_tickets must be greater than 0")
	}
	if u.TicketsBought < 0 {
		return errors.New("tickets_bought cannot be negative")
	}
	if u.TicketsResold < 0 {
		return errors.New("tickets_resold cannot be negative")
	}
	return nil
}
func (u *UserTicketLimits) BeforeUpdate(tx *gorm.DB) error {
	if u.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if u.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if u.MaxTickets <= 0 {
		return errors.New("max_tickets must be greater than 0")
	}
	if u.TicketsBought < 0 {
		return errors.New("tickets_bought cannot be negative")
	}
	if u.TicketsResold < 0 {
		return errors.New("tickets_resold cannot be negative")
	}
	return nil
}
func (UserTicketLimits) TableName() string {
	return "user_ticket_limits"
}
