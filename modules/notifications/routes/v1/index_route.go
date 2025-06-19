package routes

import (
	"ticket-zetu-api/logs/handler"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupNotificationMainRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, emailService mail_service.EmailService) {
	SetupNotificationRoutes(router, db, logHandler)
}
