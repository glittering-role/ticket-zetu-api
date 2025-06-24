package auth_service

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"ticket-zetu-api/modules/users/models/members"
)

func (s *userService) Logout(ctx context.Context, c *fiber.Ctx, sessionToken string) error {
	if sessionToken == "" {
		return s.logHandler.LogError(c, errors.New("no session token found"), fiber.StatusBadRequest)
	}

	// Start a transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return s.logHandler.LogError(c, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First check and delete from Redis
	redisKey := "session:" + sessionToken
	_, err := s.redisClient.Del(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		tx.Rollback()
		return s.logHandler.LogError(c, errors.New("failed to delete redis session"), fiber.StatusInternalServerError)
	}

	// Find and validate the active session
	var session members.UserSession
	if err := tx.Where("session_token = ? AND is_active = ?", sessionToken, true).
		First(&session).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.logHandler.LogError(c, errors.New("no active session found"), fiber.StatusBadRequest)
		}
		return s.logHandler.LogError(c, errors.New("failed to find session"), fiber.StatusInternalServerError)
	}

	// Explicitly update session to mark as terminated with LoggedOutAt
	now := time.Now()
	if err := tx.Model(&session).
		Updates(map[string]interface{}{
			"is_active":     false,
			"logged_out_at": now,
			"updated_at":    now,
		}).Error; err != nil {
		tx.Rollback()
		return s.logHandler.LogError(c, errors.New("failed to update session"), fiber.StatusInternalServerError)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return s.logHandler.LogError(c, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
	}

	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return nil
}
