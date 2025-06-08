package events

import (
	"errors"
	"ticket-zetu-api/modules/events/models/categories"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventActive   EventStatus = "active"
	EventInactive EventStatus = "inactive"
)

type EventType string

const (
	EventTypeOnline  EventType = "online"
	EventTypeOffline EventType = "offline"
	EventTypeHybrid  EventType = "hybrid"
)

type Event struct {
	ID            string                 `gorm:"type:char(36);primaryKey" json:"id"`
	Title         string                 `gorm:"size:255;not null;index" json:"title"`
	Slug          string                 `gorm:"size:255;uniqueIndex" json:"slug"`
	Description   string                 `gorm:"type:text" json:"description"`
	SubcategoryID string                 `gorm:"not null;index" json:"subcategory_id"`
	Subcategory   categories.Subcategory `gorm:"foreignKey:SubcategoryID"`
	VenueID       string                 `gorm:"type:char(36);index" json:"venue_id"`
	Venue         Venue                  `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"venue"`

	StartTime time.Time `gorm:"not null;index" json:"start_time"`
	EndTime   time.Time `gorm:"not null;index" json:"end_time"`
	Timezone  string    `gorm:"size:100" json:"timezone,omitempty"`
	Language  string    `gorm:"size:50" json:"language,omitempty"`

	OrganizerID    string    `gorm:"type:char(36);not null;index" json:"-"`
	EventType      EventType `gorm:"size:20;default:'offline'" json:"event_type"`
	MinAge         int       `gorm:"not null;default:0" json:"min_age"`
	TotalSeats     int       `gorm:"not null" json:"total_seats"`
	AvailableSeats int       `gorm:"not null" json:"available_seats"`

	IsFree     bool        `gorm:"default:false" json:"is_free"`
	HasTickets bool        `gorm:"default:true" json:"has_tickets"`
	IsFeatured bool        `gorm:"default:false;index" json:"is_featured"`
	Status     EventStatus `gorm:"size:20;not null;default:'active';index" json:"status"`
	Tags       []string    `gorm:"type:text[]" json:"tags,omitempty"`

	PublishedAt *time.Time `gorm:"index" json:"published_at,omitempty"`

	EventImages []EventImage `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"event_images"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int            `gorm:"default:1" json:"-"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.TotalSeats <= 0 {
		return errors.New("total_seats must be positive")
	}
	if e.AvailableSeats < 0 || e.AvailableSeats > e.TotalSeats {
		return errors.New("available_seats must be non-negative and not exceed total_seats")
	}
	if _, err := uuid.Parse(e.OrganizerID); err != nil {
		return errors.New("invalid organizer_id format")
	}
	return nil
}

func (e *Event) BeforeUpdate(tx *gorm.DB) (err error) {
	if e.TotalSeats <= 0 {
		return errors.New("total_seats must be positive")
	}
	if e.AvailableSeats < 0 || e.AvailableSeats > e.TotalSeats {
		return errors.New("available_seats must be non-negative and not exceed total_seats")
	}
	if _, err := uuid.Parse(e.OrganizerID); err != nil {
		return errors.New("invalid organizer_id format")
	}
	return nil
}

type EventImage struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	EventID   string         `gorm:"type:char(36);not null;index" json:"event_id"`
	ImageURL  string         `gorm:"size:255;not null" json:"image_url"`
	IsPrimary bool           `gorm:"default:false;index" json:"is_primary"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int            `gorm:"default:1" json:"version"`

	Event Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (ei *EventImage) BeforeCreate(tx *gorm.DB) (err error) {
	if ei.ID == "" {
		ei.ID = uuid.New().String()
	}
	return nil
}

func (EventImage) TableName() string {
	return "event_images"
}

func (Event) TableName() string {
	return "events"
}
