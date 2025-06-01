package organizer_dto

type OrganizerResponse struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	ContactPerson   string        `json:"contact_person"`
	Email           string        `json:"email"`
	Phone           string        `json:"phone,omitempty"`
	CompanyName     string        `json:"company_name,omitempty"`
	TaxID           string        `json:"tax_id,omitempty"`
	BankAccountInfo string        `json:"bank_account_info,omitempty"`
	ImageURL        string        `json:"image_url,omitempty"`
	CommissionRate  float64       `json:"commission_rate"`
	Balance         float64       `json:"balance"`
	Status          string        `json:"status"`
	IsFlagged       bool          `json:"is_flagged"`
	IsBanned        bool          `json:"is_banned"`
	CreatedBy       string        `json:"created_by"`
	CreatedByUser   *UserResponse `json:"created_by_user,omitempty"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
