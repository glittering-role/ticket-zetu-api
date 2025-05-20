package category

import (
	"github.com/gofiber/fiber/v2"
	"ticket-zetu-api/logs/handler"
)

type CategoryController struct {
	service    CategoryService
	logHandler *handler.LogHandler
}

func NewCategoryController(service CategoryService, logHandler *handler.LogHandler) *CategoryController {
	return &CategoryController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *CategoryController) GetCategories(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	categories, err := c.service.GetCategories(userID)
	if err != nil {
		if err.Error() == "user lacks view:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, categories, "Categories retrieved successfully", true)
}

func (c *CategoryController) GetCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	category, err := c.service.GetCategory(userID, id)
	if err != nil {
		if err.Error() == "invalid category ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "category not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks view:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, category, "Category retrieved successfully", true)
}

func (c *CategoryController) GetSubcategories(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	categoryID := ctx.Params("id")
	subcategories, err := c.service.GetSubcategories(userID, categoryID)
	if err != nil {
		if err.Error() == "invalid category ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "subcategories not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks view:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, subcategories, "Subcategories retrieved successfully", true)
}
