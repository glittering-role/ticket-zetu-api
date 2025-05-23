package organizers

import "time"

// UserDTO represents a limited user profile for organizer responses
type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// OrganizerResponse wraps Organizer with a UserDTO for created_by_user
type OrganizerResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	ContactPerson   string     `json:"contact_person"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	TaxID           string     `json:"tax_id,omitempty"`
	BankAccountInfo string     `json:"bank_account_info,omitempty"`
	ImageURL        string     `json:"image_url,omitempty"`
	CommissionRate  float64    `json:"commission_rate"`
	Balance         float64    `json:"balance"`
	Status          string     `json:"status"`
	IsFlagged       bool       `json:"is_flagged"`
	IsBanned        bool       `json:"is_banned"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       string     `json:"created_by,omitempty"`
	CreatedByUser   UserDTO    `json:"created_by_user"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
