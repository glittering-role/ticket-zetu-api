package authorization

import (
	"ticket-zetu-api/modules/users/authorization/dto"

	"github.com/gofiber/fiber/v2"
)

// AssignRoleToUser godoc
// @Summary Assign a role to a user
// @Description Assigns a role to a user if the user has sufficient permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.RoleAssignmentDto true "User and Role IDs"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/assign [post]
func (c *RoleController) AssignRoleToUser(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	hasPerm, err := c.service.HasPermission(userID, "assign:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks assign:role permission"), fiber.StatusForbidden)
	}

	var input dto.RoleAssignmentDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	err = c.service.AssignRoleToUser(input.UserID, input.RoleID, userID)
	if err != nil {
		if err.Error() == "role not found or inactive" || err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "cannot assign role with higher level than caller's" || err.Error() == "cannot assign system role directly" || err.Error() == "cannot modify user with system role" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return ctx.JSON(SuccessResponse{
		Message: "Role assigned to user successfully",
		Data:    nil,
	})
}
