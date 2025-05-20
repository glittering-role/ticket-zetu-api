package routes

import (
	"ticket-zetu-api/logs/controllers"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/logs/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(api fiber.Router, logService *service.LogService, logHandler *handler.LogHandler) {
	// Initialize log controller
	logController := controller.NewLogController(logService)

	logs := api.Group("/logs")

	// Log management routes
	logs.Get("/", func(c *fiber.Ctx) error {
		return logController.GetLogs(c, logHandler)
	})
	logs.Delete("/", func(c *fiber.Ctx) error {
		return logController.DeleteLogs(c, logHandler)
	})
}
