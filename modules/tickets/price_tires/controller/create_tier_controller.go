package tickets_controller

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func (c *PriceTierController) CreatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input struct {
		Name               string     `json:"name" validate:"required,min=2,max=50"`
		Description        string     `json:"description"`
		PercentageIncrease float64    `json:"percentage_increase" validate:"required,min=0"`
		Status             string     `json:"status" validate:"omitempty,oneof=active inactive archived"`
		IsDefault          bool       `json:"is_default"`
		EffectiveFrom      time.Time  `json:"effective_from"`
		EffectiveTo        *time.Time `json:"effective_to"`
		MinTickets         int        `json:"min_tickets" validate:"min=0"`
		MaxTickets         *int       `json:"max_tickets" validate:"omitempty,min=0"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	priceTier, err := c.service.CreatePriceTier(
		userID,
		input.Name,
		input.Description,
		input.PercentageIncrease,
		input.Status,
		input.IsDefault,
		input.EffectiveFrom,
		input.EffectiveTo,
		input.MinTickets,
		input.MaxTickets,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks create:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier created successfully", false)
}

func (c *PriceTierController) UpdatePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input struct {
		Name               string     `json:"name" validate:"omitempty,min=2,max=50"`
		Description        string     `json:"description"`
		PercentageIncrease float64    `json:"percentage_increase" validate:"omitempty,min=0"`
		Status             string     `json:"status" validate:"omitempty,oneof=active inactive archived"`
		IsDefault          *bool      `json:"is_default"`
		EffectiveFrom      time.Time  `json:"effective_from"`
		EffectiveTo        *time.Time `json:"effective_to"`
		MinTickets         int        `json:"min_tickets" validate:"omitempty,min=0"`
		MaxTickets         *int       `json:"max_tickets" validate:"omitempty,min=0"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	priceTier, err := c.service.UpdatePriceTier(
		userID,
		id,
		input.Name,
		input.Description,
		input.PercentageIncrease,
		input.Status,
		input.IsDefault,
		input.EffectiveFrom,
		input.EffectiveTo,
		input.MinTickets,
		input.MaxTickets,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks update:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier updated successfully", false)
}
