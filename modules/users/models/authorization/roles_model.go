package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleStatus string

const (
	RoleActive   RoleStatus = "active"
	RoleInactive RoleStatus = "inactive"
	RoleArchived RoleStatus = "archived"
)

type Role struct {
	ID             string         `gorm:"type:char(36);primaryKey" json:"id"`
	RoleName       string         `gorm:"size:100;not null;uniqueIndex" json:"role_name"`
	Description    string         `gorm:"size:255" json:"description"`
	Level          int            `gorm:"not null;default:1" json:"level"`
	Status         RoleStatus     `gorm:"size:20;not null;default:'active'" json:"status"`
	IsSystemRole   bool           `gorm:"default:false" json:"is_system_role"`
	NumberOfUsers  int64          `gorm:"not null;default:0" json:"number_of_users"`
	CreatedBy      string         `gorm:"type:char(36)" json:"created_by"`
	LastModifiedBy string         `gorm:"type:char(36)" json:"last_modified_by"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version        int            `gorm:"default:1" json:"version"`
	Permissions    []Permission   `gorm:"many2many:role_permissions;" json:"permissions"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return
}

func (Role) TableName() string {
	return "roles"
}
