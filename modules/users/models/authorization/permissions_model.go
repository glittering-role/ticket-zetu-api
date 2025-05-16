package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionStatus string

const (
	PermissionActive   PermissionStatus = "active"
	PermissionInactive PermissionStatus = "inactive"
)

type Permission struct {
	ID             string           `gorm:"type:char(36);primaryKey" json:"id"`
	PermissionName string           `gorm:"size:100;not null;uniqueIndex" json:"permission_name"`
	Description    string           `gorm:"size:255" json:"description"`
	Scope          string           `gorm:"size:255" json:"scope"` // e.g., "role:<=10" or "user:team_only"
	Status         PermissionStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	CreatedBy      string           `gorm:"type:char(36)" json:"created_by"`
	LastModifiedBy string           `gorm:"type:char(36)" json:"last_modified_by"`
	CreatedAt      time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
	Version        int              `gorm:"default:1" json:"version"`
	Roles          []Role           `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:PermissionID;References:ID;joinReferences:RoleID" json:"roles"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

func (Permission) TableName() string {
	return "permissions"
}
