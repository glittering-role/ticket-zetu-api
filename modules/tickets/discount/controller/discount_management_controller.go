package discount_controller

import (
	"ticket-zetu-api/modules/tickets/models/tickets"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (c *DiscountController) GetDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid discount ID format"), fiber.StatusBadRequest)
	}

	discount, err := c.service.GetDiscount(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks read:discounts permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "discount not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, discount, "Discount retrieved successfully", true)
}

func (c *DiscountController) GetDiscounts(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	discounts, err := c.service.GetDiscounts(userID)
	if err != nil {
		switch err.Error() {
		case "user lacks read:discounts permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, discounts, "Discounts retrieved successfully", true)
}

func (c *DiscountController) UpdateDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid discount ID format"), fiber.StatusBadRequest)
	}

	var discount tickets.DiscountCode
	if err := ctx.BodyParser(&discount); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	updatedDiscount, err := c.service.UpdateDiscount(userID, id, &discount)
	if err != nil {
		switch err.Error() {
		case "user lacks update:discounts permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "discount not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, updatedDiscount, "Discount updated successfully", true)
}

func (c *DiscountController) CancelDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid discount ID format"), fiber.StatusBadRequest)
	}

	if err := c.service.CancelDiscount(userID, id); err != nil {
		switch err.Error() {
		case "user lacks update:discounts permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "discount not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Discount cancelled successfully", true)
}

func (c *DiscountController) ValidateDiscount(ctx *fiber.Ctx) error {
	code := ctx.Query("code")
	if code == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Discount code is required"), fiber.StatusBadRequest)
	}

	eventID := ctx.Query("event_id")
	orderValue := ctx.QueryFloat("order_value", 0)

	discount, err := c.service.ValidateDiscountCode(code, eventID, orderValue)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, discount, "Discount is valid", true)
}
