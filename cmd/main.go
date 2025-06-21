package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	"ticket-zetu-api/internal/middleware"
	"ticket-zetu-api/internal/services"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/logs/service"
	"time"
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

	// Initialize database
	database.InitDB()
	db := database.DB
	if db == nil {
		log.Fatalf("Database initialization failed: DB is nil")
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize LogService
	logService := service.NewLogService(db, 100, 5*time.Second, appConfig.Env)
	logHandler := &handler.LogHandler{Service: logService}

	// Setup services
	cloudinaryService, emailService, jobQueue, err := services.SetupServices(appConfig, db, logService, logHandler)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.ShutdownServices(db, logService, emailService, jobQueue)

	// Setup middleware
	middleware.SetupMiddleware(app, appConfig, logHandler)

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
