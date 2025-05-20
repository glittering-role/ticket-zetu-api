package account

import (
	"log"
	"ticket-zetu-api/logs/handler"
	members_service "ticket-zetu-api/modules/users/members/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserController struct {
	service    members_service.UserService
	logHandler *handler.LogHandler
}

func NewUserController(service members_service.UserService, logHandler *handler.LogHandler) *UserController {
	return &UserController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *UserController) GetUserProfile(ctx *fiber.Ctx) error {
	identifier := ctx.Params("id")
	requesterID := ctx.Locals("user_id").(string)

	// Basic validation: ensure identifier is not empty
	if identifier == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "identifier cannot be empty"), fiber.StatusBadRequest)
	}

	profile, err := c.service.GetUserProfile(identifier, requesterID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user profile is private or preferences not set" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "User profile retrieved successfully", true)
}

func (c *UserController) GetMyProfile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "user ID not found in context"), fiber.StatusBadRequest)
	}

	// Validate UUID
	if _, err := uuid.Parse(userID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid user ID in context"), fiber.StatusBadRequest)
	}

	log.Printf("GetMyProfile: userID=%s", userID)
	profile, err := c.service.GetUserProfile(userID, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user profile is private or preferences not set" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, profile, "Own profile retrieved successfully", true)
}
