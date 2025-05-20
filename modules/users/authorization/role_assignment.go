package authorization

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignRoleToUser assigns a role to a user
func (c *RoleController) AssignRoleToUser(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check permission
	hasPerm, err := c.service.HasPermission(userID, "assign:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks assign:role permission"), fiber.StatusForbidden)
	}

	// Parse input
	var input struct {
		UserID string `json:"user_id"`
		RoleID string `json:"role_id"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Validate UUIDs
	if _, err := uuid.Parse(input.UserID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID format"), fiber.StatusBadRequest)
	}
	if _, err := uuid.Parse(input.RoleID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid role ID format"), fiber.StatusBadRequest)
	}

	// Assign role
	if err := c.service.AssignRoleToUser(input.UserID, input.RoleID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "user not found" || err.Error() == "role not found or inactive" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Role assigned to user successfully", true)
}
