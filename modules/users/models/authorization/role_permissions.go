package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type RolePermission struct {
	ID           string         `gorm:"type:char(36);primaryKey" json:"id"`
	RoleID       string         `gorm:"type:char(36);index" json:"role_id"`
	PermissionID string         `gorm:"type:char(36);index" json:"permission_id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy    string         `gorm:"type:char(36)" json:"created_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Role       Role       `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role"`
	Permission Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permission"`
}

func (r *RolePermission) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
