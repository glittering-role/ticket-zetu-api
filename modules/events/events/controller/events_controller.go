package controller

import (
	"strconv"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/events/service"
	"time"

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
// @Description Retrieves detailed information about a specific event by ID for the authenticated organizer.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID" Format(uuid)
// @Success 200 {object} map[string]interface{} "Event retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid event ID format"
// @Failure 403 {object} map[string]interface{} "User lacks read:events permission"
// @Failure 404 {object} map[string]interface{} "Event or organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{id} [get]
func (c *EventController) GetSingleEventForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	event, err := c.service.GetEvent(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks read:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found", "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid event ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Internal server error"), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, event, "Event retrieved successfully", true)
}

// GetEventsForOrganizer godoc
// @Summary Retrieve all events for an organizer
// @Description Retrieves a paginated list of events for the authenticated organizer. Supports pagination via query parameters.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query integer false "Page number (default: 1)" Minimum(1)
// @Param page_size query integer false "Items per page (default: 20, max: 100)" Minimum(1) Maximum(100)
// @Success 200 {object} map[string]interface{} "Events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 403 {object} map[string]interface{} "User lacks read:events permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events [get]
func (c *EventController) GetEventsForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Parse pagination parameters
	var filter service.SearchFilter
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page number"), fiber.StatusBadRequest)
	}
	filter.Page = page

	pageSize, err := strconv.Atoi(ctx.Query("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page_size. Must be between 1 and 100"), fiber.StatusBadRequest)
	}
	filter.PageSize = pageSize

	events, err := c.service.GetEvents(userID, filter)
	if err != nil {
		switch err.Error() {
		case "user lacks read:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Internal server error"), fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, events, "Events retrieved successfully", true)
}

// SearchEvents godoc
// @Summary Search and filter events for an organizer
// @Description Searches and filters events for the authenticated organizer based on query parameters. Supports pagination.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param query query string false "Search term for event title or description"
// @Param start_date query string false "Start date filter (ISO 8601)" Format(date-time)
// @Param end_date query string false "End date filter (ISO 8601)" Format(date-time)
// @Param event_type query string false "Event type filter" Enums(online, offline, hybrid)
// @Param is_free query boolean false "Filter by free or paid events"
// @Param status query string false "Event status filter" Enums(published, draft, cancelled)
// @Param min_price query number false "Minimum ticket price filter"
// @Param max_price query number false "Maximum ticket price filter"
// @Param page query integer false "Page number (default: 1)" Minimum(1)
// @Param page_size query integer false "Items per page (default: 20, max: 100)" Minimum(1) Maximum(100)
// @Success 200 {object} map[string]interface{} "Events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 403 {object} map[string]interface{} "User lacks read:events permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/search [get]
func (c *EventController) SearchEvents(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Parse query parameters
	var filter service.SearchFilter

	// Search query
	filter.Query = ctx.Query("query")

	// Date filters
	if startDate := ctx.Query("start_date"); startDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid start_date format"), fiber.StatusBadRequest)
		}
		filter.StartDate = &parsedDate
	}
	if endDate := ctx.Query("end_date"); endDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid end_date format"), fiber.StatusBadRequest)
		}
		filter.EndDate = &parsedDate
	}

	// Event type
	if eventType := ctx.Query("event_type"); eventType != "" {
		if eventType != "online" && eventType != "offline" && eventType != "hybrid" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid event_type. Must be one of: online, offline, hybrid"), fiber.StatusBadRequest)
		}
		filter.EventType = eventType
	}

	// Is free
	if isFree := ctx.Query("is_free"); isFree != "" {
		parsedBool, err := strconv.ParseBool(isFree)
		if err != nil {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid is_free format"), fiber.StatusBadRequest)
		}
		filter.IsFree = &parsedBool
	}

	// Status
	if status := ctx.Query("status"); status != "" {
		if status != "published" && status != "draft" && status != "cancelled" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid status. Must be one of: published, draft, cancelled"), fiber.StatusBadRequest)
		}
		filter.Status = status
	}

	// Price filters
	if minPrice := ctx.Query("min_price"); minPrice != "" {
		parsedPrice, err := strconv.ParseFloat(minPrice, 64)
		if err != nil || parsedPrice < 0 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid min_price format or negative value"), fiber.StatusBadRequest)
		}
		filter.MinPrice = &parsedPrice
	}
	if maxPrice := ctx.Query("max_price"); maxPrice != "" {
		parsedPrice, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil || parsedPrice < 0 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid max_price format or negative value"), fiber.StatusBadRequest)
		}
		filter.MaxPrice = &parsedPrice
	}

	// Pagination
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page number"), fiber.StatusBadRequest)
	}
	filter.Page = page

	pageSize, err := strconv.Atoi(ctx.Query("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page_size. Must be between 1 and 100"), fiber.StatusBadRequest)
	}
	filter.PageSize = pageSize

	// Call service
	result, err := c.service.SearchEvents(userID, filter)
	if err != nil {
		switch err.Error() {
		case "user lacks read:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Internal server error"), fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, result, "Events retrieved successfully", true)
}
