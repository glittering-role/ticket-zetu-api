package members

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type SessionDeviceType string

const (
	DeviceMobile  SessionDeviceType = "mobile"
	DeviceDesktop SessionDeviceType = "desktop"
	DeviceTablet  SessionDeviceType = "tablet"
	DeviceUnknown SessionDeviceType = "unknown"
)

type UserSession struct {
	ID            uuid.UUID         `gorm:"type:char(36);primaryKey" json:"id"`
	UserID        uuid.UUID         `gorm:"type:char(36);index" json:"user_id"`
	SessionToken  string            `gorm:"type:text;not null" json:"-"`
	RefreshToken  string            `gorm:"type:text" json:"-"`
	IPAddress     string            `gorm:"type:varchar(45);index" json:"ip_address,omitempty"`
	UserAgent     string            `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceType    SessionDeviceType `gorm:"type:varchar(20);default:'unknown'" json:"device_type,omitempty"`
	DeviceName    string            `gorm:"type:varchar(100)" json:"device_name,omitempty"`
	IsActive      bool              `gorm:"default:true;index" json:"is_active"`
	LoggedOutAt   *time.Time        `gorm:"type:timestamp" json:"logged_out_at,omitempty"`
	CreatedAt     time.Time         `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt     time.Time         `gorm:"type:timestamp;index" json:"expires_at"`
	RefreshExpiry time.Time         `gorm:"type:timestamp" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (s *UserSession) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// Session helpers
func (s *UserSession) IsValid() bool {
	return s.IsActive && s.ExpiresAt.After(time.Now())
}

func (s *UserSession) CanRefresh() bool {
	return s.RefreshToken != "" && s.RefreshExpiry.After(time.Now())
}

func (s *UserSession) LocationInfo() map[string]string {
	return map[string]string{
		"ip_address":  s.IPAddress,
		"device_type": string(s.DeviceType),
		"device_name": s.DeviceName,
	}
}

func (s *UserSession) Terminate() {
	s.IsActive = false
	now := time.Now()
	s.LoggedOutAt = &now
}

func (s *UserSession) SafeInfo() map[string]interface{} {
	return map[string]interface{}{
		"created_at":  s.CreatedAt,
		"device_type": s.DeviceType,
		"device_name": s.DeviceName,
		"last_active": s.UpdatedAt,
		"is_active":   s.IsValid(),
		"expires_at":  s.ExpiresAt,
		"location":    s.LocationInfo(),
	}
}
