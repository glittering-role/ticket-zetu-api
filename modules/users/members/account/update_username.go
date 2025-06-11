package account

import (
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateUsername godoc
// @Summary Update user username
// @Description Updates the authenticated user's username
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdateUsernameDto true "Username"
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 409 {object} object
// @Failure 500 {object} object
// @Router /users/me/username [post]
func (c *UserController) UpdateUsername(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var usernameDto dto.UpdateUsernameDto
	if err := ctx.BodyParser(&usernameDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateUsername(userID, &usernameDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.HasPrefix(err.Error(), "invalid username") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "username already taken") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Username updated successfully", true)
}
