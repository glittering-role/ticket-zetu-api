package authorization

import (
	models "ticket-zetu-api/modules/users/models/authorization"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AssignPermissionToRole assigns a permission to a role
func (c *PermissionController) AssignPermissionToRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to assign permissions
	// hasPerm, err := c.service.HasPermission(userID, "assign:permission")
	// if err != nil {
	// 	return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	// }
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks assign:permission permission"), fiber.StatusForbidden)
	// }

	var input struct {
		RoleID       string `json:"role_id"`
		PermissionID string `json:"permission_id"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(input.RoleID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid role ID format"), fiber.StatusBadRequest)
	}
	if _, err := uuid.Parse(input.PermissionID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid permission ID format"), fiber.StatusBadRequest)
	}

	// Check if the role level is valid for the user's max role level
	// userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	// if err != nil {
	// 	return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	// }
	var role models.Role
	if err := c.db.Where("id = ? AND status = ? AND deleted_at IS NULL", input.RoleID, models.RoleActive).First(&role).Error; err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found or inactive"), fiber.StatusNotFound)
	}
	// if role.Level > userMaxLevel {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot assign permission to role with higher level"), fiber.StatusForbidden)
	// }

	// Assign permission
	if err := c.service.AssignPermissionToRole(input.RoleID, input.PermissionID, userID); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission assigned to role successfully", true)
}

func (c *PermissionController) RemovePermissionFromRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to remove permissions
	hasPerm, err := c.service.HasPermission(userID, "remove:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks remove:permission permission"), fiber.StatusForbidden)
	}

	var input struct {
		RoleID       string `json:"role_id"`
		PermissionID string `json:"permission_id"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(input.RoleID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid role ID format"), fiber.StatusBadRequest)
	}
	if _, err := uuid.Parse(input.PermissionID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid permission ID format"), fiber.StatusBadRequest)
	}

	// Check if the role level is valid for the user's max role level
	userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	var role models.Role
	if err := c.db.Where("id = ? AND status = ? AND deleted_at IS NULL", input.RoleID, models.RoleActive).First(&role).Error; err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found or inactive"), fiber.StatusNotFound)
	}
	if role.Level > userMaxLevel {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot remove permission from role with higher level"), fiber.StatusForbidden)
	}

	// Remove permission
	if err := c.service.RemovePermissionFromRole(input.RoleID, input.PermissionID, userID); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission removed from role successfully", true)
}
