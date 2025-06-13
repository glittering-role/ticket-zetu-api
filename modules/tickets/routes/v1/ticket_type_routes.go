// routes/ticket_type_routes.go
package routes

import (
	"ticket-zetu-api/logs/handler"

	ticket_type_controller "ticket-zetu-api/modules/tickets/ticket_type/controller"
	ticket_type_service "ticket-zetu-api/modules/tickets/ticket_type/service"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupTicketTypeRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)

	ticketTypeService := ticket_type_service.NewTicketTypeService(db, authService)
	ticketTypeController := ticket_type_controller.NewTicketTypeController(ticketTypeService, logHandler)

	ticketTypeGroup := router.Group("/ticket-types", authMiddleware)
	{
		ticketTypeGroup.Get("/organization", ticketTypeController.GetAllTicketTypesForOrganization)
		ticketTypeGroup.Get("/event/:event_id", ticketTypeController.GetTicketTypesForEvent)
		ticketTypeGroup.Get("/:id", ticketTypeController.GetSingleTicketType)
		ticketTypeGroup.Post("/", ticketTypeController.CreateTicketType)
		ticketTypeGroup.Put("/:id", ticketTypeController.UpdateTicketType)
		ticketTypeGroup.Delete("/:id", ticketTypeController.DeleteTicketType)
		ticketTypeGroup.Post("/:ticket_type_id/price-tiers", ticketTypeController.AssociatePriceTier)
		ticketTypeGroup.Delete("/:ticket_type_id/price-tiers/:price_tier_id", ticketTypeController.DisassociatePriceTier)
	}

}
