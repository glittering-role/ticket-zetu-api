package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	"ticket-zetu-api/logs/handler"
	logs "ticket-zetu-api/logs/routes/v1"
	"ticket-zetu-api/logs/service"

	categories "ticket-zetu-api/modules/events/routes/v1/category"
	events "ticket-zetu-api/modules/events/routes/v1/events"
	venue "ticket-zetu-api/modules/events/routes/v1/venues"

	organization "ticket-zetu-api/modules/organizers/routes/v1"
	price_tier "ticket-zetu-api/modules/tickets/routes/v1"
	ticket_type "ticket-zetu-api/modules/tickets/routes/v1"

	auth "ticket-zetu-api/modules/users/routes/v1"
	roles "ticket-zetu-api/modules/users/routes/v1"
	users "ticket-zetu-api/modules/users/routes/v1"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	_ "ticket-zetu-api/docs" // Import the generated docs package for Swagger
)

// @title Ticket Zetu API
// @version 1.0
// @description This is the API documentation for Ticket Zetu application
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@ticketzetu.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080/api/v1
// @BasePath /

func main() {
	// Load configuration
	appConfig := config.LoadConfig()

	// Initialize Cloudinary
	cloudinaryService, err := cloudinary.NewCloudinaryService(appConfig.Cloudinary)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Run migrations
	if err := database.Migrate(database.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize log service and handler
	logService := service.NewLogService(database.DB, 100, 5*time.Second, appConfig.Env)
	defer logService.Shutdown()
	logHandler := &handler.LogHandler{Service: logService}

	// Initialize Fiber app
	app := fiber.New()

	// Apply global CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Apply global rate limiter middleware
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

	app.Get("/swagger/*", swagger.New()) // default

	// Create API group
	api := app.Group("/api/v1")

	// Register all module routes
	logs.SetupRoutes(api, logService, logHandler)
	roles.AuthorizationRoutes(api, database.DB, logHandler)
	auth.SetupAuthRoutes(api, database.DB, logHandler)
	users.UserRoutes(api, database.DB, logHandler, cloudinaryService)
	categories.CategoryRoutes(api, database.DB, logHandler, cloudinaryService)
	organization.OrganizerRoutes(api, database.DB, logHandler, cloudinaryService)
	venue.VenueRoutes(api, database.DB, logHandler, cloudinaryService)
	events.SetupEventsRoutes(api, database.DB, logHandler, cloudinaryService)
	price_tier.SetupPriceTierRoutes(api, database.DB, logHandler)
	ticket_type.SetupTicketTypeRoutes(api, database.DB, logHandler)

	// Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		log.Println("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting server on port %s...", appConfig.Port)
	if err := app.Listen(":" + appConfig.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
