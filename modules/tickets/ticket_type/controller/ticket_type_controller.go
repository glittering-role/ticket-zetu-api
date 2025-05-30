package ticket_type_controller

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/tickets/ticket_type/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TicketTypeController struct {
	service    ticket_type_service.TicketTypeService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

func NewTicketTypeController(service ticket_type_service.TicketTypeService, logHandler *handler.LogHandler) *TicketTypeController {
	return &TicketTypeController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

func (c *TicketTypeController) GetTicketTypesForEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("event_id")
	fields := ctx.Query("fields", "id,name,price_modifier,status,is_default,sales_start,sales_end")

	ticketTypes, err := c.service.GetTicketTypes(userID, eventID, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, ticketTypes, "Ticket types retrieved successfully", true)
}

func (c *TicketTypeController) GetSingleTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	fields := ctx.Query("fields", "id,name,description,price_modifier,benefits,max_tickets_per_user,status,is_default,sales_start,sales_end,quantity_available,min_tickets_per_user")

	ticketType, err := c.service.GetTicketType(userID, id, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "ticket type not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, ticketType, "Ticket type retrieved successfully", true)
}

func (c *TicketTypeController) DeleteTicketType(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteTicketType(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "ticket type not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "cannot delete default ticket type":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "ticket type is in use":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Ticket type deleted successfully", false)
}
