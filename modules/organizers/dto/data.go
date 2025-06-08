package organizer_dto

type CreateOrganizerData struct {
	Name            string  `json:"name" example:"Event Masters Ltd" validate:"required,min=2,max=255"`
	ContactPerson   string  `json:"contact_person" example:"Jane Doe" validate:"required,min=2,max=255"`
	Email           string  `json:"email" example:"jane.doe@eventmasters.com" validate:"required,email"`
	Phone           string  `json:"phone,omitempty" example:"+254712345678" validate:"max=50"`
	CompanyName     string  `json:"company_name,omitempty" example:"Event Masters Kenya" validate:"max=255"`
	TaxID           string  `json:"tax_id,omitempty" example:"KRA12345678A" validate:"max=100"`
	BankAccountInfo string  `json:"bank_account_info,omitempty" example:"Bank: Equity Bank, Acc No: 1234567890"`
	ImageURL        string  `json:"image_url,omitempty" example:"https://example.com/images/organizer-logo.png" validate:"max=255"`
	CommissionRate  float64 `json:"commission_rate" example:"15.5" validate:"gte=0,lte=100"`
	Balance         float64 `json:"balance" example:"10000.00" validate:"gte=0"`
	Notes           string  `json:"notes,omitempty" example:"Preferred partner with high ticket volumes."`
}
