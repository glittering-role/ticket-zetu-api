package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthorizationRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	roleService := authorization.NewRoleService(db)
	roleController := authorization.NewRoleController(roleService, logHandler)
	permissionService := authorization.NewPermissionService(db)
	permissionController := authorization.NewPermissionController(permissionService, logHandler, db)

	roleGroup := router.Group("/roles")
	{
		roleGroup.Post("/", roleController.CreateRole)
		roleGroup.Get("/", roleController.GetRoles)
		roleGroup.Get("/:id", roleController.GetRole)
		roleGroup.Put("/:id", roleController.UpdateRole)
		roleGroup.Delete("/:id", roleController.DeleteRole)
		roleGroup.Post("/permissions", roleController.AssignPermissionToRole)
	}

	permissionGroup := router.Group("/permissions")
	{
		permissionGroup.Post("/", permissionController.CreatePermission)
		permissionGroup.Get("/", permissionController.GetPermissions)
		permissionGroup.Get("/:id", permissionController.GetPermission)
		permissionGroup.Put("/:id", permissionController.UpdatePermission)
		permissionGroup.Delete("/:id", permissionController.DeletePermission)
		permissionGroup.Post("/assign", permissionController.AssignPermissionToRole)
	}
}
