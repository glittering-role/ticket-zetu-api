package tickets_controller

import (
	"strconv"
	"ticket-zetu-api/logs/handler"

	"github.com/go-playground/validator/v10"
	price_tier_service "ticket-zetu-api/modules/tickets/price_tires/service"

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

// GetPriceTiersForOrganizer godoc
// @Summary Get price tiers for organizer
// @Description Retrieves all price tiers belonging to the authenticated user's organization
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.GetPriceTierResponse "List of price tiers"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers/organization [get]
func (c *PriceTierController) GetPriceTiersForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	priceTiers, err := c.service.GetPriceTiers(userID)
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

// GetSinglePriceTier godoc
// @Summary Get a single price tier
// @Description Retrieves details of a specific price tier
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Price Tier ID"
// @Success 200 {object} map[string]interface{} "Price tier details"
// @Failure 400 {object} map[string]interface{} "Invalid price tier ID"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Price tier not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers/{id} [get]
func (c *PriceTierController) GetSinglePriceTier(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	priceTier, err := c.service.GetPriceTier(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks read:price_tiers permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, priceTier, "Price tier retrieved successfully", true)
}

// GetAllPriceTiers godoc
// @Summary Get all price tiers
// @Description Retrieves all price tiers across all organizations (paginated)
// @Tags Price Tiers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Paginated list of price tiers"
// @Failure 400 {object} map[string]interface{} "Invalid pagination parameters"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /price-tiers [get]
func (c *PriceTierController) GetAllPriceTiers(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	page := ctx.Query("page", "1")
	limit := ctx.Query("limit", "10")

	// Validate page and limit
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter"), fiber.StatusBadRequest)
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter"), fiber.StatusBadRequest)
	}

	priceTiers, err := c.service.GetAllPriceTiers(userID, pageInt, limitInt)
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
