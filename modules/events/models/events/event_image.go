package events

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type EventImage struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	EventID   string         `gorm:"type:char(36);not null;index" json:"event_id"`
	ImageURL  string         `gorm:"size:255;not null" json:"image_url"`
	IsPrimary bool           `gorm:"default:false;index" json:"is_primary"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int            `gorm:"default:1" json:"version"`

	// Relationship
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
