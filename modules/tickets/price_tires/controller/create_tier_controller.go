package tickets_controller

import (
	"ticket-zetu-api/modules/tickets/price_tires/dto"

	"github.com/gofiber/fiber/v2"
)

// CreatePriceTier godoc
// @Summary Create a new price tier
// @Description Create a new price tier for the organizer
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreatePriceTierData true "Price tier data"
// @Success 200 {object} map[string]interface{}  "Price tier created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers [post]
func (c *PriceTierController) CreatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreatePriceTierRequest
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	priceTier, err := c.service.CreatePriceTier(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks create:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer is not active":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier created successfully", true)
}
