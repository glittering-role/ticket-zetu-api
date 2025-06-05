package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authentication/service"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Constants for configuration
const (
	Argon2Time          = 1
	Argon2Memory        = 64 * 1024
	Argon2Threads       = 4
	Argon2KeyLength     = 32
	TokenLength         = 32
	SessionDuration     = 24 * time.Hour
	RefreshDuration     = 7 * 24 * time.Hour
	LongSessionDuration = 7 * 24 * time.Hour
)

// AuthController holds dependencies
type AuthController struct {
	db          *gorm.DB
	logHandler  *handler.LogHandler
	userService auth_service.UserService
	validator   *validator.Validate
}

// SetupAuthRoutes sets up authentication routes
// @title Authentication API
// @version 1.0
// @description Handles user authentication including signup, login, logout, and username validation
// @BasePath /api/v1
func NewAuthController(
	db *gorm.DB,
	logHandler *handler.LogHandler,
	userService auth_service.UserService,
) *AuthController {
	return &AuthController{
		db:          db,
		logHandler:  logHandler,
		userService: userService,
		validator:   validator.New(),
	}
}

// generateSecureToken creates a cryptographically secure token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// detectDeviceType determines the device type from the user agent
func detectDeviceType(userAgent string) members.SessionDeviceType {
	userAgent = strings.ToLower(userAgent)
	switch {
	case strings.Contains(userAgent, "mobile"):
		return members.DeviceMobile
	case strings.Contains(userAgent, "tablet"):
		return members.DeviceTablet
	case strings.Contains(userAgent, "windows") || strings.Contains(userAgent, "macintosh"):
		return members.DeviceDesktop
	default:
		return members.DeviceUnknown
	}
}

// isUserOver16 checks if the user is at least 16 years old
func isUserOver16(dob time.Time) bool {
	return time.Since(dob).Hours()/24/365 >= 16
}
