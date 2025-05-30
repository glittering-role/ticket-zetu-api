package tickets_controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (c *PriceTierController) GetAllPriceTiers(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	fields := ctx.Query("fields", "id,name,percentage_increase,status,is_default")
	page := ctx.Query("page", "1")    // Default to page 1
	limit := ctx.Query("limit", "10") // Default to limit 10

	// Validate page and limit
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter"), fiber.StatusBadRequest)
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter"), fiber.StatusBadRequest)
	}

	priceTiers, err := c.service.GetAllPriceTiers(userID, fields, pageInt, limitInt)
	if err != nil {
		switch err.Error() {
		case "user lacks read:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, priceTiers, "All price tiers retrieved successfully", true)
}
