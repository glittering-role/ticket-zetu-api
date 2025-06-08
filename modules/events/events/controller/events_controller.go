package controller

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/events/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type EventController struct {
	service    service.EventService
	logHandler *handler.LogHandler
	cloudinary *cloudinary.CloudinaryService
	validator  *validator.Validate
}

func NewEventController(service service.EventService, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) *EventController {
	return &EventController{
		service:    service,
		logHandler: logHandler,
		cloudinary: cloudinary,
		validator:  validator.New(),
	}
}

// GetSingleEventForOrganizer godoc
// @Summary Retrieve a single event for an organizer
// @Description Retrieves details of a specific event by ID for the authenticated organizer, with optional field selection.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Param fields query string false "Comma-separated list of fields to include (e.g., id,title,subcategory_id,description,venue_id,total_seats,available_seats,start_time,end_time,price_tier_id,base_price,is_featured,status)"
// @Success 200 {object} map[string]interface{} "Event retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid event ID format or invalid request"
// @Failure 403 {object} map[string]interface{} "User lacks read:events permission"
// @Failure 404 {object} map[string]interface{} "Event or organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{id} [get]
func (c *EventController) GetSingleEventForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	// Exclude event_images and metadata fields; removed category_id
	fields := ctx.Query("fields", "id,title,subcategory_id,description,venue_id,total_seats,available_seats,start_time,end_time,price_tier_id,base_price,is_featured,status")

	event, err := c.service.GetEvent(userID, id, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid event ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, event, "Event retrieved successfully", true)
}

// GetEventsForOrganizer godoc
// @Summary Retrieve all events for an organizer
// @Description Retrieves a list of events for the authenticated organizer, with optional field selection.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param fields query string false "Comma-separated list of fields to include (e.g., id,title,subcategory_id,description,venue_id,total_seats,available_seats,start_time,end_time,price_tier_id,base_price,is_featured,status)"
// @Success 200 {object} map[string]interface{} "Events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "User lacks read:events permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events [get]
func (c *EventController) GetEventsForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	// Exclude event_images and metadata fields; removed category_id
	fields := ctx.Query("fields", "id,title,subcategory_id,description,venue_id,total_seats,available_seats,start_time,end_time,price_tier_id,base_price,is_featured,status")

	events, err := c.service.GetEvents(userID, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, events, "Events retrieved successfully", true)
}
