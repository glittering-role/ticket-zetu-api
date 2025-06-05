package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	venues_controller "ticket-zetu-api/modules/events/venues/controller"
	service "ticket-zetu-api/modules/events/venues/service"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func VenueRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)
	venueService := service.NewVenueService(db, authService, cloudinary)
	venueController := venues_controller.NewVenueController(venueService, logHandler, cloudinary)

	venueGroup := router.Group("/venues", authMiddleware)
	{
		// Venue routes for organizers
		venueGroup.Get("/all", venueController.GetAllVenue)
		venueGroup.Get("/", venueController.GetVenuesForOrganizer)
		venueGroup.Get("/:id", venueController.GetSingleVenueForOrganizer)
		venueGroup.Post("/", venueController.CreateVenue)
		venueGroup.Put("/:id", venueController.UpdateVenue)
		venueGroup.Delete("/:id", venueController.DeleteVenue)
		venueGroup.Post("/:id/images", venueController.AddVenueImage)
		venueGroup.Delete("/:venue_id/images/:image_id", venueController.DeleteVenueImage)
	}
}
