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

func (c *VenueController) CreateVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	var input struct {
		Name        string  `json:"name" validate:"required,min=2,max=255"`
		Description string  `json:"description" validate:"max=1000"`
		Address     string  `json:"address" validate:"required"`
		City        string  `json:"city" validate:"required,min=2,max=100"`
		State       string  `json:"state" validate:"max=100"`
		Country     string  `json:"country" validate:"required,min=2,max=100"`
		Capacity    int     `json:"capacity" validate:"gte=0"`
		ContactInfo string  `json:"contact_info" validate:"max=255"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	}

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

	venue, err := c.service.CreateVenue(
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
	return c.logHandler.LogSuccess(ctx, venue, "Venue created successfully", true)
}

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
