package dto

type ResetPasswordRequest struct {
	UsernameOrEmail string `json:"username_or_email" example:"johndoe99 or john.doe@example.com" validate:"required,min=3,max=255"`
}

type SetNewPasswordRequest struct {
	ResetToken  string `json:"reset_token" example:"abc123xyz" validate:"required,min=32,max=64"`
	NewPassword string `json:"new_password" example:"NewP@ssw0rd123" validate:"required,min=8"`
}
