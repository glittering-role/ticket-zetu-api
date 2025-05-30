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
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "invalid event ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, event, "Event retrieved successfully", true)
}

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
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, events, "Events retrieved successfully", true)
}
