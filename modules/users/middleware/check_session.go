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
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	DeviceType string    `json:"device_type"`
	DeviceName string    `json:"device_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IsAuthenticated ensures the session_token cookie is valid
func IsAuthenticated(db *gorm.DB, logHandler *handler.LogHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionToken := c.Cookies("session_token")
		if sessionToken == "" {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session token missing"), fiber.StatusUnauthorized)
		}

		// Get request metadata
		clientIP := c.IP()

		// Check Redis cache
		redisClient := database.GetRedisClient()
		ctx := context.Background()
		cacheKey := "session:" + sessionToken
		cachedData, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var user UserSessionData
			if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
				// Verify session in database for critical security checks
				var session struct {
					IsActive      bool
					ExpiresAt     time.Time
					LoggedOutAt   *time.Time
					RefreshToken  string
					RefreshExpiry time.Time
					IPAddress     string
					UserAgent     string
					DeviceType    string
					DeviceName    string
					UpdatedAt     time.Time
				}
				err = db.Table("user_sessions").
					Select("is_active, expires_at, logged_out_at, refresh_token, refresh_expiry, ip_address, user_agent, device_type, device_name, updated_at").
					Where("session_token = ?", sessionToken).
					Scan(&session).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						redisClient.Del(ctx, cacheKey)
						return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: invalid session"), fiber.StatusUnauthorized)
					}
					logHandler.LogError(c, err, fiber.StatusInternalServerError)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to validate session",
					})
				}

				// Security checks
				if !session.IsActive {
					redisClient.Del(ctx, cacheKey)
					return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session is inactive"), fiber.StatusUnauthorized)
				}
				if session.LoggedOutAt != nil && !session.LoggedOutAt.Before(time.Now()) {
					redisClient.Del(ctx, cacheKey)
					return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session has been terminated"), fiber.StatusUnauthorized)
				}
				if session.ExpiresAt.Before(time.Now()) {
					redisClient.Del(ctx, cacheKey)
					return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session expired"), fiber.StatusUnauthorized)
				}
				if session.RefreshToken != "" && session.RefreshExpiry.Before(time.Now()) {
					redisClient.Del(ctx, cacheKey)
					return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: refresh token expired"), fiber.StatusUnauthorized)
				}

				if time.Since(session.UpdatedAt) > 24*time.Hour {
					redisClient.Del(ctx, cacheKey)
					return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session inactive for too long"), fiber.StatusUnauthorized)
				}

				// Session is valid, set locals and proceed
				c.Locals("user_id", user.ID)
				c.Locals("username", user.Username)
				c.Locals("email", user.Email)
				return c.Next()
			}
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process session data",
			})
		} else if err != redis.Nil {
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate session",
			})
		}

		// Database query if not in Redis
		var result struct {
			UserID        string
			Username      string
			Email         string
			ExpiresAt     time.Time
			IsActive      bool
			LoggedOutAt   *time.Time
			RefreshToken  string
			RefreshExpiry time.Time
			IPAddress     string
			UserAgent     string
			DeviceType    string
			DeviceName    string
			UpdatedAt     time.Time
		}
		err = db.Table("user_sessions s").
			Joins("JOIN user_profiles u ON s.user_id = u.id").
			Select("u.id AS user_id, u.username, u.email, s.expires_at, s.is_active, s.logged_out_at, s.refresh_token, s.refresh_expiry, s.ip_address, s.user_agent, s.device_type, s.device_name, s.updated_at").
			Where("s.session_token = ? AND s.is_active = ? AND s.expires_at > ?", sessionToken, true, time.Now()).
			Scan(&result).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: invalid session"), fiber.StatusUnauthorized)
			}
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate session",
			})
		}

		// Additional security checks
		if result.LoggedOutAt != nil && !result.LoggedOutAt.Before(time.Now()) {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session has been terminated"), fiber.StatusUnauthorized)
		}
		if result.RefreshToken != "" && result.RefreshExpiry.Before(time.Now()) {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: refresh token expired"), fiber.StatusUnauthorized)
		}
		if result.IPAddress != "" && result.IPAddress != clientIP {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: IP address mismatch"), fiber.StatusUnauthorized)
		}

		if time.Since(result.UpdatedAt) > 24*time.Hour {
			return logHandler.LogError(c, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: session inactive for too long"), fiber.StatusUnauthorized)
		}

		// Store in Redis
		user := UserSessionData{
			ID:         result.UserID,
			Username:   result.Username,
			Email:      result.Email,
			IPAddress:  result.IPAddress,
			UserAgent:  result.UserAgent,
			DeviceType: result.DeviceType,
			DeviceName: result.DeviceName,
			UpdatedAt:  result.UpdatedAt,
		}

		userData, err := json.Marshal(user)
		if err != nil {
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process session data",
			})
		}
		ttl := time.Until(result.ExpiresAt)
		if err := redisClient.Set(ctx, cacheKey, userData, ttl).Err(); err != nil {
			logHandler.LogError(c, err, fiber.StatusInternalServerError)
		}

		// Set locals and proceed
		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("email", user.Email)
		return c.Next()
	}
}
