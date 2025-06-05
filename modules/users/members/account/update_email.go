package account

import (
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateEmail godoc
// @Summary Update user email
// @Description Updates the authenticated user's email, requiring verification
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdateEmailDto true "Email address"
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 409 {object} object
// @Failure 500 {object} object
// @Router /users/me/email [post]
func (c *UserController) UpdateEmail(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var emailDto dto.UpdateEmailDto
	if err := ctx.BodyParser(&emailDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	profile, err := c.service.UpdateUserEmail(userID, &emailDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.HasPrefix(err.Error(), "invalid") || strings.Contains(err.Error(), "failed to send") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "email already exists") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "Email updated successfully, verification required", true)
}
