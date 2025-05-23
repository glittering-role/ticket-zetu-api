package venues_controller

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/venues/service"

	"github.com/gofiber/fiber/v2"
)

type VenueController struct {
	service    service.VenueService
	logHandler *handler.LogHandler
	cloudinary *cloudinary.CloudinaryService
}

func NewVenueController(service service.VenueService, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) *VenueController {
	return &VenueController{
		service:    service,
		logHandler: logHandler,
		cloudinary: cloudinary,
	}
}

func (c *VenueController) UpdateVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	var input struct {
		Name        string  `form:"name" validate:"required,min=2,max=255"`
		Description string  `form:"description" validate:"max=1000"`
		Address     string  `form:"address" validate:"required"`
		City        string  `form:"city" validate:"required,min=2,max=100"`
		State       string  `form:"state" validate:"max=100"`
		Country     string  `form:"country" validate:"required,min=2,max=100"`
		Capacity    int     `form:"capacity" validate:"gte=0"`
		ContactInfo string  `form:"contact_info" validate:"max=255"`
		Latitude    float64 `form:"latitude"`
		Longitude   float64 `form:"longitude"`
		Status      string  `form:"status" validate:"oneof=active inactive suspended"`
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
	if input.Status != "active" && input.Status != "inactive" && input.Status != "suspended" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Status must be one of active, inactive, suspended"), fiber.StatusBadRequest)
	}

	// Handle file uploads
	var imageURLs []string
	form, err := ctx.MultipartForm()
	if err == nil { // Only process files if multipart form is present
		files := form.File["images"]
		for _, file := range files {
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
			imageURLs = append(imageURLs, url)
		}
	}

	venue, err := c.service.UpdateVenue(
		userID,
		id,
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
		input.Status,
		imageURLs,
	)
	if err != nil {
		if err.Error() == "user lacks update:venues permission" {
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
	return c.logHandler.LogSuccess(ctx, venue, "Venue updated successfully", true)
}

func (c *VenueController) DeleteVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteVenue(userID, id)
	if err != nil {
		if err.Error() == "user lacks delete:venues permission" {
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
	return c.logHandler.LogSuccess(ctx, nil, "Venue deleted successfully", true)
}

func (c *VenueController) GetVenue(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	venue, err := c.service.GetVenue(userID, id)
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

func (c *VenueController) GetVenues(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	venues, err := c.service.GetVenues(userID)
	if err != nil {
		if err.Error() == "user lacks read:venues permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, venues, "Venues retrieved successfully", true)
}
