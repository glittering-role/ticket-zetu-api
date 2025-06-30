package tickets

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TicketHold struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	TicketTypeID string    `gorm:"type:char(36);not null;index" json:"ticket_type_id"`
	UserID       string    `gorm:"type:char(36);not null;index" json:"user_id"`
	SessionID    string    `gorm:"type:char(36);not null;index" json:"session_id"`
	Quantity     int       `gorm:"not null;default:1" json:"quantity"`
	HeldUntil    time.Time `gorm:"not null" json:"held_until"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Indexes for faster cleanup
	Indexes []struct{} `gorm:"index:,expression:EXTRACT(EPOCH FROM (held_until - NOW())) WHERE held_until > NOW()"`
}

func (th *TicketHold) BeforeCreate() error {
	if th.ID == "" {
		th.ID = uuid.New().String()
	}
	if th.TicketTypeID == "" {
		return errors.New("ticket_type_id cannot be empty")
	}
	if th.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if th.SessionID == "" {
		return errors.New("session_id cannot be empty")
	}
	if th.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if th.HeldUntil.IsZero() {
		return errors.New("held_until cannot be empty")
	}
	return nil
}

func (th *TicketHold) BeforeUpdate() error {
	return th.BeforeCreate()
}

func (TicketHold) TableName() string {
	return "ticket_holds"
}
