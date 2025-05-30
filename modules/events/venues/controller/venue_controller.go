package venues_controller

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/venues/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type VenueController struct {
	service    service.VenueService
	logHandler *handler.LogHandler
	cloudinary *cloudinary.CloudinaryService
	validator  *validator.Validate
}

func NewVenueController(service service.VenueService, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) *VenueController {
	return &VenueController{
		service:    service,
		logHandler: logHandler,
		cloudinary: cloudinary,
		validator:  validator.New(),
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

func (c *VenueController) GetSingleVenueForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	// Exclude venue_images from fields as it's a relationship, not a column
	fields := ctx.Query("fields", "id,name,description,address,city,state,country,capacity,contact_info,latitude,longitude,status,organizer_id")

	venue, err := c.service.GetVenue(userID, id, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:venues permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, venue, "Venue retrieved successfully", true)
}

func (c *VenueController) GetVenuesForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	// Exclude venue_images from fields as it's a relationship, not a column
	fields := ctx.Query("fields", "id,name,description,address,city,state,country,capacity,contact_info,latitude,longitude,status,organizer_id")

	venues, err := c.service.GetVenues(userID, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:venues permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, venues, "Venues retrieved successfully", true)
}
