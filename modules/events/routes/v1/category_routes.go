package routes

import (
	"ticket-zetu-api/logs/handler"
	category "ticket-zetu-api/modules/events/category/controller"
	"ticket-zetu-api/modules/events/category/services"
	"ticket-zetu-api/modules/users/authorization"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CategoryRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization.NewPermissionService(db)

	// Category service and controller
	categoryService := services.NewCategoryService(db, authService)
	categoryController := category.NewCategoryController(categoryService, logHandler)

	// Subcategory service and controller
	subcategoryService := services.NewSubcategoryService(db, authService)
	subcategoryController := category.NewSubcategoryController(subcategoryService, logHandler)

	categoryGroup := router.Group("/categories", authMiddleware)
	{
		// Category routes
		categoryGroup.Get("/", categoryController.GetCategories)
		categoryGroup.Get("/:id", categoryController.GetCategory)
		categoryGroup.Post("/", categoryController.CreateCategory)
		categoryGroup.Put("/:id", categoryController.UpdateCategory)
		categoryGroup.Delete("/:id", categoryController.DeleteCategory)
	}

	subCategoryGroup := router.Group("/sub_categories", authMiddleware)
	{
		// Subcategory routes
		subCategoryGroup.Get("/", subcategoryController.GetSubcategories)
		subCategoryGroup.Post("/", subcategoryController.CreateSubcategory)
		subCategoryGroup.Put("/subcategories/:id", subcategoryController.UpdateSubcategory)
		subCategoryGroup.Delete("/subcategories/:id", subcategoryController.DeleteSubcategory)
	}
}
