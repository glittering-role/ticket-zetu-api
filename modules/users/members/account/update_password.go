package account

import (
	"errors"
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SetNewPassword godoc
// @Summary Update user password
// @Description Updates the authenticated user's password with a new password and confirmation
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.NewPasswordDto true "New password and confirmation"
// @Success 200 {object} map[string]interface{} "Password updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, password requirements not met, or same password used"
// @Failure 401 {object} map[string]interface{} "Unauthorized access"
// @Failure 404 {object} map[string]interface{} "User or security attributes not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me/password [post]
func (c *UserController) SetNewPassword(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, errors.New("user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, errors.New("invalid user ID in context"), fiber.StatusBadRequest)
	}

	var passwordDto dto.NewPasswordDto
	if err := ctx.BodyParser(&passwordDto); err != nil {
		return c.logHandler.LogError(ctx, errors.New("invalid request body"), fiber.StatusBadRequest)
	}

	_, err := c.service.SetNewPassword(userID, &passwordDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.Contains(err.Error(), "password must contain") || strings.Contains(err.Error(), "new password must be different") {
			return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" || err.Error() == "user security attributes not found" {
			return c.logHandler.LogError(ctx, err, fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Password updated successfully", true)
}
