package members

import (
	model "ticket-zetu-api/modules/users/models/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             string          `gorm:"type:char(36);primaryKey" json:"id"`
	Username       string          `gorm:"size:50;not null;uniqueIndex" json:"username"`
	FirstName      string          `gorm:"size:100;not null" json:"first_name"`
	LastName       string          `gorm:"size:100;not null" json:"last_name"`
	Email          string          `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Phone          string          `gorm:"size:20;not null;uniqueIndex" json:"phone"`
	AvatarURL      string          `gorm:"type:text" json:"avatar_url"`
	DateOfBirth    *time.Time      `gorm:"type:date" json:"date_of_birth,omitempty"`
	Gender         string          `gorm:"type:varchar(50)" json:"gender,omitempty"`
	RoleID         string          `gorm:"type:char(36);index" json:"role_id"`
	Role           model.Role      `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"role"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	CreatedBy      string          `gorm:"type:char(36)" json:"created_by"`
	LastModifiedBy string          `gorm:"type:char(36)" json:"last_modified_by"`
	Preferences    UserPreferences `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"preferences"`
	Location       UserLocation    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"location"`
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
