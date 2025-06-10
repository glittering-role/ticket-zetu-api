package preference

import (
	"strings"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/members/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserPreferencesController handles user preferences HTTP requests
type UserPreferencesController struct {
	service    service.UserPreferencesService
	logHandler *handler.LogHandler
}

// NewUserPreferencesController initializes the controller
func NewUserPreferencesController(service service.UserPreferencesService, logHandler *handler.LogHandler) *UserPreferencesController {
	return &UserPreferencesController{
		service:    service,
		logHandler: logHandler,
	}
}

// GetUserPreferences godoc
// @Summary Get user preferences
// @Description Retrieves the authenticated user's preferences
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserPreferencesDto
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /users/me/preferences [get]
func (c *UserPreferencesController) GetUserPreferences(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	preferences, err := c.service.GetUserPreferences(userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, preferences, "User preferences retrieved successfully", true)
}

// UpdateUserPreferences godoc
// @Summary Update user preferences
// @Description Updates the authenticated user's preferences
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdateUserPreferencesDto true "Preferences to update"
// @Success 200 {object} dto.UserPreferencesDto
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /users/me/preferences [post]
func (c *UserPreferencesController) UpdateUserPreferences(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var preferencesDto dto.UpdateUserPreferencesDto
	if err := ctx.BodyParser(&preferencesDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateUserPreferences(userID, &preferencesDto)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid") || strings.HasPrefix(err.Error(), "validation") || strings.HasPrefix(err.Error(), "no updates") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "User preferences updated successfully", true)
}
