package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authorization"
	"ticket-zetu-api/modules/users/middleware"
)

func AuthorizationRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)

	roleService := authorization.NewRoleService(db)
	roleController := authorization.NewRoleController(roleService, logHandler)

	permissionService := authorization.NewPermissionService(db)
	permissionController := authorization.NewPermissionController(permissionService, logHandler, db)

	roleGroup := router.Group("/roles", authMiddleware)
	{
		roleGroup.Post("/", roleController.CreateRole)
		roleGroup.Get("/", roleController.GetRoles)
		roleGroup.Get("/:id", roleController.GetRole)
		roleGroup.Put("/:id", roleController.UpdateRole)
		roleGroup.Post("/assign", roleController.AssignRoleToUser)
		roleGroup.Delete("/:id", roleController.DeleteRole)
	}

	permissionGroup := router.Group("/permissions", authMiddleware)
	{
		permissionGroup.Post("/", permissionController.CreatePermission)
		permissionGroup.Get("/", permissionController.GetPermissions)
		permissionGroup.Get("/:id", permissionController.GetPermission)
		permissionGroup.Put("/:id", permissionController.UpdatePermission)
		permissionGroup.Delete("/:id", permissionController.DeletePermission)
		permissionGroup.Post("/assign", permissionController.AssignPermissionToRole)
		permissionGroup.Post("/remove", permissionController.RemovePermissionFromRole)
	}
}
