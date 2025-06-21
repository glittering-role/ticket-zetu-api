package events

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Favorite struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(36);not null;index:idx_user_event" json:"user_id"`
	EventID   string    `gorm:"type:char(36);not null;index:idx_user_event" json:"event_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type VoteType string

const (
	VoteTypeUp   VoteType = "upvote"
	VoteTypeDown VoteType = "downvote"
)

type Vote struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(36);not null;index:idx_user_event_type" json:"user_id"`
	EventID   string    `gorm:"type:char(36);not null;index:idx_user_event_type" json:"event_id"`
	Type      VoteType  `gorm:"type:varchar(10);not null;index:idx_user_event_type" json:"type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Comment struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(36);not null;index" json:"user_id"`
	EventID   string    `gorm:"type:char(36);not null;index" json:"event_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *string   `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Replies []Comment `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"replies,omitempty"`
}

// BeforeCreate hooks (remain unchanged)
func (f *Favorite) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	if f.UserID == "" || f.EventID == "" {
		return errors.New("user_id and event_id are required")
	}
	return nil
}

func (v *Vote) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	if v.UserID == "" || v.EventID == "" {
		return errors.New("user_id and event_id are required")
	}
	if v.Type != VoteTypeUp && v.Type != VoteTypeDown {
		return errors.New("invalid vote type")
	}
	return nil
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.UserID == "" || c.EventID == "" {
		return errors.New("user_id and event_id are required")
	}
	if c.Content == "" {
		return errors.New("content cannot be empty")
	}
	return nil
}

// Table names (remain unchanged)
func (Favorite) TableName() string {
	return "event_favorites"
}

func (Vote) TableName() string {
	return "event_votes"
}

func (Comment) TableName() string {
	return "event_comments"
}
