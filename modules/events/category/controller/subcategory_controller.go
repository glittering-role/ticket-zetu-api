package category

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/category/dto"
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

// CreateSubcategory godoc
// @Summary Create a new subcategory
// @Description Creates a new subcategory under a specific category
// @Tags Subcategories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param category_id body string true "Category ID"
// @Param name body string true "Subcategory name"
// @Param description body string false "Subcategory description"
// @Success 201 {object} map[string]interface{} "Subcategory created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 409 {object} map[string]interface{} "Subcategory name already exists in this category"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subcategories [post]
func (c *SubcategoryController) CreateSubcategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	var input dto.CreateSubcategoryDto

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

// GetSubcategories godoc
// @Summary Get subcategories by category ID
// @Description Retrieves all subcategories under a specific category
// @Tags Subcategories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]interface{} "Subcategories retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category ID format"
// @Failure 403 {object} map[string]interface{} "User lacks view permission"
// @Failure 404 {object} map[string]interface{} "Subcategories not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /categories/{id}/subcategories [get]
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
