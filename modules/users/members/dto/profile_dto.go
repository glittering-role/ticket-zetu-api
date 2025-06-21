package dto

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
	IsVerified     bool                  `json:"is_verified" example:"true"`
	CreatedAt      string                `json:"created_at" example:"2025-06-05T10:00:00Z"`
	UpdatedAt      string                `json:"updated_at" example:"2025-10-01T12:00:00Z"`
}

type UserLocationDto struct {
	Country   string `json:"country,omitempty" example:"US"`
	State     string `json:"state,omitempty" example:"NY"`
	StateName string `json:"state_name,omitempty" example:"New York"`
	Continent string `json:"continent,omitempty" example:"North America"`
	City      string `json:"city,omitempty" example:"New York"`
	Zip       string `json:"zip,omitempty" example:"10001"`
	Timezone  string `json:"timezone" example:"America/New_York"`
}

type UpdateUserDto struct {
	FirstName   *string `json:"first_name" validate:"omitempty,max=100" example:"Alice"`
	LastName    *string `json:"last_name" validate:"omitempty,max=100" example:"Johnson"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,url" example:"https://cdn.example.com/avatars/alice.jpg"`
	DateOfBirth *string `json:"date_of_birth" validate:"omitempty,datetime=2006-01-02" example:"1990-04-15"`
	Gender      *string `json:"gender" validate:"omitempty,max=50" example:"Female"`
}

type UpdatePhoneDto struct {
	Phone string `json:"phone" validate:"required,max=20" example:"+1234567890"`
}

type UpdateEmailDto struct {
	Email string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
}

type UpdateUsernameDto struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
}

type NewPasswordDto struct {
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=128"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}

type UserLocationUpdateDto struct {
	Country    *string `json:"country,omitempty" validate:"omitempty,max=100" example:"Kenya"`
	State      *string `json:"state,omitempty" validate:"omitempty,max=100" example:"Nairobi"`
	StateName  *string `json:"state_name,omitempty" validate:"omitempty,max=100" example:"Nairobi County"`
	Continent  *string `json:"continent,omitempty" validate:"omitempty,max=50" example:"Africa"`
	City       *string `json:"city,omitempty" validate:"omitempty,max=100" example:"Nairobi"`
	Zip        *string `json:"zip,omitempty" validate:"omitempty,max=20" example:"00100"`
	Timezone   *string `json:"timezone,omitempty" validate:"omitempty,max=50" example:"Africa/Nairobi"`
	LastActive *string `json:"last_active,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2025-06-10T14:30:00+03:00"`
}
