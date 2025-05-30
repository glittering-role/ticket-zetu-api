package ticket_type_controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (c *TicketTypeController) CreateTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input struct {
		EventID           string     `json:"event_id" validate:"required,uuid"`
		Name              string     `json:"name" validate:"required,min=2,max=100"`
		Description       string     `json:"description"`
		PriceModifier     float64    `json:"price_modifier" validate:"required,min=0"`
		Benefits          string     `json:"benefits"`
		MaxTicketsPerUser int        `json:"max_tickets_per_user" validate:"required,min=1"`
		Status            string     `json:"status" validate:"omitempty,oneof=active inactive archived"`
		IsDefault         bool       `json:"is_default"`
		SalesStart        time.Time  `json:"sales_start" validate:"required"`
		SalesEnd          *time.Time `json:"sales_end"`
		QuantityAvailable *int       `json:"quantity_available" validate:"omitempty,min=0"`
		MinTicketsPerUser int        `json:"min_tickets_per_user" validate:"min=1"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	ticketType, err := c.service.CreateTicketType(
		userID,
		input.EventID,
		input.Name,
		input.Description,
		input.PriceModifier,
		input.Benefits,
		input.MaxTicketsPerUser,
		input.Status,
		input.IsDefault,
		input.SalesStart,
		input.SalesEnd,
		input.QuantityAvailable,
		input.MinTicketsPerUser,
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

func (c *TicketTypeController) UpdateTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input struct {
		Name              string     `json:"name" validate:"omitempty,min=2,max=100"`
		Description       string     `json:"description"`
		PriceModifier     float64    `json:"price_modifier" validate:"omitempty,min=0"`
		Benefits          string     `json:"benefits"`
		MaxTicketsPerUser int        `json:"max_tickets_per_user" validate:"omitempty,min=1"`
		Status            string     `json:"status" validate:"omitempty,oneof=active inactive archived"`
		IsDefault         *bool      `json:"is_default"`
		SalesStart        time.Time  `json:"sales_start"`
		SalesEnd          *time.Time `json:"sales_end"`
		QuantityAvailable *int       `json:"quantity_available" validate:"omitempty,min=0"`
		MinTicketsPerUser int        `json:"min_tickets_per_user" validate:"omitempty,min=1"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	ticketType, err := c.service.UpdateTicketType(
		userID,
		id,
		input.Name,
		input.Description,
		input.PriceModifier,
		input.Benefits,
		input.MaxTicketsPerUser,
		input.Status,
		input.IsDefault,
		input.SalesStart,
		input.SalesEnd,
		input.QuantityAvailable,
		input.MinTicketsPerUser,
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
