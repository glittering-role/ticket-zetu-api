package organizers

import "github.com/gofiber/fiber/v2"

func (c *OrganizerController) CreateOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	var input struct {
		Name            string  `json:"name" validate:"required,min=2,max=255"`
		ContactPerson   string  `json:"contact_person" validate:"required,min=2,max=255"`
		Email           string  `json:"email" validate:"required,email"`
		Phone           string  `json:"phone,omitempty" validate:"max=50"`
		CompanyName     string  `json:"company_name,omitempty" validate:"max=255"`
		TaxID           string  `json:"tax_id,omitempty" validate:"max=100"`
		BankAccountInfo string  `json:"bank_account_info,omitempty"`
		ImageURL        string  `json:"image_url,omitempty" validate:"max=255"`
		CommissionRate  float64 `json:"commission_rate" validate:"gte=0,lte=100"`
		Balance         float64 `json:"balance" validate:"gte=0"`
		Notes           string  `json:"notes,omitempty"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Basic validation
	if len(input.Name) < 2 || len(input.Name) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 255 characters"), fiber.StatusBadRequest)
	}
	if len(input.ContactPerson) < 2 || len(input.ContactPerson) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Contact person must be between 2 and 255 characters"), fiber.StatusBadRequest)
	}
	if len(input.Phone) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Phone must be 50 characters or less"), fiber.StatusBadRequest)
	}
	if len(input.CompanyName) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Company name must be 255 characters or less"), fiber.StatusBadRequest)
	}
	if len(input.TaxID) > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Tax ID must be 100 characters or less"), fiber.StatusBadRequest)
	}

	if input.CommissionRate < 0 || input.CommissionRate > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Commission rate must be between 0 and 100"), fiber.StatusBadRequest)
	}
	if input.Balance < 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Balance cannot be negative"), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateOrganizer(userID, input.Name, input.ContactPerson, input.Email, input.Phone, input.CompanyName, input.TaxID, input.BankAccountInfo, input.CommissionRate, input.Balance, input.Notes)
	if err != nil {
		if err.Error() == "user lacks create:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "organizer email already exists" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Organizer created successfully", true)
}
