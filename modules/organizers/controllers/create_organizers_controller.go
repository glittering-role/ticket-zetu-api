package organizers

import (
	"github.com/gofiber/fiber/v2"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
)

// CreateOrganizer godoc
// @Summary Create a new organizer
// @Description Creates a new organizer with the provided details
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body organizer_dto.CreateOrganizerRequest true "Organizer details"
// @Success 200 {object} map[string]interface{} "Organizer created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 409 {object} map[string]interface{} "Organizer email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers [post]
func (c *OrganizerController) CreateOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input organizer_dto.CreateOrganizerRequest
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

	_, err := c.service.CreateOrganizer(
		userID,
		input.Name,
		input.ContactPerson,
		input.Email,
		input.Phone,
		input.CompanyName,
		input.TaxID,
		input.BankAccountInfo,
		input.CommissionRate,
		input.Balance,
		input.Notes,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks create:organizers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer email already exists":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organizer created successfully", true)
}
