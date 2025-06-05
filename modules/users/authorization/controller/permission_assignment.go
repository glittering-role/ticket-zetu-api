package authorization

import (
	"ticket-zetu-api/modules/users/authorization/dto"

	"github.com/gofiber/fiber/v2"
)

// SuccessResponse defines a standard success response structure
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AssignPermissionToRole godoc
// @Summary Assign a permission to a role
// @Description Assigns a permission to a role if the user has sufficient permissions
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.PermissionAssignmentDto true "Role and Permission IDs"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions/assign [post]
func (c *PermissionController) AssignPermissionToRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	hasPerm, err := c.service.HasPermission(userID, "assign:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks assign:permission permission"), fiber.StatusForbidden)
	}

	var input dto.PermissionAssignmentDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	err = c.service.AssignPermissionToRole(input.RoleID, input.PermissionID, userID)
	if err != nil {
		if err.Error() == "role not found or inactive" || err.Error() == "permission not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "permission already assigned to role" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission assigned to role successfully", true)
}

// RemovePermissionFromRole godoc
// @Summary Remove a permission from a role
// @Description Removes a permission from a role if the user has sufficient permissions
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.PermissionAssignmentDto true "Role and Permission IDs"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions/remove [delete]
func (c *PermissionController) RemovePermissionFromRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	hasPerm, err := c.service.HasPermission(userID, "remove:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks remove:permission permission"), fiber.StatusForbidden)
	}

	var input dto.PermissionAssignmentDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	err = c.service.RemovePermissionFromRole(input.RoleID, input.PermissionID, userID)
	if err != nil {
		if err.Error() == "role not found or inactive" || err.Error() == "permission not found" || err.Error() == "role-permission assignment not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "cannot remove permission from role with higher level" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission removed from role successfully", true)
}
