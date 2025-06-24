package authentication

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Logout godoc
// @Summary Terminate user session
// @Description Logs out the user and invalidates the session
// @Tags Authentication
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 400 {object} map[string]interface{} "No session found"
// @Failure 500 {object} map[string]interface{} "Failed to process logout"
// @Router /auth/logout [post]
// Logout handles session termination
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	sessionToken := ctx.Cookies("session_token")
	if sessionToken == "" {
		return c.logHandler.LogError(ctx, errors.New("no session token found"), fiber.StatusBadRequest)
	}

	// Call the service to handle logout logic
	err := c.userService.Logout(context.Background(), ctx, sessionToken)
	if err != nil {
		// Log the detailed error internally but return a generic message to the client
		c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process logout",
		})
	}

	return c.logHandler.LogSuccess(ctx, nil, "Logout successful", true)
}
