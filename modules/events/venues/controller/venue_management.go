package venues_controller

import (
	"strings"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"

	"github.com/gofiber/fiber/v2"
)

// UpdateVenue godoc
// @Summary Update Venue
// @Description Update an existing venue.
// @Tags Venue Group
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Venue ID"
// @Param input body venue_dto.UpdateVenueDto true "Venue details"
// @Param images formData file false "Venue images (optional)"
// @Success 200 {object} map[string]interface{} "Venue updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{id} [put]
func (c *VenueController) UpdateVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input venue_dto.UpdateVenueDto

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	// Validate array fields (Layout, AccessibilityFeatures, Facilities)
	if input.Layout != "" {
		layouts := strings.Split(input.Layout, ",")
		for _, l := range layouts {
			if len(strings.TrimSpace(l)) > 100 {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Each layout item must be 100 characters or less"), fiber.StatusBadRequest)
			}
		}
	}
	if input.AccessibilityFeatures != "" {
		features := strings.Split(input.AccessibilityFeatures, ",")
		for _, f := range features {
			if len(strings.TrimSpace(f)) > 100 {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Each accessibility feature must be 100 characters or less"), fiber.StatusBadRequest)
			}
		}
	}
	if input.Facilities != "" {
		facilities := strings.Split(input.Facilities, ",")
		for _, f := range facilities {
			if len(strings.TrimSpace(f)) > 100 {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Each facility must be 100 characters or less"), fiber.StatusBadRequest)
			}
		}
	}

	// Map all fields to UpdateVenueDto
	newUpdateVenue := venue_dto.UpdateVenueDto{
		Name:                  input.Name,
		Description:           input.Description,
		Address:               input.Address,
		City:                  input.City,
		State:                 input.State,
		PostalCode:            input.PostalCode,
		Country:               input.Country,
		Capacity:              input.Capacity,
		VenueType:             input.VenueType,
		Layout:                input.Layout,
		AccessibilityFeatures: input.AccessibilityFeatures,
		Facilities:            input.Facilities,
		ContactInfo:           input.ContactInfo,
		Timezone:              input.Timezone,
		Latitude:              input.Latitude,
		Longitude:             input.Longitude,
		Status:                input.Status,
	}

	venue, err := c.service.UpdateVenue(
		userID,
		id,
		newUpdateVenue,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks update:venues permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, venue, "Venue updated successfully", true)
}

// DeleteVenue godoc
// @Summary Delete Venue
// @Description Delete Venue.
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Venue ID"
// @Success 200 {object} map[string]interface{} "Venue deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{id} [delete]
func (c *VenueController) DeleteVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteVenue(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:venues permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "cannot delete an active venue":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Venue deleted successfully", true)
}
