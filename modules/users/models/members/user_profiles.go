package members

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	authorization "ticket-zetu-api/modules/users/models/authorization"
	"time"
)

type User struct {
	ID             string             `gorm:"type:char(36);primaryKey" json:"id"`
	FirstName      string             `gorm:"size:100;not null" json:"first_name"`
	LastName       string             `gorm:"size:100;not null" json:"last_name"`
	Email          string             `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Phone          string             `gorm:"size:20;not null;uniqueIndex" json:"phone"`
	AvatarURL      string             `gorm:"type:text" json:"avatar_url"`
	DateOfBirth    *time.Time         `gorm:"type:date" json:"date_of_birth,omitempty"`
	RoleID         string             `gorm:"type:char(36);index" json:"role_id"`
	Role           authorization.Role `gorm:"foreignKey:RoleID;references:ID" json:"role"`
	CreatedAt      time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	CreatedBy      string             `gorm:"type:char(36)" json:"created_by"`
	LastModifiedBy string             `gorm:"type:char(36)" json:"last_modified_by"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

func (User) TableName() string {
	return "user_profiles"
}
