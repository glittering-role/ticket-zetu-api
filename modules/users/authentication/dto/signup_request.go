package dto

// SignUpRequest defines the structure for signup requests
type SignUpRequest struct {
	Username    string `json:"username" example:"john_doe99" validate:"required,min=6,max=50,regexp=^[a-zA-Z0-9_]+$"`
	FirstName   string `json:"first_name" example:"John" validate:"required,min=2,max=100"`
	LastName    string `json:"last_name" example:"Doe" validate:"required,min=2,max=100"`
	Email       string `json:"email" example:"john.doe@example.com" validate:"required,email,max=255"`
	Phone       string `json:"phone" example:"+12345678901" validate:"required,min=10,max=20"`
	Password    string `json:"password" example:"P@ssw0rd123" validate:"required,min=8"`
	DateOfBirth string `json:"date_of_birth,omitempty" example:"1990-01-15" validate:"omitempty,datetime=2006-01-02"`
}
