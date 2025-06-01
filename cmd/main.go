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
	categories "ticket-zetu-api/modules/events/routes/v1"
	events "ticket-zetu-api/modules/events/routes/v1"
	venue "ticket-zetu-api/modules/events/routes/v1"
	organization "ticket-zetu-api/modules/organizers/routes/v1"
	price_tier "ticket-zetu-api/modules/tickets/routes/v1"
	ticket_type "ticket-zetu-api/modules/tickets/routes/v1"
	auth "ticket-zetu-api/modules/users/routes/v1"
	roles "ticket-zetu-api/modules/users/routes/v1"
	users "ticket-zetu-api/modules/users/routes/v1"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	appConfig := config.LoadConfig()

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

	// Initialize log service
	logService := service.NewLogService(database.DB, 100, 5*time.Second, appConfig.Env)
	defer logService.Shutdown()

	// Initialize log handler
	logHandler := &handler.LogHandler{Service: logService}

	// Initialize Fiber app
	app := fiber.New()

	// Create API group
	api := app.Group("/api/v1")

	// Register log management routes
	logs.SetupRoutes(api, logService, logHandler)
	roles.AuthorizationRoutes(api, database.DB, logHandler)
	auth.SetupAuthRoutes(api, database.DB, logHandler)
	users.UserRoutes(api, database.DB, logHandler)
	categories.CategoryRoutes(api, database.DB, logHandler, cloudinaryService)
	organization.OrganizerRoutes(api, database.DB, logHandler)
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
