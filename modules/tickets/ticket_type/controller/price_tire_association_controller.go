package ticket_type_controller

import (
	"ticket-zetu-api/modules/tickets/ticket_type/dto"

	"github.com/gofiber/fiber/v2"
)

// AssociatePriceTier godoc
// @Summary Associate a PriceTier with a TicketType
// @Description Associates an existing PriceTier with a TicketType by adding a link in the join table.
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ticket_type_id path string true "TicketType ID"
// @Param input body dto.AssociatePriceTierInput true "PriceTier ID to associate"
// @Success 200 {object} map[string]interface{} "PriceTier associated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or IDs"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "TicketType or PriceTier not found"
// @Failure 409 {object} map[string]interface{} "PriceTier already associated"
// @Failure 422 {object} map[string]interface{} "PriceTier is invalid (inactive, expired, or incompatible)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/{ticket_type_id}/price-tiers [post]
func (c *TicketTypeController) AssociatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	ticketTypeID := ctx.Params("ticket_type_id")

	var input dto.AssociatePriceTierInput
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	priceTier, err := c.service.AssociatePriceTier(userID, ticketTypeID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks create:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "ticket type not found", "price tier not found or not accessible":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid ticket type ID format", "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "price tier already associated with ticket type":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "price tier must be active", "price tier is not yet effective", "price tier has expired",
			"price tier max_tickets is less than ticket type max_tickets_per_user",
			"price tier min_tickets is greater than ticket type min_tickets_per_user":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusUnprocessableEntity, err.Error()), fiber.StatusUnprocessableEntity)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier associated successfully", false)
}

// DisassociatePriceTier godoc
// @Summary Disassociate a PriceTier from a TicketType
// @Description Removes the association between a TicketType and a PriceTier from the join table.
// @Tags TicketType Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ticket_type_id path string true "TicketType ID"
// @Param price_tier_id path string true "PriceTier ID"
// @Success 200 {object} map[string]interface{} "PriceTier disassociated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid IDs"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "TicketType, PriceTier, or association not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ticket-types/{ticket_type_id}/price-tiers/{price_tier_id} [delete]
func (c *TicketTypeController) DisassociatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	ticketTypeID := ctx.Params("ticket_type_id")
	priceTierID := ctx.Params("price_tier_id")

	if err := c.service.DisassociatePriceTier(userID, ticketTypeID, priceTierID); err != nil {
		switch err.Error() {
		case "user lacks delete:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "ticket type not found", "price tier not found or not accessible", "price tier not associated with ticket type":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid ticket type ID format", "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Price tier disassociated successfully", false)
}
