package ticket_type_controller

import (
	"github.com/gofiber/fiber/v2"
	"ticket-zetu-api/modules/tickets/ticket_type/dto"
)

// CreateTicketType godoc
// @Summary Create a new TicketType
// @Description Creates a new TicketType.
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateTicketTypeInputData true "TicketType details"
// @Success 200 {object} map[string]interface{} "TicketType created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types [post]
func (c *TicketTypeController) CreateTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreateTicketTypeInput

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	// Import the dto package at the top if not already imported:

	createInput := dto.CreateTicketTypeInput{
		EventID:           input.EventID,
		Name:              input.Name,
		Description:       input.Description,
		PriceModifier:     input.PriceModifier,
		Benefits:          input.Benefits,
		MaxTicketsPerUser: input.MaxTicketsPerUser,
		Status:            input.Status,
		IsDefault:         input.IsDefault,
		SalesStart:        input.SalesStart,
		SalesEnd:          input.SalesEnd,
		QuantityAvailable: input.QuantityAvailable,
		MinTicketsPerUser: input.MinTicketsPerUser,
	}

	ticketType, err := c.service.CreateTicketType(
		userID,
		createInput,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks create:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, ticketType, "Ticket type created successfully", false)
}

// UpdateTicketType godoc
// @Summary Update TicketType
// @Description Update TicketType.
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "TicketType ID"
// @Param input body dto.UpdateTicketTypeInput true "TicketType details"
// @Success 200 {object} map[string]interface{} "TicketType updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/{id} [put]
func (c *TicketTypeController) UpdateTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input dto.UpdateTicketTypeInput

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	updateInput := dto.UpdateTicketTypeInput{
		Name:              input.Name,
		Description:       input.Description,
		PriceModifier:     input.PriceModifier,
		Benefits:          input.Benefits,
		MaxTicketsPerUser: input.MaxTicketsPerUser,
		Status:            input.Status,
		IsDefault:         input.IsDefault,
		SalesStart:        input.SalesStart,
		SalesEnd:          input.SalesEnd,
		QuantityAvailable: input.QuantityAvailable,
		MinTicketsPerUser: input.MinTicketsPerUser,
	}

	ticketType, err := c.service.UpdateTicketType(
		userID,
		id,
		updateInput,
	)

	if err != nil {
		switch err.Error() {
		case "user lacks update:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "ticket type not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, ticketType, "Ticket type updated successfully", false)
}
