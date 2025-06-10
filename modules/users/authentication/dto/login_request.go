package dto

// LoginRequest defines the structure for login requests
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	RememberMe      bool   `json:"remember_me"`
}
