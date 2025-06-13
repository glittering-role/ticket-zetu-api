package tickets_controller

import (
	"ticket-zetu-api/modules/tickets/price_tires/dto"

	"github.com/gofiber/fiber/v2"
)

// UpdatePriceTier godoc
// @Summary Update a price tier
// @Description Update an existing price tier
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Price Tier ID"
// @Param input body dto.UpdatePriceTierRequest true "Price tier update data"
// @Success 200 {object} map[string]interface{} "Price tier updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Price tier not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers/{id} [put]
func (c *PriceTierController) UpdatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input dto.UpdatePriceTierRequest
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdatePriceTier(userID, id, input)
	if err != nil {
		switch err.Error() {
		case "user lacks update:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Price tier updated successfully", true)
}

// DeletePriceTier godoc
// @Summary Delete a price tier
// @Description Delete an existing price tier
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Price Tier ID"
// @Success 200 {object} map[string]interface{} "Price tier deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid price tier ID or cannot delete default price tier"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Price tier not found"
// @Failure 409 {object} map[string]interface{} "Price tier is in use"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers/{id} [delete]
func (c *PriceTierController) DeletePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeletePriceTier(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "cannot delete default price tier":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "price tier is in use by events":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Price tier deleted successfully", true)
}
