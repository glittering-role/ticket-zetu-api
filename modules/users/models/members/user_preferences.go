package members

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

type UserPreferences struct {
	ID             uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	UserID         uuid.UUID      `gorm:"type:char(36);uniqueIndex" json:"user_id"`
	ShowEmail      bool           `gorm:"default:false" json:"show_email"`
	ShowPhone      bool           `gorm:"default:false" json:"show_phone"`
	ShowLocation   bool           `gorm:"default:false" json:"show_location"`
	ShowGender     bool           `gorm:"default:false" json:"show_gender"`
	ShowRole       bool           `gorm:"default:false" json:"show_role"`
	ShowProfile    bool           `gorm:"default:true;index" json:"show_profile"`
	AllowFollowing bool           `gorm:"default:true;index" json:"allow_following"`
	Language       string         `gorm:"type:varchar(10);default:'en';index" json:"language"`
	Theme          string         `gorm:"type:varchar(20);default:'light'" json:"theme"`
	Timezone       string         `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *UserPreferences) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

// Helpers
func (p *UserPreferences) IsPublicProfile() bool {
	return p.ShowProfile && (p.ShowEmail || p.ShowPhone || p.ShowLocation || p.ShowGender || p.ShowRole)
}

func (p *UserPreferences) VisibleFields() []string {
	fields := make([]string, 0, 5)
	if p.ShowEmail {
		fields = append(fields, "email")
	}
	if p.ShowPhone {
		fields = append(fields, "phone")
	}
	if p.ShowLocation {
		fields = append(fields, "location")
	}
	if p.ShowGender {
		fields = append(fields, "gender")
	}
	if p.ShowRole {
		fields = append(fields, "role")
	}
	return fields
}

func (p *UserPreferences) GetLocale() string {
	switch p.Language {
	case "en":
		return "en_US"
	case "es":
		return "es_ES"
	case "fr":
		return "fr_FR"
	default:
		return p.Language + "_" + strings.ToUpper(p.Language)
	}
}

func (p *UserPreferences) ShouldShow(field string) bool {
	switch field {
	case "email":
		return p.ShowEmail && p.ShowProfile
	case "phone":
		return p.ShowPhone && p.ShowProfile
	case "location":
		return p.ShowLocation && p.ShowProfile
	case "gender":
		return p.ShowGender && p.ShowProfile
	case "role":
		return p.ShowRole && p.ShowProfile
	default:
		return false
	}
}
