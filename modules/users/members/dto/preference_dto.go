package dto

import (
	"github.com/google/uuid"
)

// UserPreferencesDto represents the response for user preferences
type UserPreferencesDto struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	ShowEmail      bool      `json:"show_email"`
	ShowPhone      bool      `json:"show_phone"`
	ShowLocation   bool      `json:"show_location"`
	ShowGender     bool      `json:"show_gender"`
	ShowRole       bool      `json:"show_role"`
	ShowProfile    bool      `json:"show_profile"`
	AllowFollowing bool      `json:"allow_following"`
	Language       string    `json:"language"`
	Theme          string    `json:"theme"`
	Timezone       string    `json:"timezone"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// UpdateUserPreferencesDto represents the input for updating user preferences
type UpdateUserPreferencesDto struct {
	ShowEmail      *bool   `json:"show_email" validate:"omitempty"`
	ShowPhone      *bool   `json:"show_phone" validate:"omitempty"`
	ShowLocation   *bool   `json:"show_location" validate:"omitempty"`
	ShowGender     *bool   `json:"show_gender" validate:"omitempty"`
	ShowRole       *bool   `json:"show_role" validate:"omitempty"`
	ShowProfile    *bool   `json:"show_profile" validate:"omitempty"`
	AllowFollowing *bool   `json:"allow_following" validate:"omitempty"`
	Language       *string `json:"language" validate:"omitempty,max=10" example:"en"`
	Theme          *string `json:"theme" validate:"omitempty,oneof=light dark" example:"dark"`
	Timezone       *string `json:"timezone" validate:"omitempty,max=100" example:"Africa/Nairobi"`
}
