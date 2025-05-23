package categories

import (
	"errors"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subcategory struct {
	ID                string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name              string         `gorm:"type:varchar(50);not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description,omitempty"`
	ImageURL          string         `gorm:"type:text" json:"image_url,omitempty"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	LastUpdatedBy     string         `gorm:"type:char(36);index" json:"last_updated_by,omitempty"`
	LastUpdatedByUser members.User   `gorm:"foreignKey:LastUpdatedBy;references:ID;constraint:OnDelete:SET NULL" json:"last_updated_by_user,omitempty"`
	CategoryID        string         `gorm:"type:char(36);not null;index" json:"category_id"`
	Category          Category       `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"category,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (s *Subcategory) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	// Ensure unique (category_id, name)
	var count int64
	if err := tx.Model(&Subcategory{}).
		Where("category_id = ? AND name = ? AND deleted_at IS NULL", s.CategoryID, s.Name).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("subcategory name must be unique within category")
	}
	return nil
}

func (Subcategory) TableName() string {
	return "sub_categories"
}
