package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/tickets/category"
	"ticket-zetu-api/modules/users/authorization"
	"ticket-zetu-api/modules/users/middleware"
)

func CategoryRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization.NewPermissionService(db)
	categoryService := category.NewCategoryService(db, authService)
	category.NewCategoryController(categoryService, logHandler)
	categoryController := category.NewCategoryController(categoryService, logHandler)

	categoryGroup := router.Group("/categories", authMiddleware)
	{
		categoryGroup.Get("/", categoryController.GetCategories)
		categoryGroup.Get("/:id", categoryController.GetCategory)
		categoryGroup.Get("/:id/subcategories", categoryController.GetSubcategories)

	}
}
