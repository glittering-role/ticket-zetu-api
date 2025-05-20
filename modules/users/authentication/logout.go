package authentication

import (
	"errors"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logout handles session termination
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	sessionToken := ctx.Cookies("session_token")
	if sessionToken == "" {
		return c.logHandler.LogError(ctx, errors.New("no session token found"), fiber.StatusBadRequest)
	}

	err := c.db.
		Where("session_token = ?", sessionToken).
		Delete(&members.UserSession{}).Error
	if err != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to invalidate session"), fiber.StatusInternalServerError)
	}

	// Clear cookies
	ctx.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.logHandler.LogSuccess(ctx, nil, "Logout successful", true)
}
