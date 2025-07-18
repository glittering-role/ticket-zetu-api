package members

import (
	"ticket-zetu-api/modules/users/models/artist"
	model "ticket-zetu-api/modules/users/models/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             string               `gorm:"type:char(36);primaryKey" json:"id"`
	Username       string               `gorm:"size:50;not null;uniqueIndex" json:"username"`
	FirstName      string               `gorm:"size:100;not null" json:"first_name"`
	LastName       string               `gorm:"size:100;not null" json:"last_name"`
	Email          string               `gorm:"size:255;uniqueIndex" json:"email"`
	Phone          string               `gorm:"size:20;not null;uniqueIndex" json:"phone"`
	AvatarURL      string               `gorm:"type:text" json:"avatar_url"`
	DateOfBirth    *time.Time           `gorm:"type:date" json:"date_of_birth,omitempty"`
	Gender         string               `gorm:"type:varchar(50)" json:"gender,omitempty"`
	IsVerified     bool                 `gorm:"default:false;index" json:"is_verified"`
	VerifiedAt     *time.Time           `gorm:"type:timestamp" json:"verified_at,omitempty"`
	RoleID         string               `gorm:"type:char(36);index" json:"role_id"`
	Role           model.Role           `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"role"`
	ArtistProfile  artist.ArtistProfile `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"artist_profile,omitempty"`
	CreatedAt      time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt       `gorm:"index" json:"deleted_at,omitempty"`
	LastModifiedBy string               `gorm:"type:char(36)" json:"last_modified_by"`
	Preferences    UserPreferences      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"preferences"`
	Location       UserLocation         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"location"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (User) TableName() string {
	return "user_profiles"
}
