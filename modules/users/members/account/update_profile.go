package account

import (
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateLocation godoc
// @Summary Update user location
// @Description Updates the authenticated user's location details
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UserLocationUpdateDto true "Location details"
// / @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /users/me/location [post]
func (c *UserController) UpdateLocation(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var locationDto dto.UserLocationUpdateDto
	if err := ctx.BodyParser(&locationDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	profile, err := c.service.UpdateUserLocation(userID, &locationDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.HasPrefix(err.Error(), "invalid") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "User location updated successfully", true)
}
