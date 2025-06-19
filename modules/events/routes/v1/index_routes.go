package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupEventsMainRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService, emailService mail_service.EmailService) {
	CategoryRoutes(router, db, logHandler, cloudinary)
	SetupEventsRoutes(router, db, logHandler, cloudinary)
	VenueRoutes(router, db, logHandler, cloudinary)
}
