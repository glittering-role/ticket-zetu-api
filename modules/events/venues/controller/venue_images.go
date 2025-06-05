package venues_controller

import (
	"github.com/gofiber/fiber/v2"
)

// isValidFileType checks if the file type is an image or video
func isValidFileType(contentType string) bool {
	validTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"video/mp4", "video/mpeg", "video/webm",
	}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// AddVenueImage godoc
// @Summary Add Venue Image
// @Description Add an image to a venue.
// @Tags Venue Group
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Venue ID"
// @Param is_primary formData bool false "Is this image primary?"
// @Param image formData file true "Image file to upload"
// @Success 200 {object} map[string]interface{} "Venue image added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or file"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{id}/images [post]
func (c *VenueController) AddVenueImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	venueID := ctx.Params("id")
	var input struct {
		IsPrimary bool `form:"is_primary"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Handle file upload
	file, err := ctx.FormFile("image")
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Failed to parse file"), fiber.StatusBadRequest)
	}

	if file.Size > 10*1024*1024 { // Limit file size to 10MB
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit"), fiber.StatusBadRequest)
	}
	if !isValidFileType(file.Header.Get("Content-Type")) {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid file type. Only images and videos are allowed"), fiber.StatusBadRequest)
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

	venueImage, err := c.service.AddVenueImage(userID, venueID, url, input.IsPrimary)
	if err != nil {
		if err.Error() == "user lacks create:venue_images permission" {
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
	return c.logHandler.LogSuccess(ctx, venueImage, "Venue image added successfully", true)
}

// DeleteVenueImage godoc
// @Summary Delete Venue Image
// @Description Delete an image from a venue.
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param venue_id path string true "Venue ID"
// @Param image_id path string true "Image ID"
// @Success 200 {object} map[string]interface{} "Venue image deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Venue or image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{venue_id}/images/{image_id} [delete]
func (c *VenueController) DeleteVenueImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	venueID := ctx.Params("venue_id")
	imageID := ctx.Params("image_id")

	err := c.service.DeleteVenueImage(userID, venueID, imageID)
	if err != nil {
		if err.Error() == "user lacks delete:venue_images permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "venue not found" || err.Error() == "venue image not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Venue image deleted successfully", true)
}
