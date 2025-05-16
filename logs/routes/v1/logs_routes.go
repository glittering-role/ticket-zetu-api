package routes

import (
	"ticket-zetu-api/logs/controllers"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/logs/service"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes registers all routes, initializing controllers internally
func SetupRoutes(api fiber.Router, logService *service.LogService, logHandler *handler.LogHandler) {
	// Initialize log controller
	logController := controller.NewLogController(logService)

	// Create logs group
	logs := api.Group("/logs")

	// Log management routes
	logs.Get("/", func(c *fiber.Ctx) error {
		return logController.GetLogs(c, logHandler)
	})
	logs.Delete("/", func(c *fiber.Ctx) error {
		return logController.DeleteLogs(c, logHandler)
	})
}

// package routes

// import (
// 	"ticket-zetu-api/logs/controllers" // Fixed from controllers to controller
// 	"ticket-zetu-api/logs/handler"
// 	"ticket-zetu-api/logs/service"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// // SetupRoutes registers all routes, initializing controllers internally
// func SetupRoutes(api *fiber.Router, logService *service.LogService, logHandler *handler.LogHandler, db *gorm.DB) {
// 	// Initialize log controller
// 	logController := controller.NewLogController(logService, db)

// 		// Create logs group
// 	logs := api.Group("/logs")

// 	// Log management routes
// 	logs.Get("/", func(c *fiber.Ctx) error {
// 		return logController.GetLogs(c, logHandler)
// 	})
// 	logs.Delete("/", func(c *fiber.Ctx) error {
// 		return logController.DeleteLogs(c, logHandler)
// 	})
// }
