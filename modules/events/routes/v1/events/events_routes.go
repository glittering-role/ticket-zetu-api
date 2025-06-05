package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	events_controller "ticket-zetu-api/modules/events/events/controller"
	service "ticket-zetu-api/modules/events/events/service"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupEventsRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)
	// Event routes
	eventService := service.NewEventService(db, authService, cloudinary)
	eventController := events_controller.NewEventController(eventService, logHandler, cloudinary)

	eventGroup := router.Group("/events", authMiddleware)
	{
		eventGroup.Get("/", eventController.GetEventsForOrganizer)
		eventGroup.Get("/:id", eventController.GetSingleEventForOrganizer)
		eventGroup.Post("/", eventController.CreateEvent)
		eventGroup.Put("/:id", eventController.UpdateEvent)
		eventGroup.Delete("/:id", eventController.DeleteEvent)
		eventGroup.Post("/:id/images", eventController.AddEventImage)
		eventGroup.Delete("/:event_id/images/:image_id", eventController.DeleteEventImage)
	}
}
