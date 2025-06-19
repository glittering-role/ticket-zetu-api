package discount_controller

import (
	"ticket-zetu-api/modules/tickets/discount/dto"

	"github.com/gofiber/fiber/v2"
)

// CreateDiscount godoc
// @Summary Create a new discount
// @Description Create a new discount code for the organizer
// @Tags Discounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateDiscountCodeInput true "Discount data"
// @Success 200 {object} map[string]interface{} "Discount created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or organizer/event not found"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /discounts [post]
func (c *DiscountController) CreateDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreateDiscountCodeInput
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	createdDiscount, err := c.service.CreateDiscount(userID, &input)
	if err != nil {
		switch err.Error() {
		case "user lacks create:discounts permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "event not found or not owned by organizer":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, createdDiscount, "Discount created successfully", true)
}
