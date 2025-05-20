package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"ticket-zetu-api/modules/users/models/members"
)

// IsAuthenticated ensures the session_token cookie is valid
func IsAuthenticated(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionToken := c.Cookies("session_token")
		if sessionToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: session token missing"})
		}

		var session members.UserSession
		err := db.Where("session_token = ? AND expires_at > ?", sessionToken, time.Now()).First(&session).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: invalid session"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server error validating session"})
		}

		// Load minimal user info (ID, Username, Email)
		var user struct {
			ID       string
			Username string
			Email    string
		}
		err = db.Model(&members.User{}).
			Select("id, username, email").
			Where("id = ?", session.UserID).
			Scan(&user).Error
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: user not found"})
		}

		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("email", user.Email)

		return c.Next()
	}
}
