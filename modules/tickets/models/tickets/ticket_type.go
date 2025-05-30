// models/ticket_type.go
package tickets

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTypeStatus string

const (
	TicketTypeActive   TicketTypeStatus = "active"
	TicketTypeInactive TicketTypeStatus = "inactive"
	TicketTypeArchived TicketTypeStatus = "archived"
)

type TicketType struct {
	ID                string           `gorm:"type:char(36);primaryKey" json:"id"`
	EventID           string           `gorm:"type:char(36);not null;index" json:"event_id"`
	OrganizerID       string           `gorm:"type:char(36);not null;index" json:"-"`
	Name              string           `gorm:"size:100;not null" json:"name"`
	Description       string           `gorm:"type:text" json:"description"`
	PriceModifier     float64          `gorm:"type:numeric(5,2);not null;default:1.0" json:"price_modifier"`
	Benefits          string           `gorm:"type:text" json:"benefits"`
	MaxTicketsPerUser int              `gorm:"not null;default:10" json:"max_tickets_per_user"`
	Status            TicketTypeStatus `gorm:"type:varchar(20);default:'active';check:status IN ('active','inactive','archived')" json:"status"`
	IsDefault         bool             `gorm:"default:false" json:"is_default"`
	SalesStart        time.Time        `gorm:"type:timestamp" json:"sales_start"`
	SalesEnd          *time.Time       `gorm:"type:timestamp;null" json:"sales_end"`
	QuantityAvailable *int             `gorm:"null" json:"quantity_available"`
	MinTicketsPerUser int              `gorm:"default:1" json:"min_tickets_per_user"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
	Version           int              `gorm:"default:1" json:"version"`

	// Relationships
	Event      events.Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"event"`
	PriceTiers []PriceTier  `gorm:"many2many:ticket_type_price_tiers;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"price_tiers"`
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) (err error) {
	if tt.ID == "" {
		tt.ID = uuid.New().String()
	}
	if tt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if tt.PriceModifier < 0 {
		return errors.New("price_modifier cannot be negative")
	}
	if tt.MaxTicketsPerUser < 1 {
		return errors.New("max_tickets_per_user must be at least 1")
	}
	if tt.MinTicketsPerUser < 1 {
		return errors.New("min_tickets_per_user must be at least 1")
	}
	if tt.MaxTicketsPerUser < tt.MinTicketsPerUser {
		return errors.New("max_tickets_per_user cannot be less than min_tickets_per_user")
	}
	if tt.SalesEnd != nil && tt.SalesStart.After(*tt.SalesEnd) {
		return errors.New("sales_start cannot be after sales_end")
	}
	if tt.QuantityAvailable != nil && *tt.QuantityAvailable < 0 {
		return errors.New("quantity_available cannot be negative")
	}
	return nil
}

func (tt *TicketType) BeforeUpdate(tx *gorm.DB) (err error) {
	if tt.Name == "" {
		return errors.New("name cannot be empty")
	}
	if tt.PriceModifier < 0 {
		return errors.New("price_modifier cannot be negative")
	}
	if tt.MaxTicketsPerUser < 1 {
		return errors.New("max_tickets_per_user must be at least 1")
	}
	if tt.MinTicketsPerUser < 1 {
		return errors.New("min_tickets_per_user must be at least 1")
	}
	if tt.MaxTicketsPerUser < tt.MinTicketsPerUser {
		return errors.New("max_tickets_per_user cannot be less than min_tickets_per_user")
	}
	if tt.SalesEnd != nil && tt.SalesStart.After(*tt.SalesEnd) {
		return errors.New("sales_start cannot be after sales_end")
	}
	if tt.QuantityAvailable != nil && *tt.QuantityAvailable < 0 {
		return errors.New("quantity_available cannot be negative")
	}
	return nil
}

func (TicketType) TableName() string {
	return "ticket_types"
}
