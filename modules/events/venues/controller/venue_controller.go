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

// GetSingleVenueForOrganizer godoc
// @Summary Get Venue details for Organizer
// @Description Retrieves details of a specific Venue by its ID for the Organizer
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Venue ID"
// @Param fields query string false "Comma-separated list of fields to include in the response"
// @Success 200 {object} map[string]interface{} "Venue retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid Venue ID format"
// @Failure 403 {object} map[string]interface{} "User lacks view permission"
// @Failure 404 {object} map[string]interface{} "Venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues/{id} [get]
func (c *VenueController) GetSingleVenueForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

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

// GetVenuesForOrganizer godoc
// @Summary Get all Venues for Organizer
// @Description Retrieves all Venues for the Organizer
// @Tags Venue Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param fields query string false "Comma-separated list of fields to include in the response"
// @Success 200 {object} map[string]interface{} "Venues retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /venues [get]
func (c *VenueController) GetVenuesForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

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
