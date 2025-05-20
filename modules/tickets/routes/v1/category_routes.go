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
	categoryController := category.NewCategoryController(categoryService, logHandler)
	subcategoryController := category.NewSubcategoryController(categoryService, logHandler)

	categoryGroup := router.Group("/categories", authMiddleware)
	{
		// Category routes
		categoryGroup.Get("/", categoryController.GetCategories)
		categoryGroup.Get("/:id", categoryController.GetCategory)
		categoryGroup.Get("/:id/subcategories", categoryController.GetSubcategories)
		categoryGroup.Post("/", categoryController.CreateCategory)
		categoryGroup.Put("/:id", categoryController.UpdateCategory)
		categoryGroup.Delete("/:id", categoryController.DeleteCategory)
	}

	subCategoryGroup := router.Group("/sub_categories", authMiddleware)
	{
		// Subcategory routes
		subCategoryGroup.Post("/:id/subcategories", subcategoryController.CreateSubcategory)
		subCategoryGroup.Put("/subcategories/:id", subcategoryController.UpdateSubcategory)
		subCategoryGroup.Delete("/subcategories/:id", subcategoryController.DeleteSubcategory)
	}
}
