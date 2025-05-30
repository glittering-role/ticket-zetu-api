package discount_controller

import (
	"ticket-zetu-api/modules/tickets/models/tickets"

	"github.com/gofiber/fiber/v2"
)

func (c *DiscountController) CreateDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var discount tickets.DiscountCode
	if err := ctx.BodyParser(&discount); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	createdDiscount, err := c.service.CreateDiscount(userID, &discount)
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
