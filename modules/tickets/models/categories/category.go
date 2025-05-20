package categories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Category struct {
	ID                string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name              string         `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description,omitempty"`
	ImageURL          string         `gorm:"type:text" json:"image_url,omitempty"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	LastUpdatedBy     string         `gorm:"type:char(36);index" json:"last_updated_by,omitempty"`
	LastUpdatedByUser User           `gorm:"foreignKey:LastUpdatedBy;references:ID;constraint:OnDelete:SET NULL" json:"last_updated_by_user,omitempty"`
	Subcategories     []Subcategory  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"subcategories,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate assigns a UUID if not set
func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (Category) TableName() string {
	return "categories"
}

// User is a minimal struct to avoid circular imports
type User struct {
	ID string `gorm:"type:char(36);primaryKey"`
}
