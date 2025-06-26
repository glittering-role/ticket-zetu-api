package events

import (
	"encoding/json"
	"errors"
	"time"

	"ticket-zetu-api/modules/organizers/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VenueType string

const (
	VenueTypeStadium VenueType = "stadium"
	VenueTypeHotel   VenueType = "hotel"
	VenueTypePark    VenueType = "park"
	VenueTypeTheater VenueType = "theater"
	VenueTypeOther   VenueType = "other"
)

type VenueStatus string

const (
	VenueStatusActive    VenueStatus = "active"
	VenueStatusInactive  VenueStatus = "inactive"
	VenueStatusSuspended VenueStatus = "suspended"
)

type Venue struct {
	ID                    string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name                  string         `gorm:"size:255;not null;index" json:"name"`
	Description           string         `gorm:"type:text" json:"description,omitempty"`
	Address               string         `gorm:"type:text;not null" json:"address"`
	City                  string         `gorm:"size:100;not null;index" json:"city"`
	State                 string         `gorm:"size:100;index" json:"state"`
	PostalCode            string         `gorm:"size:20" json:"postal_code"`
	Country               string         `gorm:"size:100;not null;index" json:"country"`
	Latitude              float64        `gorm:"type:decimal(10,6)" json:"latitude"`
	Longitude             float64        `gorm:"type:decimal(10,6)" json:"longitude"`
	Capacity              int            `gorm:"default:0;check:capacity >= 0" json:"capacity"`
	VenueType             VenueType      `gorm:"type:varchar(20);not null;default:'other'" json:"venue_type"`
	Layout                string         `gorm:"type:json" json:"layout,omitempty"`
	AccessibilityFeatures string         `gorm:"type:json" json:"accessibility_features,omitempty"`
	Facilities            string         `gorm:"type:json" json:"facilities,omitempty"`
	ContactInfo           string         `gorm:"type:text" json:"contact_info"`
	Timezone              string         `gorm:"size:100" json:"timezone"`
	Status                VenueStatus    `gorm:"type:varchar(20);not null;default:'active';check:status IN ('active','inactive','suspended')" json:"status"`
	OrganizerID           string         `gorm:"type:char(36);not null;index" json:"organizer_id"`
	CreatedAt             time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version               int            `gorm:"default:1" json:"version"`

	// Relationships
	Events      []Event              `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"events"`
	VenueImages []VenueImage         `gorm:"foreignKey:VenueID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"venue_images"`
	Organizer   organizers.Organizer `gorm:"foreignKey:OrganizerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"organizer"`
}

type VenueImage struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	VenueID   string         `gorm:"not null;index" json:"venue_id"`
	ImageURL  string         `gorm:"type:varchar(255);not null" json:"image_url"`
	IsPrimary bool           `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (v *Venue) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	if err := v.validate(tx); err != nil {
		return err
	}
	return nil
}

func (v *Venue) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := v.validate(tx); err != nil {
		return err
	}
	return nil
}

func (v *Venue) validate(tx *gorm.DB) error {
	if v.OrganizerID == "" {
		return errors.New("organizer_id cannot be empty")
	}
	if v.Status != VenueStatusActive && v.Status != VenueStatusInactive && v.Status != VenueStatusSuspended {
		return errors.New("status must be one of 'active', 'inactive', 'suspended'")
	}
	if v.Layout != "" {
		var layout map[string]interface{}
		if err := json.Unmarshal([]byte(v.Layout), &layout); err != nil {
			return errors.New("layout must be valid JSON")
		}
	}
	if v.AccessibilityFeatures != "" {
		var features []string
		if err := json.Unmarshal([]byte(v.AccessibilityFeatures), &features); err != nil {
			return errors.New("accessibility_features must be valid JSON")
		}
	}
	if v.Facilities != "" {
		var facilities []string
		if err := json.Unmarshal([]byte(v.Facilities), &facilities); err != nil {
			return errors.New("facilities must be valid JSON")
		}
	}

	return nil
}

func (vi *VenueImage) BeforeCreate(tx *gorm.DB) (err error) {
	if vi.ID == "" {
		vi.ID = uuid.New().String()
	}
	if vi.VenueID == "" {
		return errors.New("venue_id cannot be empty")
	}
	if vi.ImageURL == "" {
		return errors.New("image_url cannot be empty")
	}
	return nil
}

func (vi *VenueImage) BeforeUpdate(tx *gorm.DB) (err error) {
	if vi.VenueID == "" {
		return errors.New("venue_id cannot be empty")
	}
	if vi.ImageURL == "" {
		return errors.New("image_url cannot be empty")
	}
	return nil
}

func (Venue) TableName() string {
	return "venues"
}

func (VenueImage) TableName() string {
	return "venue_images"
}
