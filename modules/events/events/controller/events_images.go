package controller

import (
	"github.com/gofiber/fiber/v2"
)

func (c *EventController) AddEventImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("id")

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
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid file type. Only images are allowed"), fiber.StatusBadRequest)
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
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, image, "Event image added successfully", true)
}

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
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
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
