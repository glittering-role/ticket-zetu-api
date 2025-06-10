package account

import (
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdatePhone godoc
// @Summary Update user phone number
// @Description Updates the authenticated user's phone number
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdatePhoneDto true "Phone number"
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 409 {object} object
// @Failure 500 {object} object
// @Router /users/me/phone [post]
func (c *UserController) UpdatePhone(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var phoneDto dto.UpdatePhoneDto
	if err := ctx.BodyParser(&phoneDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdatePhone(userID, &phoneDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.HasPrefix(err.Error(), "invalid") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "phone already exists") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Phone number updated successfully", true)
}
