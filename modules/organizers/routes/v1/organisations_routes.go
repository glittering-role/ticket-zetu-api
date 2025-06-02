package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	organizers "ticket-zetu-api/modules/organizers/controllers"
	organizers_services "ticket-zetu-api/modules/organizers/services"
	"ticket-zetu-api/modules/users/authorization"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func OrganizerRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinaryService *cloudinary.CloudinaryService) {

	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization.NewPermissionService(db)
	organizerService := organizers_services.NewOrganizerService(db, authService)
	organizerController := organizers.NewOrganizerController(organizerService, logHandler)

	// Organization Image Routes
	organizationImageService := organizers_services.NewOrganizationImageService(db, authService, cloudinaryService)

	organizationImageController := organizers.NewOrganizationImageController(organizationImageService, logHandler)

	organizerGroup := router.Group("/organizers", authMiddleware)
	{
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

}
