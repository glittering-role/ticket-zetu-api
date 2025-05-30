package tickets_controller

import (
	"ticket-zetu-api/logs/handler"

	price_tier_service "ticket-zetu-api/modules/tickets/price_tires/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PriceTierController struct {
	service    price_tier_service.PriceTierService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

func NewPriceTierController(service price_tier_service.PriceTierService, logHandler *handler.LogHandler) *PriceTierController {
	return &PriceTierController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

func (c *PriceTierController) GetPriceTiersForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	fields := ctx.Query("fields", "id,name,percentage_increase,status,is_default")

	priceTiers, err := c.service.GetPriceTiers(userID, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, priceTiers, "Price tiers retrieved successfully", true)
}

func (c *PriceTierController) GetSinglePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	fields := ctx.Query("fields", "id,name,description,percentage_increase,status,is_default,effective_from,effective_to,min_tickets,max_tickets")

	priceTier, err := c.service.GetPriceTier(userID, id, fields)
	if err != nil {
		switch err.Error() {
		case "user lacks read:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier retrieved successfully", true)
}

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
		case "cannot delete default price tier":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "price tier is in use by events":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Price tier deleted successfully", false)
}
