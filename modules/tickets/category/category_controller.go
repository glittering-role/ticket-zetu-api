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

// CreateCategory handles the creation of a new category
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

	// Basic validation
	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateCategory(userID, input.Name, input.Description, input.ImageURL)
	if err != nil {
		if err.Error() == "user lacks create:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "category name already exists" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category created successfully", true)
}

// UpdateCategory handles updating an existing category
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

	// Basic validation
	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateCategory(userID, id, input.Name, input.Description, input.ImageURL, input.IsActive)
	if err != nil {
		if err.Error() == "invalid category ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "category not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks update:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "category name already exists" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category updated successfully", true)
}

// DeleteCategory handles soft deletion of a category
func (c *CategoryController) DeleteCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteCategory(userID, id)
	if err != nil {
		if err.Error() == "invalid category ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "category not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks delete:categories permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category deleted successfully", true)
}

// Existing methods remain unchanged
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
