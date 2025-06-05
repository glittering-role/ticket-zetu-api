package routes

import (
	"ticket-zetu-api/logs/handler"
	tickets_controller "ticket-zetu-api/modules/tickets/price_tires/controller"
	price_tier_service "ticket-zetu-api/modules/tickets/price_tires/service"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupPriceTierRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)

	priceTierService := price_tier_service.NewPriceTierService(db, authService)
	priceTierController := tickets_controller.NewPriceTierController(priceTierService, logHandler)

	priceTierGroup := router.Group("/price-tiers", authMiddleware)
	{
		priceTierGroup.Get("/organizer/:organizer_id", priceTierController.GetPriceTiersForOrganizer)
		priceTierGroup.Get("/:id", priceTierController.GetSinglePriceTier)
		priceTierGroup.Get("/", priceTierController.GetAllPriceTiers)
		priceTierGroup.Post("/", priceTierController.CreatePriceTier)
		priceTierGroup.Put("/:id", priceTierController.UpdatePriceTier)
		priceTierGroup.Delete("/:id", priceTierController.DeletePriceTier)
	}
}
