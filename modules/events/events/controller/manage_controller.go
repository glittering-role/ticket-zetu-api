package controller

import (
	"strings"
	"ticket-zetu-api/modules/events/events/dto"

	"github.com/gofiber/fiber/v2"
)

// UpdateEvent godoc
// @Summary Update an existing event
// @Description Updates an event with the provided details
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Param input body dto.UpdateEvent true "Event update details"
// @Success 200 {object} map[string]interface{} "Event updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Event, venue or subcategory not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{id} [put]
func (c *EventController) UpdateEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	// Parse request body into dto.UpdateEvent
	var input dto.UpdateEvent
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	// Parse and validate tags from form data
	tags := ctx.FormValue("tags")
	if tags != "" {
		*input.Tags = strings.Split(strings.TrimSpace(tags), ",")
		for i, tag := range *input.Tags {
			(*input.Tags)[i] = strings.TrimSpace(tag)
			if (*input.Tags)[i] == "" {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid tag format: empty tags are not allowed"), fiber.StatusBadRequest)
			}
		}
	}

	// Validate input
	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	event, err := c.service.UpdateEvent(input, userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks update:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found or not owned by organizer":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found", "venue not found", "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "invalid event ID format", "invalid subcategory ID format", "invalid venue ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, event, "Event updated successfully", true)
}

// DeleteEvent godoc
// @Summary Delete an event
// @Description Deletes an event and its associated resources
// @Tags Event Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Success 204 {object} map[string]interface{} "Event deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid event ID"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Event not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{id} [delete]
func (c *EventController) DeleteEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteEvent(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "cannot delete an active event":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Event deleted successfully", true)
}
