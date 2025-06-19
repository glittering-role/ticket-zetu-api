package routes

import (
	"ticket-zetu-api/logs/handler"
	discount_controller "ticket-zetu-api/modules/tickets/discount/controller"
	discount_service "ticket-zetu-api/modules/tickets/discount/services"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupDiscountRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db, logHandler)
	authService := authorization_service.NewPermissionService(db)

	discountService := discount_service.NewDiscountService(db, authService)
	discountController := discount_controller.NewDiscountController(discountService, logHandler)

	discountGroup := router.Group("/discounts", authMiddleware)
	{
		discountGroup.Get("/", discountController.GetDiscounts)
		discountGroup.Post("/", discountController.CreateDiscount)
		discountGroup.Get("/:id", discountController.GetDiscount)
		discountGroup.Put("/:id", discountController.UpdateDiscount)
		discountGroup.Patch("/:id/cancel", discountController.CancelDiscount)
	}

	// Public route for validation
	router.Get("/discounts/validate", discountController.ValidateDiscount)
}
