package events

import (
	"errors"
	"ticket-zetu-api/modules/organizers/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Venue struct {
	ID          string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Address     string         `gorm:"type:text;not null" json:"address"`
	City        string         `gorm:"size:100;not null;index" json:"city"`
	State       string         `gorm:"size:100;index" json:"state"`
	Country     string         `gorm:"size:100;not null;index" json:"country"`
	Capacity    int            `gorm:"default:0" json:"capacity"`
	ContactInfo string         `gorm:"type:text" json:"contact_info"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Status      string         `gorm:"type:varchar(20);default:'active';check:status IN ('active','inactive','suspended')" json:"status"`
	OrganizerID string         `gorm:"type:char(36);not null;index" json:"organizer_id"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version     int            `gorm:"default:1" json:"version"`

	// Relationships
	Events      []Event              `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"events"`
	VenueImages []VenueImage         `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"venue_images"`
	Organizer   organizers.Organizer `gorm:"foreignKey:OrganizerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"organizer"`
}

type VenueImage struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	VenueID   string         `gorm:"type:char(36);not null;index" json:"venue_id"`
	ImageURL  string         `gorm:"type:varchar(255);not null" json:"image_url"`
	IsPrimary bool           `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (v *Venue) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	if v.Name == "" {
		return errors.New("name cannot be empty")
	}
	if v.Address == "" {
		return errors.New("address cannot be empty")
	}
	if v.City == "" {
		return errors.New("city cannot be empty")
	}
	if v.Country == "" {
		return errors.New("country cannot be empty")
	}
	if v.Capacity < 0 {
		return errors.New("capacity cannot be negative")
	}
	if v.OrganizerID == "" {
		return errors.New("organizer ID cannot be empty")
	}
	return nil
}

func (v *Venue) BeforeUpdate(tx *gorm.DB) (err error) {
	if v.Name == "" {
		return errors.New("name cannot be empty")
	}
	if v.Address == "" {
		return errors.New("address cannot be empty")
	}
	if v.City == "" {
		return errors.New("city cannot be empty")
	}
	if v.Country == "" {
		return errors.New("country cannot be empty")
	}
	if v.Capacity < 0 {
		return errors.New("capacity cannot be negative")
	}
	if v.OrganizerID == "" {
		return errors.New("organizer ID cannot be empty")
	}
	return nil
}

func (vi *VenueImage) BeforeCreate(tx *gorm.DB) (err error) {
	if vi.ID == "" {
		vi.ID = uuid.New().String()
	}
	if vi.VenueID == "" {
		return errors.New("venue ID cannot be empty")
	}
	if vi.ImageURL == "" {
		return errors.New("image URL cannot be empty")
	}
	return nil
}

func (Venue) TableName() string {
	return "venues"
}

func (VenueImage) TableName() string {
	return "venue_images"
}
