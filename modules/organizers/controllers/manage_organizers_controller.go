package organizers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
)

// UpdateOrganizer godoc
// @Summary Update an organizer
// @Description Updates an existing organizer's details
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Param input body organizer_dto.UpdateOrganizerRequest true "Organizer details"
// @Success 200 {object} map[string]interface{} "Organizer updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 409 {object} map[string]interface{} "Organizer email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id} [put]
func (c *OrganizerController) UpdateOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	var input organizer_dto.UpdateOrganizerRequest
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Basic field validation
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

	// Update logic
	_, err := c.service.UpdateOrganizer(
		userID, id,
		input.Name, input.ContactPerson, input.Email, input.Phone,
		input.CompanyName, input.TaxID, input.BankAccountInfo,
		input.CommissionRate, input.Balance, input.Notes, input.AllowSubscriptions,
	)

	if err != nil {
		switch err.Error() {
		case "invalid organizer ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks update:organizers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer email already exists":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organizer updated successfully", true)
}

// DeleteOrganizer godoc
// @Summary Delete an organizer
// @Description Deletes an inactive organizer
// @Tags Organizers
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format or organizer is active"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id} [delete]
func (c *OrganizerController) DeleteOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.DeleteOrganizer(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks delete:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "cannot delete an active organizer" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Organizer deleted successfully", true)
}

// DeactivateOrganizer godoc
// @Summary Deactivate an organizer
// @Description Sets an organizer's status to inactive
// @Tags Organizers
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer deactivated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format or already inactive"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id}/deactivate [patch]
func (c *OrganizerController) DeactivateOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.DeactivateOrganizer(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks update:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "organizer is already inactive" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Organizer deactivated successfully", true)
}

// ToggleOrganizerStatus godoc
// @Summary Toggle organizer status
// @Description Toggles the active/inactive status of an organizer
// @Tags Organizers
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer status toggled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id}/toggle-status [patch]
func (c *OrganizerController) ToggleOrganizerStatus(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.ToggleOrganizationsStatus(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks update:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organizer status toggled successfully", true)
}

// FlagOrganizer godoc
// @Summary Flag or unflag an organizer
// @Description Marks or unmarks an organizer as flagged
// @Tags Organizers
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer flag status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id}/flag [patch]
func (c *OrganizerController) FlagOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.FlagOrganization(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks update:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organizer flag status updated successfully", true)
}

// BanOrganizer godoc
// @Summary Ban or unban an organizer
// @Description Toggles the banned status of an organizer
// @Tags Organizers
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer ban status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id}/ban [patch]
func (c *OrganizerController) BanOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.BanOrganization(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks update:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organizer ban status updated successfully", true)
}
