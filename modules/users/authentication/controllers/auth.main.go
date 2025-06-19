package authentication

import (
	"time"

	"github.com/redis/go-redis/v9"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authentication/mail"
	auth_service "ticket-zetu-api/modules/users/authentication/service"

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
	db           *gorm.DB
	redisClient  *redis.Client
	logHandler   *handler.LogHandler
	userService  auth_service.UserService
	emailService mail_service.EmailService
}

func NewAuthController(
	db *gorm.DB,
	emailService mail_service.EmailService,
	userService auth_service.UserService,
	logHandler *handler.LogHandler,
) (*AuthController, error) {

	return &AuthController{
		db:           db,
		emailService: emailService,
		userService:  userService,
		logHandler:   logHandler,
	}, nil
}

// isUserOver16 checks if the user is at least 16 years old
func isUserOver16(dob time.Time) bool {
	return time.Since(dob).Hours()/24/365 >= 16
}
