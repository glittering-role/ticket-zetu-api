package services

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	"ticket-zetu-api/logs/handler"
	logs "ticket-zetu-api/logs/routes/v1"
	"ticket-zetu-api/logs/service"
	"ticket-zetu-api/mail"
	events "ticket-zetu-api/modules/events/routes/v1"
	notifications "ticket-zetu-api/modules/notifications/routes/v1"
	organization "ticket-zetu-api/modules/organizers/routes/v1"
	tickets "ticket-zetu-api/modules/tickets/routes/v1"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"
	"ticket-zetu-api/modules/users/helpers"
	user "ticket-zetu-api/modules/users/routes/v1"
	"ticket-zetu-api/queue"
)

func SetupServices(cfg *config.AppConfig, db *gorm.DB, logService *service.LogService, logHandler *handler.LogHandler) (*cloudinary.CloudinaryService, mail_service.EmailService, *queue.JobQueue, *helpers.GeolocationService, *helpers.DeviceDetectionService, error) {
	// Initialize Cloudinary
	cloudinaryService, err := cloudinary.NewCloudinaryService(cfg.Cloudinary)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary: %v", err)
		return nil, nil, nil, nil, nil, err
	}

	// Initialize job queue
	jobQueue := queue.NewJobQueue(5)

	// Initialize email service
	emailConfig, err := mail.NewConfig(logHandler)
	if err != nil {
		log.Printf("Failed to initialize email config: %v", err)
		return nil, nil, nil, nil, nil, err
	}
	emailService := mail_service.NewEmailService(emailConfig, logHandler, 5)

	// Initialize GeolocationService
	geoService := helpers.NewGeolocationService(logHandler, cfg.ApiToken)

	// Initialize DeviceDetectionService
	deviceService := helpers.NewDeviceDetectionService(logHandler)

	// Initialize Redis
	database.InitRedis()
	redisClient := database.GetRedisClient()
	if redisClient == nil {
		log.Println("Redis initialization failed: redisClient is nil")
		return nil, nil, nil, nil, nil, fmt.Errorf("redis initialization failed")
	}
	database.SetRedisClient(redisClient)

	return cloudinaryService, emailService, jobQueue, geoService, deviceService, nil
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

func SetupRoutes(app *fiber.App, db *gorm.DB, logService *service.LogService, cloudinaryService *cloudinary.CloudinaryService, emailService mail_service.EmailService, geoService *helpers.GeolocationService, deviceService *helpers.DeviceDetectionService) {
	api := app.Group("/api/v1")
	logHandler := &handler.LogHandler{Service: logService}

	logs.SetupRoutes(api, logService, logHandler)
	user.SetupUsersMainRoutes(api, db, database.GetRedisClient(), logHandler, cloudinaryService, emailService, geoService, deviceService)
	events.SetupEventsMainRoutes(api, db, logHandler, cloudinaryService, emailService)
	organization.SetupOrganizationMainRoutes(api, db, logHandler, cloudinaryService, emailService)
	events.SetupEventsRoutes(api, db, logHandler, cloudinaryService)
	tickets.SetupTicketMainRoutes(api, db, logHandler, cloudinaryService, emailService)
	notifications.SetupNotificationMainRoutes(api, db, logHandler, emailService)
}
