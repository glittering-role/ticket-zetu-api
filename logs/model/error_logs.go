package model

import (
	"gorm.io/gorm"
)

// Log represents the structure of a log record.
type Log struct {
	ID          int             `gorm:"primaryKey;autoIncrement" json:"id"`
	Level       string          `gorm:"type:varchar(20);not null" json:"level"`
	Message     string          `gorm:"type:text;not null" json:"message"`
	Stack       *string         `gorm:"type:text" json:"stack,omitempty"`
	Context     *string         `gorm:"type:json" json:"context,omitempty"`
	Route       *string         `gorm:"type:varchar(255)" json:"route,omitempty"`
	Method      *string         `gorm:"type:varchar(10)" json:"method,omitempty"`
	StatusCode  *int            `gorm:"column:status_code" json:"status_code,omitempty"`
	UserID      *int            `gorm:"column:user_id" json:"user_id,omitempty"`
	IPAddress   *string         `gorm:"column:ip_address;type:varchar(100)" json:"ip_address,omitempty"`
	UserAgent   *string         `gorm:"column:user_agent;type:varchar(255)" json:"user_agent,omitempty"`
	File        *string         `gorm:"type:varchar(255)" json:"file,omitempty"`
	Line        *int            `gorm:"type:int" json:"line,omitempty"`
	Environment *string         `gorm:"type:varchar(255)" json:"environment,omitempty"`
	Occurrences int             `gorm:"default:1" json:"occurrences"`
	CreatedAt   *gorm.DeletedAt `json:"created_at,omitempty"`
	UpdatedAt   *gorm.DeletedAt `json:"updated_at,omitempty"`
}

// TableName sets the table name for the Log model
func (Log) TableName() string {
	return "logs"
}
