package category

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/category/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SubcategoryController struct {
	service    services.SubcategoryService
	logHandler *handler.LogHandler
}

func NewSubcategoryController(service services.SubcategoryService, logHandler *handler.LogHandler) *SubcategoryController {
	return &SubcategoryController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *SubcategoryController) CreateSubcategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	var input struct {
		CategoryID  string `json:"category_id" validate:"required"`
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

	if _, err := uuid.Parse(input.CategoryID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid category ID format"), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateSubcategory(userID, input.CategoryID, input.Name, input.Description)
	if err != nil {
		switch err.Error() {
		case "user lacks create:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "subcategory name already exists in this category":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Subcategory created successfully", true)
}

func (c *SubcategoryController) GetSubcategories(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	categoryID := ctx.Params("id")
	subcategories, err := c.service.GetSubcategories(userID, categoryID)
	if err != nil {
		switch err.Error() {
		case "invalid category ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subcategories not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks view:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, subcategories, "Subcategories retrieved successfully", true)
}
