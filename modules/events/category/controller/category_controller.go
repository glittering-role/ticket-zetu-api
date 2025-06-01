package category

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/category/services"

	"github.com/gofiber/fiber/v2"
)

type CategoryController struct {
	service    services.CategoryService
	logHandler *handler.LogHandler
}

func NewCategoryController(service services.CategoryService, logHandler *handler.LogHandler) *CategoryController {
	return &CategoryController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *CategoryController) CreateCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	var input struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		Description string `json:"description,omitempty"`
		ImageURL    string `json:"image_url,omitempty"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateCategory(userID, input.Name, input.Description, input.ImageURL)
	if err != nil {
		switch err.Error() {
		case "user lacks create:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "category name already exists":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category created successfully", true)
}

func (c *CategoryController) UpdateCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")
	var input struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		Description string `json:"description,omitempty"`
		ImageURL    string `json:"image_url,omitempty"`
		IsActive    bool   `json:"is_active"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateCategory(userID, id, input.Name, input.Description, input.ImageURL, input.IsActive)
	if err != nil {
		switch err.Error() {
		case "invalid category ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks update:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "category name already exists":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category updated successfully", true)
}

func (c *CategoryController) DeleteCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteCategory(userID, id)
	if err != nil {
		switch err.Error() {
		case "invalid category ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks delete:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category deleted successfully", true)
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
		switch err.Error() {
		case "invalid category ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks view:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, category, "Category retrieved successfully", true)
}
