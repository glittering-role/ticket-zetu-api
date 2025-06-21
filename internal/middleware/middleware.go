package middleware

import (
	"ticket-zetu-api/config"
	_ "ticket-zetu-api/docs"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/helpers"
	"time"

	"github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupMiddleware(app *fiber.App, cfg *config.AppConfig, logHandler *handler.LogHandler) {
	// Initialize services
	geoService := helpers.NewGeolocationService(logHandler, cfg.ApiToken)
	deviceService := helpers.NewDeviceDetectionService(logHandler)

	// Apply middlewares
	app.Use(geoService.GeolocationMiddleware())
	app.Use(deviceService.DeviceDetectionMiddleware())

	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.ApiUrl,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Apply rate limiter middleware
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Rate limit exceeded. Please try again later.",
			})
		},
	}))

	// Configure Swagger
	swaggerConfig := swagger.Config{
		URL: cfg.ApiUrl + "/swagger/doc.json",
	}

	app.Get("/swagger/*", swagger.New(swaggerConfig))
}
