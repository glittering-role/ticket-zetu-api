package authentication

import (
	"context"
	"errors"
	"ticket-zetu-api/modules/users/models/members"
	"time"

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
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/logout [post]
// Logout handles session termination
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	sessionToken := ctx.Cookies("session_token")
	if sessionToken == "" {
		return c.logHandler.LogError(ctx, errors.New("no session token found"), fiber.StatusBadRequest)
	}

	// Start a transaction
	tx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete session from database
	err := tx.
		Where("session_token = ?", sessionToken).
		Delete(&members.UserSession{}).Error
	if err != nil {
		tx.Rollback()
		return c.logHandler.LogError(ctx, errors.New("failed to invalidate session"), fiber.StatusInternalServerError)
	}

	// Delete session from Redis
	redisKey := "session:" + sessionToken
	_, err = c.redisClient.Del(context.Background(), redisKey).Result()
	if err != nil {
		tx.Rollback()
		return c.logHandler.LogError(ctx, errors.New("failed to invalidate redis session"), fiber.StatusInternalServerError)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
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
