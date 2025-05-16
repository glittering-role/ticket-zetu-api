package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	"ticket-zetu-api/logs/handler"
	logs "ticket-zetu-api/logs/routes/v1"
	"ticket-zetu-api/logs/service"
	roles "ticket-zetu-api/modules/users/routes/v1"

	"github.com/gofiber/fiber/v2"
)

// @title           Ticket Zetu API
// @version         1.0
// @description     This is a log management API for Ticket Zetu.
// @host            localhost:8080
// @BasePath        /api/v1

// @contact.name   API Support
// @contact.email  support@ticketzetu.com

func main() {
	// Load configuration
	appConfig := config.LoadConfig()

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
