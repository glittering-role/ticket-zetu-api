package category

import (
	"ticket-zetu-api/modules/events/category/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateSubcategory godoc
// @Summary Update an existing subcategory
// @Description Creates a new subcategory under a specific category
// @Tags Subcategories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Subcategory ID"
// @Param input body dto.UpdateSubSubcategoryDto true "Category details"
// @Success 201 {object} map[string]interface{} "Subcategory updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "SubCategory not found"
// @Failure 409 {object} map[string]interface{} "Subcategory name already exists in this category"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subcategories/{id} [put]
func (c *SubcategoryController) UpdateSubcategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input dto.CreateCategoryDto

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if len(input.Name) < 2 || len(input.Name) > 50 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Name must be between 2 and 50 characters"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid subcategory ID format"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateSubcategory(userID, id, input.Name, input.Description)
	if err != nil {
		switch err.Error() {
		case "invalid subcategory ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "parent category not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks update:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "subcategory name already exists in this category":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Subcategory updated successfully", true)
}

// DeleteSubcategory godoc
// @Summary Delete a subcategory
// @Description Deletes a specific subcategory by ID
// @Tags Subcategories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Subcategory ID"
// @Success 200 {object} map[string]interface{} "Subcategory deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid subcategory ID format"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Subcategory not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subcategories/{id} [delete]
func (c *SubcategoryController) DeleteSubcategory(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid subcategory ID format"), fiber.StatusBadRequest)
	}

	err := c.service.DeleteSubcategory(userID, id)
	if err != nil {
		switch err.Error() {
		case "invalid subcategory ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks delete:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Subcategory deleted successfully", true)
}

// ToggleSubcategoryStatus godoc
// @Summary Toggle subcategory status
// @Description Activates or deactivates a specific subcategory
// @Tags Subcategories
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Subcategory ID"
// @Param input body dto.ToggleCategoryStatus true "Toggle subcategory status"
// @Success 200 {object} map[string]interface{} "Subcategory status toggled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid subcategory ID format"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Subcategory not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subcategories/{id}/toggle-status [put]
func (c *SubcategoryController) ToggleSubcategoryStatus(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid subcategory ID format"), fiber.StatusBadRequest)
	}

	var input dto.ToggleCategoryStatus
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	err := c.service.ToggleSubCategoryStatus(userID, id, input.IsActive)
	if err != nil {
		switch err.Error() {
		case "invalid subcategory ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks update:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Subcategory status toggled successfully", true)
}
