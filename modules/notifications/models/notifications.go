package notifications

import (
	"errors"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeAction  NotificationType = "action"  // User actions (e.g., payment, follow)
	NotificationTypeSystem  NotificationType = "system"  // System-wide (e.g., maintenance)
	NotificationTypeWarning NotificationType = "warning" // Alerts (e.g., payment failure)
)

// NotificationStatus defines the status of a user notification
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

// Notification represents a notification event
type Notification struct {
	ID        string           `gorm:"type:char(36);primaryKey" json:"id"`
	Type      NotificationType `gorm:"type:varchar(50);not null;index" json:"type"`
	Title     string           `gorm:"type:varchar(100);not null" json:"title"`
	Content   string           `gorm:"type:text;not null" json:"content"`
	SenderID  string           `gorm:"type:char(36);index" json:"sender_id,omitempty"`
	Sender    members.User     `gorm:"foreignKey:SenderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"sender,omitempty"`
	RelatedID string           `gorm:"type:char(36);index" json:"related_id,omitempty"`
	Module    string           `gorm:"type:varchar(50);index" json:"module,omitempty"`
	Metadata  string           `gorm:"type:jsonb" json:"metadata,omitempty"`
	IsSystem  bool             `gorm:"default:false;index" json:"is_system"`
	CreatedAt time.Time        `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if n.Type == "" {
		return errors.New("notification type is required")
	}
	if n.Title == "" {
		return errors.New("notification title is required")
	}
	if n.Content == "" {
		return errors.New("notification content is required")
	}
	if n.SenderID != "" {
		if _, err := uuid.Parse(n.SenderID); err != nil {
			return errors.New("invalid sender ID format")
		}
	}
	if n.RelatedID != "" {
		if _, err := uuid.Parse(n.RelatedID); err != nil {
			return errors.New("invalid related ID format")
		}
	}
	if n.Module == "" {
		return errors.New("module is required")
	}
	return nil
}

func (Notification) TableName() string {
	return "notifications"
}

// UserNotification tracks notification delivery and status for a user
type UserNotification struct {
	ID             string             `gorm:"type:char(36);primaryKey" json:"id"`
	UserID         string             `gorm:"type:char(36);not null;index" json:"user_id"`
	User           members.User       `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	NotificationID string             `gorm:"type:char(36);not null;index" json:"notification_id"`
	Notification   Notification       `gorm:"foreignKey:NotificationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"notification"`
	Status         NotificationStatus `gorm:"type:varchar(20);not null;default:'unread';index" json:"status"`
	ReadAt         *time.Time         `gorm:"type:timestamp" json:"read_at,omitempty"`
	CreatedAt      time.Time          `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
}

func (un *UserNotification) BeforeCreate(tx *gorm.DB) (err error) {
	if un.ID == "" {
		un.ID = uuid.New().String()
	}
	if un.UserID == "" {
		return errors.New("user ID is required")
	}
	if _, err := uuid.Parse(un.UserID); err != nil {
		return errors.New("invalid user ID format")
	}
	if un.NotificationID == "" {
		return errors.New("notification ID is required")
	}
	if _, err := uuid.Parse(un.NotificationID); err != nil {
		return errors.New("invalid notification ID format")
	}
	if un.Status == "" {
		un.Status = NotificationStatusUnread
	}
	return nil
}

func (UserNotification) TableName() string {
	return "user_notifications"
}
