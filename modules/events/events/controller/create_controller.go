package controller

import (
	"strings"
	"ticket-zetu-api/modules/events/events/dto"

	"github.com/gofiber/fiber/v2"
)

// CreateEvent godoc
// @Summary Create a new Event
// @Description Creates a new event with its details, validates organizer status, subcategory, venue, and handles image associations.
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateEventInput true "Event details including venue, category, and optional images"
// @Success 200 {object} map[string]interface{} "Event created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload"
// @Failure 403 {object} map[string]interface{} "User lacks permission or organizer is inactive, flagged, or banned"
// @Failure 404 {object} map[string]interface{} "Subcategory or venue not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events [post]
func (c *EventController) CreateEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreateEvent

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Handle comma-separated tags
	tags := ctx.FormValue("tags")
	if tags != "" {
		input.Tags = strings.Split(strings.TrimSpace(tags), ",")
		for i, tag := range input.Tags {
			input.Tags[i] = strings.TrimSpace(tag)
			if input.Tags[i] == "" {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid tag format: empty tags are not allowed"), fiber.StatusBadRequest)
			}
		}
	}

	// Validate input
	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	newCreateEvent := dto.CreateEvent{
		Title:          input.Title,
		Description:    input.Description,
		SubcategoryID:  input.SubcategoryID,
		VenueID:        input.VenueID,
		StartTime:      input.StartTime,
		EndTime:        input.EndTime,
		Timezone:       input.Timezone,
		Language:       input.Language,
		EventType:      input.EventType,
		MinAge:         input.MinAge,
		TotalSeats:     input.TotalSeats,
		AvailableSeats: input.AvailableSeats,
		IsFree:         input.IsFree,
		HasTickets:     input.HasTickets,
		IsFeatured:     input.IsFeatured,
		Status:         input.Status,
		Tags:           input.Tags,
	}

	event, err := c.service.CreateEvent(
		newCreateEvent,
		userID,
	)

	if err != nil {
		switch err.Error() {
		case "user lacks create:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "organizer is not active":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "invalid subcategory ID format", "invalid venue ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, event, "Event created successfully", true)
}
