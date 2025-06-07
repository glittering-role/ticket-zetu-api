package tickets

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"ticket-zetu-api/modules/events/models/events"
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
	PriceModifier     float64          `gorm:"type:numeric(5,2);not null;default:1.0;check:price_modifier >= 0" json:"price_modifier"`
	Benefits          string           `gorm:"type:text" json:"benefits"`
	MinTicketsPerUser int              `gorm:"default:1;check:min_tickets_per_user >= 1" json:"min_tickets_per_user"`
	MaxTicketsPerUser int              `gorm:"default:10;check:max_tickets_per_user >= 1" json:"max_tickets_per_user"`
	QuantityAvailable *int             `gorm:"check:quantity_available >= 0" json:"quantity_available,omitempty"`
	Status            TicketTypeStatus `gorm:"type:varchar(20);default:'active';check:status IN ('active','inactive','archived')" json:"status"`
	IsDefault         bool             `gorm:"default:false" json:"is_default"`
	SalesStart        time.Time        `gorm:"type:timestamp" json:"sales_start"`
	SalesEnd          *time.Time       `gorm:"type:timestamp" json:"sales_end,omitempty"`
	Version           int              `gorm:"default:1" json:"version"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Event      events.Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"event"`
	PriceTiers []PriceTier  `gorm:"many2many:ticket_type_price_tiers;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"price_tiers"`
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) error {
	tt.setDefaults()
	return tt.validate()
}

func (tt *TicketType) BeforeUpdate(tx *gorm.DB) error {
	return tt.validate()
}

func (tt *TicketType) setDefaults() {
	if tt.ID == "" {
		tt.ID = uuid.New().String()
	}
}

func (tt *TicketType) validate() error {
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
