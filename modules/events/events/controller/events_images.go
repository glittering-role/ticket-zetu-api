package controller

import (
	"github.com/gofiber/fiber/v2"
)

// AddEventImage godoc
// @Summary Add an image to an event
// @Description Adds a single image to an event, with an optional primary flag. If set as primary, demotes existing primary image. Users can upload up to 5 images per event, one at a time in separate requests.
// @Tags Event Group
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Param image formData file true "Image file (max 10MB, JPEG, PNG, GIF, WEBP, MP4, MPEG, WEBM)"
// @Param is_primary formData boolean false "Set as primary image (default: false)"
// @Success 200 {object} map[string]interface{} "Image added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid form data, file type, file size, ID format, or exceeded image limit"
// @Failure 403 {object} map[string]interface{} "User lacks permission or organizer is inactive"
// @Failure 404 {object} map[string]interface{} "Event not found or not owned by organizer"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/images [post]
func (c *EventController) AddEventImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("event_id")

	form, err := ctx.MultipartForm()
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid form data"), fiber.StatusBadRequest)
	}

	files := form.File["image"]
	if len(files) != 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Exactly one image must be provided"), fiber.StatusBadRequest)
	}

	file := files[0]
	if file.Size > 10*1024*1024 {
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

	url, err := c.cloudinary.UploadFile(ctx.Context(), f, "event_images")
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file to Cloudinary"), fiber.StatusInternalServerError)
	}

	isPrimary := ctx.FormValue("is_primary", "false") == "true"

	image, err := c.service.AddEventImage(userID, eventID, url, isPrimary)
	if err != nil {
		switch err.Error() {
		case "user lacks create:event_images permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "maximum 5 images allowed per event":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "invalid event ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, image, "Event image added successfully", true)
}

// DeleteEventImage godoc
// @Summary Delete an event image
// @Description Deletes an event image from both database (soft delete) and cloud storage using event and image IDs.
// @Tags Event Group
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Param image_id path string true "Image ID"
// @Success 200 {object} map[string]interface{} "Event image deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Event or image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/images/{image_id} [delete]
func (c *EventController) DeleteEventImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("event_id")
	imageID := ctx.Params("image_id")

	err := c.service.DeleteEventImage(userID, eventID, imageID)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:event_images permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "event image not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "invalid event ID format", "invalid image ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Event image deleted successfully", true)
}

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
