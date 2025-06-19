package routes

import (
	"github.com/redis/go-redis/v9"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupUsersMainRoutes(router fiber.Router, db *gorm.DB, redisClient *redis.Client, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService, emailService mail_service.EmailService) {
	ArtistRoutes(router, db, logHandler)
	SetupAuthRoutes(router, db, redisClient, logHandler, emailService)
	AuthorizationRoutes(router, db, logHandler)
	UserRoutes(router, db, logHandler, cloudinary, emailService)
}
