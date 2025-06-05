package venues_controller

import (
	venue_dto "ticket-zetu-api/modules/events/venues/dto"

	"github.com/gofiber/fiber/v2"
)

// CreateVenue godoc
// @Summary Create Venue
// @Description Create a new venue.
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body venue_dto.CreateVenueDto true "Venue details"
// @Success 201 {object} map[string]interface{} "Venue created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues [post]
func (c *VenueController) CreateVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input venue_dto.CreateVenueDto

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Basic validation
	if len(input.Name) < 2 || len(input.Name) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 255 characters"), fiber.StatusBadRequest)
	}
	if input.Address == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Address cannot be empty"), fiber.StatusBadRequest)
	}
	if len(input.City) < 2 || len(input.City) > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "City must be between 2 and 100 characters"), fiber.StatusBadRequest)
	}
	if len(input.State) > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "State must be 100 characters or less"), fiber.StatusBadRequest)
	}
	if len(input.Country) < 2 || len(input.Country) > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Country must be between 2 and 100 characters"), fiber.StatusBadRequest)
	}
	if input.Capacity < 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Capacity cannot be negative"), fiber.StatusBadRequest)
	}
	if len(input.ContactInfo) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Contact info must be 255 characters or less"), fiber.StatusBadRequest)
	}
	if len(input.Description) > 1000 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Description must be 1000 characters or less"), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateVenue(
		userID,
		input.Name,
		input.Description,
		input.Address,
		input.City,
		input.State,
		input.Country,
		input.Capacity,
		input.ContactInfo,
		input.Latitude,
		input.Longitude,
	)
	if err != nil {
		if err.Error() == "user lacks create:venues permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Venue created successfully", true)
}
