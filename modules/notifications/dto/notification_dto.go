package dto

import (
	notifications "ticket-zetu-api/modules/notifications/models"
	"time"
)

// CreateNotificationDTO represents input for creating a notification
type CreateNotificationDTO struct {
	Module       string                         `json:"module" validate:"required,max=50" example:"events"`
	Title        string                         `json:"title" validate:"required,max=100" example:"New Event Invite"`
	Content      string                         `json:"content" validate:"required,max=1000" example:"You've been invited to DJ Wave's event!"`
	Type         notifications.NotificationType `json:"type" validate:"required,oneof=action system warning" example:"action"`
	SenderID     string                         `json:"sender_id,omitempty" validate:"omitempty,uuid" example:"d6837fe1-645d-4451-8238-205c3796cd72"`
	RelatedID    string                         `json:"related_id,omitempty" validate:"omitempty,uuid" example:"54a5c863-90f5-4ecd-9c3d-f5298776cd19"`
	RecipientIDs []string                       `json:"recipient_ids" validate:"required,dive,uuid" example:"[\"e1f2a3b4-c5d6-4e7f-8a9b-0c1d2e3f4a5b\"]"`
	Metadata     map[string]interface{}         `json:"metadata,omitempty" example:"{\"event_id\": \"abc123\"}"`
}

// NotificationResponseDTO represents a notification in responses
type NotificationResponseDTO struct {
	ID        string                         `json:"id" example:"a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"`
	Type      notifications.NotificationType `json:"type" example:"action"`
	Title     string                         `json:"title" example:"New Event Invite"`
	Content   string                         `json:"content" example:"You've been invited to DJ Wave's event!"`
	SenderID  string                         `json:"sender_id,omitempty" example:"d6837fe1-645d-4451-8238-205c3796cd72"`
	Sender    *SenderDTO                     `json:"sender,omitempty"`
	RelatedID string                         `json:"related_id,omitempty" example:"54a5c863-90f5-4ecd-9c3d-f5298776cd19"`
	Module    string                         `json:"module" example:"events"`
	Metadata  map[string]interface{}         `json:"metadata,omitempty" example:"{\"event_id\": \"abc123\"}"`
	IsSystem  bool                           `json:"is_system" example:"false"`
	CreatedAt time.Time                      `json:"created_at" example:"2025-06-09T01:02:00Z"`
	UpdatedAt time.Time                      `json:"updated_at" example:"2025-06-09T01:02:00Z"`
}

// UserNotificationResponseDTO represents a user notification in responses
type UserNotificationResponseDTO struct {
	ID             string                           `json:"id" example:"b1c2d3e4-f5a6-4b7c-8d9e-0a1b2c3d4e5f"`
	NotificationID string                           `json:"notification_id" example:"a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"`
	Notification   NotificationResponseDTO          `json:"notification"`
	Status         notifications.NotificationStatus `json:"status" example:"unread"`
	ReadAt         *time.Time                       `json:"read_at,omitempty" example:"2025-06-09T01:05:00Z"`
	CreatedAt      time.Time                        `json:"created_at" example:"2025-06-09T01:02:00Z"`
	UpdatedAt      time.Time                        `json:"updated_at" example:"2025-06-09T01:02:00Z"`
}

type SenderDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
