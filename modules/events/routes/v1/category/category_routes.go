package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	category "ticket-zetu-api/modules/events/category/controller"
	"ticket-zetu-api/modules/events/category/services"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CategoryRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService) {
	authMiddleware := middleware.IsAuthenticated(db)
	authService := authorization_service.NewPermissionService(db)

	// Category service and controller
	categoryService := services.NewCategoryService(db, authService)
	categoryController := category.NewCategoryController(categoryService, logHandler)

	// Subcategory service and controller
	subcategoryService := services.NewSubcategoryService(db, authService)
	subcategoryController := category.NewSubcategoryController(subcategoryService, logHandler)

	// Image service and controller
	imageService := services.NewImageService(db, authService, cloudinary)
	imageController := category.NewImageController(imageService, logHandler)

	categoryGroup := router.Group("/categories", authMiddleware)
	{
		// Category routes
		categoryGroup.Get("/", categoryController.GetCategories)
		categoryGroup.Get("/all", categoryController.GetAllCategoriesWithTheirSubCategories)
		categoryGroup.Get("/:id", categoryController.GetCategory)
		categoryGroup.Post("/", categoryController.CreateCategory)
		categoryGroup.Delete("/:id", categoryController.DeleteCategory)
		categoryGroup.Put("/:id/toggle-status", categoryController.ToggleCategoryStatus)
		categoryGroup.Put("/:id", categoryController.UpdateCategory)

	}

	subCategoryGroup := router.Group("/subcategories", authMiddleware)
	{
		// Subcategory routes
		subCategoryGroup.Get("/:id/subcategories", subcategoryController.GetSubcategories)
		subCategoryGroup.Post("/", subcategoryController.CreateSubcategory)
		subCategoryGroup.Put("/:id", subcategoryController.UpdateSubcategory)
		subCategoryGroup.Delete("/:id", subcategoryController.DeleteSubcategory)
		subCategoryGroup.Put("/:id/toggle-status", subcategoryController.ToggleSubcategoryStatus)
	}

	imageGroup := router.Group("/category_images", authMiddleware)
	{
		// Image routes
		imageGroup.Post("/:entity_type/:entity_id", imageController.AddImage)
		imageGroup.Delete("/:entity_type/:entity_id", imageController.DeleteImage)
	}
}
