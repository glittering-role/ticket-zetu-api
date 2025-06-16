package account

import (
	"strings"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/dto"
	members_service "ticket-zetu-api/modules/users/members/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserController handles user-related HTTP requests
type UserController struct {
	service    members_service.UserService
	logHandler *handler.LogHandler
}

// NewUserController initializes the controller
func NewUserController(service members_service.UserService, logHandler *handler.LogHandler) *UserController {
	return &UserController{
		service:    service,
		logHandler: logHandler,
	}
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Retrieves a user's profile by ID, username, or email, respecting privacy settings
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param identifier path string true "User ID, username, or email" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /users/{identifier} [get]
func (c *UserController) GetUserProfile(ctx *fiber.Ctx) error {
	identifier := ctx.Params("identifier")
	requesterID := ctx.Locals("user_id").(string)

	if identifier == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "identifier cannot be empty"), fiber.StatusBadRequest)
	}

	profile, err := c.service.GetUserProfile(identifier, requesterID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if strings.HasPrefix(err.Error(), "invalid identifier format") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "User profile retrieved successfully", true)
}

// GetMyProfile godoc
// @Summary Get own profile
// @Description Retrieves the authenticated user's profile
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /users/me [get]
func (c *UserController) GetMyProfile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	profile, err := c.service.GetUserProfile(userID, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if strings.HasPrefix(err.Error(), "invalid identifier format") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "Own profile retrieved successfully", true)
}

// UpdateDetails godoc
// @Summary Update user details
// @Description Updates the authenticated user's basic profile details
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdateUserDto true "User details"
// @Success 200 {object} dto.UserProfileResponseDto
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /users/me/details [patch]
func (c *UserController) UpdateDetails(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	var userDto dto.UpdateUserDto
	if err := ctx.BodyParser(&userDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	_, err := c.service.UpdateUserDetails(userID, &userDto, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "validation failed") || strings.HasPrefix(err.Error(), "invalid") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "User details updated successfully", true)
}
