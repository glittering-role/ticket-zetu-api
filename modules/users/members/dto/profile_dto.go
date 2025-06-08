package dto

// UserProfileResponseDto defines the response structure for user profile data
type UserProfileResponseDto struct {
	ID             string                `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username       string                `json:"username" example:"johndoe"`
	FirstName      string                `json:"first_name" example:"John"`
	LastName       string                `json:"last_name" example:"Doe"`
	AvatarURL      string                `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Email          string                `json:"email,omitempty" example:"john.doe@example.com"`
	Phone          string                `json:"phone,omitempty" example:"+1234567890"`
	DateOfBirth    *string               `json:"date_of_birth,omitempty" example:"1990-01-01"`
	Gender         string                `json:"gender,omitempty" example:"male"`
	RoleName       string                `json:"role_name,omitempty" example:"Admin"`
	Location       string                `json:"location,omitempty" example:"New York, NY, US"`
	LocationDetail *UserLocationDto      `json:"location_detail,omitempty"`
	ArtistProfile  *ReadArtistProfileDTO `json:"artist_profile,omitempty"`

	CreatedAt string `json:"created_at" example:"2025-06-05T10:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2025-10-01T12:00:00Z"`
}

// UserLocationDto defines detailed location data for user profiles
type UserLocationDto struct {
	Country    string  `json:"country,omitempty" example:"US"`
	State      string  `json:"state,omitempty" example:"NY"`
	StateName  string  `json:"state_name,omitempty" example:"New York"`
	Continent  string  `json:"continent,omitempty" example:"North America"`
	City       string  `json:"city,omitempty" example:"New York"`
	Zip        string  `json:"zip,omitempty" example:"10001"`
	Timezone   string  `json:"timezone" example:"America/New_York"`
	LastActive *string `json:"last_active,omitempty" example:"2025-06-05T08:00:00Z"`
}

// UpdateUserDto defines the input structure for updating user details
type UpdateUserDto struct {
	FirstName   *string `json:"first_name" validate:"omit,max=100"`
	LastName    *string `json:"last_name" validate:"omit,max=100"`
	AvatarURL   *string `json:"avatar_url" validate:"omit,url"`
	DateOfBirth *string `json:"date_of_birth" validate:"omit,datetime=2006-01-02"`
	Gender      *string `json:"gender" validate:"omit,max=50"`
}

// UpdatePhoneDto defines the input structure for updating phone number
type UpdatePhoneDto struct {
	Phone string `json:"phone" validate:"required,max=20"`
}

// UpdateEmailDto defines the input structure for updating email
type UpdateEmailDto struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// UserLocationUpdateDto defines the input structure for updating user location
type UserLocationUpdateDto struct {
	Country    *string `json:"country,omitempty" validate:"omitempty,max=100"`
	State      *string `json:"state,omitempty" validate:"omitempty,max=100"`
	StateName  *string `json:"state_name,omitempty" validate:"omitempty,max=100"`
	Continent  *string `json:"continent,omitempty" validate:"omitempty,max=50"`
	City       *string `json:"city,omitempty" validate:"omitempty,max=100"`
	Zip        *string `json:"zip,omitempty" validate:"omitempty,max=20"`
	Timezone   *string `json:"timezone,omitempty" validate:"omitempty,max=50"`
	LastActive *string `json:"last_active,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
