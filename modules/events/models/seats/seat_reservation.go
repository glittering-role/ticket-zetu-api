package seats

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SeatReservation struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    string         `gorm:";not null;index" json:"user_id"`
	EventID   string         `gorm:";not null;index" json:"event_id"`
	SeatID    string         `gorm:";not null;index" json:"seat_id"`
	Status    string         `gorm:"type:varchar(20);not null;default:'held';check:status IN ('held','confirmed','released')" json:"status"`
	ExpiresAt time.Time      `gorm:"type:timestamp;not null;index" json:"expires_at"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int            `gorm:"default:1" json:"version"`

	User  members.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user"`
	Event events.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"event"`
	Seat  Seat         `gorm:"foreignKey:SeatID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"seat"`
}

func (sr *SeatReservation) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == "" {
		sr.ID = uuid.New().String()
	}
	if sr.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if sr.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if sr.SeatID == "" {
		return errors.New("seat_id cannot be empty")
	}
	if sr.ExpiresAt.IsZero() {
		return errors.New("expires_at cannot be empty")
	}
	return nil
}

func (sr *SeatReservation) BeforeUpdate(tx *gorm.DB) error {
	return sr.BeforeCreate(tx)
}

func (SeatReservation) TableName() string {
	return "seat_reservations"
}
