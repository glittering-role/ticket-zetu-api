package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	organizers "ticket-zetu-api/modules/organizers/controllers"
	organizers_services "ticket-zetu-api/modules/organizers/services"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func OrganizerRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinaryService *cloudinary.CloudinaryService) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)

	// Organizer Service and Controller
	organizerService := organizers_services.NewOrganizerService(db, authService)
	organizerController := organizers.NewOrganizerController(organizerService, logHandler)

	// Organization Image Service and Controller
	organizationImageService := organizers_services.NewOrganizationImageService(db, authService, cloudinaryService)
	organizationImageController := organizers.NewOrganizationImageController(organizationImageService, logHandler)

	// Subscription Service and Controller
	subscriptionService := organizers_services.NewSubscriptionService(db, authService)
	subscriptionController := organizers.NewSubscriptionController(subscriptionService, logHandler)

	organizerGroup := router.Group("/organizers", authMiddleware)
	{
		// Existing Organizer Routes
		organizerGroup.Get("/", organizerController.GetOrganizers)
		organizerGroup.Get("/my-organization", organizerController.GetMyOrganizer)
		organizerGroup.Get("/:id", organizerController.GetOrganizer)
		organizerGroup.Get("/search", organizerController.SearchOrganizers)

		organizerGroup.Post("/", organizerController.CreateOrganizer)
		organizerGroup.Put("/:id", organizerController.UpdateOrganizer)
		organizerGroup.Delete("/:id", organizerController.DeleteOrganizer)
		organizerGroup.Patch("/:id/deactivate", organizerController.DeactivateOrganizer)

		// Management actions
		organizerGroup.Patch("/:id/toggle-status", organizerController.ToggleOrganizerStatus)
		organizerGroup.Patch("/:id/flag", organizerController.FlagOrganizer)
		organizerGroup.Patch("/:id/ban", organizerController.BanOrganizer)
	}

	imageRoutes := router.Group("/organizations/:organization_id/image", authMiddleware)
	{
		imageRoutes.Post("/", organizationImageController.AddOrganizationImage)
		imageRoutes.Delete("/", organizationImageController.DeleteOrganizationImage)
	}

	subscriptionGroup := router.Group("/subscriptions", authMiddleware)
	{
		// Subscription Routes
		subscriptionGroup.Get("/", subscriptionController.GetSubscriptionsForUser)
		subscriptionGroup.Get("/my-subscribers", subscriptionController.GetSubscriptionsForOrganizer)
		subscriptionGroup.Post("/:organizer_id/subscribe", subscriptionController.Subscribe)
		subscriptionGroup.Delete("/:organizer_id/subscribe", subscriptionController.UnsubscribeFromOrganization)
		subscriptionGroup.Patch("/:organizer_id/preferences", subscriptionController.UpdatePreferences)
		subscriptionGroup.Patch("/subscribers/:subscriber_id/ban", subscriptionController.BanSubscriber)
	}
}
