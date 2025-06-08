package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/notifications/service"
	authorization_service "ticket-zetu-api/modules/users/authorization/service"

	"ticket-zetu-api/modules/notifications/controllers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupNotificationRoutes(api fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authService := authorization_service.NewPermissionService(db)
	notificationService := notification_service.NewNotificationService(db, authService)
	notificationController := notification_controllers.NewNotificationController(notificationService, logHandler)

	// Group routes under /users
	users := api.Group("/users")
	{
		users.Get("/:user_id/notifications", notificationController.GetUserNotifications)
		users.Delete("/:user_id/notifications", notificationController.DeleteUserNotifications)
		users.Patch("/:user_id/notifications/read", notificationController.MarkNotificationsAsRead)
	}
}
