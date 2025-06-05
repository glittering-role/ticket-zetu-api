package venues_controller

import "github.com/gofiber/fiber/v2"

// Read Venues godoc
// @Summary Get Venue
// @Description Retrieve a venue by ID.
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param fields query string false "Fields to include in the response, comma-separated"
// @Success 200 {object} map[string]interface{} "Venue retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/all [get]
func (c *VenueController) GetAllVenue(ctx *fiber.Ctx) error {
	// _ := ctx.Locals("user_id").(string)
	fields := ctx.Query("fields", "id,name,description,address,city,state,country,capacity,contact_info,latitude,longitude,status,organizer_id,venue_images")

	venue, err := c.service.GetAllVenues(fields)
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
