package venues_controller

import (
	"github.com/gofiber/fiber/v2"
	venue_dto "ticket-zetu-api/modules/events/venues/dto"
)

// CreateVenue godoc
// @Summary Update Venue
// @Description Update Venue.
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Venue ID"
// @Param input body venue_dto.CreateVenueDto true "Venue details"
// @Success 200 {object} map[string]interface{} "Venue updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{id} [put]
func (c *VenueController) UpdateVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input venue_dto.CreateVenueDto

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	var imageURLs []string
	form, err := ctx.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for _, file := range files {
			if file.Size > 10*1024*1024 {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit"), fiber.StatusBadRequest)
			}
			if !isValidFileType(file.Header.Get("Content-Type")) {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid file type. Only images are allowed"), fiber.StatusBadRequest)
			}

			f, err := file.Open()
			if err != nil {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to open file"), fiber.StatusInternalServerError)
			}
			defer f.Close()

			url, err := c.cloudinary.UploadFile(ctx.Context(), f, "venues")
			if err != nil {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file to Cloudinary"), fiber.StatusInternalServerError)
			}
			imageURLs = append(imageURLs, url)
		}
	}

	newUpdateVenue := venue_dto.UpdateVenueDto{
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		City:        input.City,
		State:       input.State,
		Country:     input.Country,
		Capacity:    input.Capacity,
		ContactInfo: input.ContactInfo,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Status:      input.Status,
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
