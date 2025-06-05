package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/account"
	"ticket-zetu-api/modules/users/members/preference"
	members_service "ticket-zetu-api/modules/users/members/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UserRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db)

	userService := members_service.NewUserService(db)
	userController := account.NewUserController(userService, logHandler)

	preferencesService := members_service.NewUserPreferencesService(db)
	preferencesController := preference.NewUserPreferencesController(preferencesService, logHandler)

	userGroup := router.Group("/users", authMiddleware)
	{
		userGroup.Get("/me", userController.GetMyProfile)
		userGroup.Get("/:identifier", userController.GetUserProfile)
		userGroup.Post("/me/details", userController.UpdateDetails)
		userGroup.Post("/me/location", userController.UpdateLocation)
		userGroup.Post("/me/phone", userController.UpdatePhone)
		userGroup.Post("/me/email", userController.UpdateEmail)
	}

	userGroupPr := router.Group("/users/me", authMiddleware)
	{
		userGroupPr.Get("/preferences", preferencesController.GetUserPreferences)
		userGroupPr.Post("/preferences", preferencesController.UpdateUserPreferences)
	}
}
