package category

import (
	"ticket-zetu-api/modules/events/category/dto"

	"github.com/gofiber/fiber/v2"
)

// UpdateCategory godoc
// @Summary Update Category
// @Description Update category.
// @Tags Category Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Param input body dto.CreateCategoryDto true "Category details"
// @Success 200 {object} map[string]interface{} "Category updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /categories/{id} [put]
func (c *CategoryController) UpdateCategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input dto.CreateCategoryDto

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateCategory(userID, id, input.Name, input.Description)
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

// DeleteCategory godoc
// @Summary Delete a category
// @Description Deletes a specific category by ID
// @Tags Category Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]interface{} "Category deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category ID format"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /categories/{id} [delete]
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

// ToggleCategoryStatus godoc
// @Summary Toggle category status
// @Description Activates or deactivates a specific category
// @Tags Category Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Category ID"
// @Param input body dto.ToggleCategoryStatusInput true "Status details"
// @Success 200 {object} map[string]interface{} "Category status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or category ID"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /categories/{id}/toggle-status [put]
func (c *CategoryController) ToggleCategoryStatus(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input dto.ToggleCategoryStatusInput

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	err := c.service.ToggleCategoryStatus(userID, id, input.IsActive)
	if err != nil {
		switch err.Error() {
		case "invalid category ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks update:categories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Category status updated successfully", true)
}
