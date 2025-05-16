package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RolePermission struct {
	ID           string         `gorm:"type:char(36);primaryKey;index" json:"role_id"`
	PermissionID string         `gorm:"type:char(36);primaryKey;index" json:"permission_id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy    string         `gorm:"type:char(36)" json:"created_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Role         Role           `gorm:"foreignKey:RoleID;references:ID" json:"role"`
	Permission   Permission     `gorm:"foreignKey:PermissionID;references:ID" json:"permission"`
}

func (r *RolePermission) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
