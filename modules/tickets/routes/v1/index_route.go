package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupTicketRoutes consolidates all ticket-related routes
func SetupTicketMainRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService, emailService mail_service.EmailService) {
	// Call individual route setup functions
	SetupTicketTypeRoutes(router, db, logHandler)
	SetupPriceTierRoutes(router, db, logHandler)
	SetupDiscountRoutes(router, db, logHandler)
}
