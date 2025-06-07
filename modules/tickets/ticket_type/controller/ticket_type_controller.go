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

// GetAllTicketTypesForOrganization godoc
// @Summary Get all ticket types for an organization
// @Description Retrieves all ticket types belonging to the user's organization
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param fields query string false "Comma-separated list of fields to include in response" example(name,description,price_modifier,status)
// @Success 200 {object} map[string]interface{} "List of ticket types"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/organization [get]
func (c *TicketTypeController) GetAllTicketTypesForOrganization(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	fields := ctx.Query("fields", "")

	ticketTypes, err := c.service.GetAllTicketTypesForOrganization(userID, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:ticket_types permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, ticketTypes, "Organization ticket types retrieved successfully", true)
}

// GetTicketTypesForEvent godoc
// @Summary Get ticket types for an event
// @Description Retrieves all ticket types for a specific event
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Param fields query string false "Comma-separated list of fields to include in response" default(id,name,price_modifier,status,is_default,sales_start,sales_end)
// @Success 200 {object} map[string]interface{} "List of ticket types"
// @Failure 400 {object} map[string]interface{} "Invalid event ID"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Event not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/event/{event_id} [get]
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

// GetSingleTicketType godoc
// @Summary Get a single ticket type
// @Description Retrieves details of a specific ticket type
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Ticket Type ID"
// @Param fields query string false "Comma-separated list of fields to include in response" default(id,name,description,price_modifier,benefits,max_tickets_per_user,status,is_default,sales_start,sales_end,quantity_available,min_tickets_per_user)
// @Success 200 {object} map[string]interface{} "Ticket type details"
// @Failure 400 {object} map[string]interface{} "Invalid ticket type ID"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Ticket type not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/{id} [get]
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

// DeleteTicketType godoc
// @Summary Delete a ticket type
// @Description Deletes a specific ticket type (cannot delete default or in-use ticket types)
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Ticket Type ID"
// @Success 200 {object} map[string]interface{} "Ticket type deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ticket type ID or cannot delete default ticket type"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Ticket type not found"
// @Failure 409 {object} map[string]interface{} "Ticket type is in use"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/{id} [delete]
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
