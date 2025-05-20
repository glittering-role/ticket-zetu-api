package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/account"
	members_service "ticket-zetu-api/modules/users/members/service"

	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UserRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)

	userService := members_service.NewUserService(db)
	userController := account.NewUserController(userService, logHandler)

	userGroup := router.Group("/users", authMiddleware)
	{
		userGroup.Get("/me", userController.GetMyProfile)
		userGroup.Get("/:id", userController.GetUserProfile)
	}
}
