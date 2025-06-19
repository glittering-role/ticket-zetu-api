package discount_controller

import (
	"ticket-zetu-api/modules/tickets/discount/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetDiscount godoc
// @Summary Get a single discount
// @Description Get details of a specific discount
// @Tags Discounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Discount ID"
// @Success 200 {object} map[string]interface{} "Discount retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid discount ID or organizer not found"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Discount not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /discounts/{id} [get]
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

// GetDiscounts godoc
// @Summary Get all discounts for organizer
// @Description Get all discounts belonging to the organizer
// @Tags Discounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} map[string]interface{} "Discounts retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Organizer not found"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /discounts [get]
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

// UpdateDiscount godoc
// @Summary Update a discount
// @Description Update an existing discount code
// @Tags Discounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Discount ID"
// @Param input body tickets.DiscountCode true "Discount update data"
// @Success 200 {object} map[string]interface{} "Discount updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or organizer not found"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Discount not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /discounts/{id} [put]
func (c *DiscountController) UpdateDiscount(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid discount ID format"), fiber.StatusBadRequest)
	}

	var discount dto.UpdateDiscountCodeInput
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

// CancelDiscount godoc
// @Summary Cancel a discount
// @Description Mark a discount as inactive/cancelled
// @Tags Discounts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Discount ID"
// @Success 200 {object} map[string]interface{} "Discount cancelled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid discount ID or organizer not found"
// @Failure 403 {object} map[string]interface{} "User lacks permission"
// @Failure 404 {object} map[string]interface{} "Discount not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /discounts/{id}/cancel [patch]
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

// ValidateDiscount godoc
// @Summary Validate a discount code
// @Description Check if a discount code is valid for use
// @Tags Discounts
// @Accept json
// @Produce json
// @Param code query string true "Discount code"
// @Param event_id query string false "Event ID"
// @Param order_value query number false "Order value"
// @Success 200 {object} map[string]interface{} "Discount is valid"
// @Failure 400 {object} map[string]interface{} "Invalid discount code or validation failed"
// @Router /discounts/validate [get]

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
