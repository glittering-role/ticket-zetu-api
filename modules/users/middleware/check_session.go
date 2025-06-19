package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"ticket-zetu-api/database"
	"ticket-zetu-api/logs/handler"
)

// UserSessionData holds cached user data
type UserSessionData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// IsAuthenticated ensures the session_token cookie is valid
func IsAuthenticated(db *gorm.DB, logHandler *handler.LogHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionToken := c.Cookies("session_token")
		if sessionToken == "" {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session token missing"), fiber.StatusUnauthorized)
		}

		// Check Redis cache
		redisClient := database.GetRedisClient()
		ctx := context.Background()
		cacheKey := "session:" + sessionToken
		cachedData, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var user UserSessionData
			if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
				//c.Locals("user", user)
				c.Locals("user_id", user.ID) // Maintain compatibility with existing routes
				c.Locals("username", user.Username)
				c.Locals("email", user.Email)
				return c.Next()
			}
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
		} else if err != redis.Nil {
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
		}

		// Database query
		var result struct {
			UserID    string
			Username  string
			Email     string
			ExpiresAt time.Time
		}
		err = db.Table("user_sessions s").
			Joins("JOIN users u ON s.user_id = u.id").
			Select("u.id AS user_id, u.username, u.email, s.expires_at").
			Where("s.session_token = ? AND s.expires_at > ?", sessionToken, time.Now()).
			Scan(&result).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: invalid session"), fiber.StatusUnauthorized)
			}
			return logHandler.LogError(c, err, fiber.StatusInternalServerError)
		}

		// Store in Redis
		user := UserSessionData{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		}
		userData, err := json.Marshal(user)
		if err != nil {
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
		} else {
			ttl := time.Until(result.ExpiresAt)
			if err := redisClient.Set(ctx, cacheKey, userData, ttl).Err(); err != nil {
				logHandler.LogError(c, err, fiber.StatusInternalServerError)
			}
		}

		//c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("email", user.Email)
		return c.Next()
	}
}
