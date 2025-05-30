package venues_controller

import "github.com/gofiber/fiber/v2"

func (c *VenueController) GetVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	fields := ctx.Query("fields", "id,name,description,address,city,state,country,capacity,contact_info,latitude,longitude,status,organizer_id,venue_images")

	venue, err := c.service.GetVenues(userID, fields)
	if err != nil {
		if err.Error() == "user lacks read:venues permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "venue not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, venue, "Venue retrieved successfully", true)
}
