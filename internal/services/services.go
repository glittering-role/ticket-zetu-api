package services

import (
	"fmt"
	"log"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	"ticket-zetu-api/logs/handler"
	logs "ticket-zetu-api/logs/routes/v1"
	"ticket-zetu-api/logs/service"
	"ticket-zetu-api/mail"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"
	"ticket-zetu-api/queue"
	"time"

	events "ticket-zetu-api/modules/events/routes/v1"
	notifications "ticket-zetu-api/modules/notifications/routes/v1"
	organization "ticket-zetu-api/modules/organizers/routes/v1"
	tickets "ticket-zetu-api/modules/tickets/routes/v1"
	user "ticket-zetu-api/modules/users/routes/v1"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupServices(cfg *config.AppConfig) (*cloudinary.CloudinaryService, *gorm.DB, *service.LogService, mail_service.EmailService, *queue.JobQueue, error) {
	// Initialize Cloudinary
	cloudinaryService, err := cloudinary.NewCloudinaryService(cfg.Cloudinary)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary: %v", err)
		return nil, nil, nil, nil, nil, err
	}

	// Initialize database
	database.InitDB()
	db := database.DB
	if db == nil {
		log.Println("Database initialization failed: DB is nil")
		return nil, nil, nil, nil, nil, fmt.Errorf("database initialization failed")
	}
	if err := database.Migrate(db); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		return nil, nil, nil, nil, nil, err
	}

	// Initialize job queue
	jobQueue := queue.NewJobQueue(5)

	// Initialize log service
	logService := service.NewLogService(db, 100, 5*time.Second, cfg.Env)
	logHandler := &handler.LogHandler{Service: logService}

	// Initialize email service
	emailConfig, err := mail.NewConfig(logHandler)
	if err != nil {
		log.Printf("Failed to initialize email config: %v", err)
		return nil, nil, nil, nil, nil, err
	}
	emailService := mail_service.NewEmailService(emailConfig, logHandler, 5)

	// Initialize Redis
	database.InitRedis()
	redisClient := database.GetRedisClient()
	if redisClient == nil {
		log.Println("Redis initialization failed: redisClient is nil")
		return nil, nil, nil, nil, nil, fmt.Errorf("redis initialization failed")
	}
	database.SetRedisClient(redisClient)

	return cloudinaryService, db, logService, emailService, jobQueue, nil
}

func ShutdownServices(db *gorm.DB, logService *service.LogService, emailService mail_service.EmailService, jobQueue *queue.JobQueue) {
	if db != nil {
		database.CloseDB()
	}
	if logService != nil {
		logService.Shutdown()
	}
	if emailService != nil {
		emailService.Shutdown()
	}
	if jobQueue != nil {
		jobQueue.Close()
	}
	redisClient := database.GetRedisClient()
	if redisClient != nil {
		database.CloseRedis(redisClient)
	}
}

func SetupRoutes(app *fiber.App, db *gorm.DB, logService *service.LogService, cloudinaryService *cloudinary.CloudinaryService, emailService mail_service.EmailService) {
	api := app.Group("/api/v1")
	logs.SetupRoutes(api, logService, &handler.LogHandler{Service: logService})
	user.SetupUsersMainRoutes(api, db, database.GetRedisClient(), &handler.LogHandler{Service: logService}, cloudinaryService, emailService)
	events.SetupEventsMainRoutes(api, db, &handler.LogHandler{Service: logService}, cloudinaryService, emailService)
	organization.SetupOrganizationMainRoutes(api, db, &handler.LogHandler{Service: logService}, cloudinaryService, emailService)
	events.SetupEventsRoutes(api, db, &handler.LogHandler{Service: logService}, cloudinaryService)
	tickets.SetupTicketMainRoutes(api, db, &handler.LogHandler{Service: logService}, cloudinaryService, emailService)
	notifications.SetupNotificationMainRoutes(api, db, &handler.LogHandler{Service: logService}, emailService)
}
