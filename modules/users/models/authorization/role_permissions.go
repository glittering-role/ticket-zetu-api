package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RolePermission struct {
	ID           string         `gorm:"type:char(36);primaryKey" json:"id"`
	RoleID       string         `gorm:"type:char(36);index;not null" json:"role_id"`
	PermissionID string         `gorm:"type:char(36);index;not null" json:"permission_id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;not null" json:"created_at"`
	CreatedBy    string         `gorm:"type:char(36);not null" json:"created_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Role       Role       `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"role"`
	Permission Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"permission"`
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
