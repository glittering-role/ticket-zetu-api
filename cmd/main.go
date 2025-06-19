package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"ticket-zetu-api/config"
	"ticket-zetu-api/internal/middleware"
	"ticket-zetu-api/internal/services"
)

// @title Ticket Zetu API
// @version 1.0
// @description This is the API documentation for Ticket Zetu application. The API host is configurable via the API_URL environment variable (e.g., https://<your-ngrok-url>.ngrok-free.app).
// @termsOfService https://ticketzetu.com/terms
// @contact.name API Support
// @contact.email support@ticketzetu.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1

func main() {
	// Load configuration
	appConfig := config.LoadConfig()

	// Initialize Fiber app
	app := fiber.New()

	// Setup services
	cloudinaryService, db, logService, emailService, jobQueue, err := services.SetupServices(appConfig)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.ShutdownServices(db, logService, emailService, jobQueue)

	// Setup middleware
	middleware.SetupMiddleware(app, appConfig)

	// Setup routes
	services.SetupRoutes(app, db, logService, cloudinaryService, emailService)

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
	log.Printf("Starting server on port %s with API_URL %s...", appConfig.Port, appConfig.ApiUrl)
	if err := app.Listen(":" + appConfig.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
