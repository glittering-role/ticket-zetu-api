package members

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

type UserLocation struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);index" json:"user_id"`
	Country   string    `gorm:"type:varchar(100);index" json:"country"`
	Continent string    `gorm:"type:varchar(50)" json:"continent"`
	Timezone  string    `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`
	State     string    `gorm:"type:varchar(100)" json:"state,omitempty"`
	StateName string    `gorm:"type:varchar(100)" json:"state_name,omitempty"`
	City      string    `gorm:"type:varchar(100);index" json:"city,omitempty"`
	Zip       string    `gorm:"type:varchar(20)" json:"zip,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (l *UserLocation) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func (UserLocation) TableName() string {
	return "user_locations"
}

// Helpers
func (l *UserLocation) FullLocation() string {
	parts := make([]string, 0, 3)
	if l.City != "" {
		parts = append(parts, l.City)
	}
	if l.StateName != "" {
		parts = append(parts, l.StateName)
	} else if l.State != "" {
		parts = append(parts, l.State)
	}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	if len(parts) == 0 {
		return "Unknown location"
	}
	return strings.Join(parts, ", ")
}
