package routes

import (
	"ticket-zetu-api/logs/handler"

	"gorm.io/gorm"
	"ticket-zetu-api/modules/users/authentication/controllers"
	"ticket-zetu-api/modules/users/authentication/service"
	"ticket-zetu-api/modules/users/authentication/utils"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(api fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	userService := auth_service.NewUserService(db)
	authController := authentication.NewAuthController(db, logHandler, userService)
	userNameCheck := auth_utils.CheckUsernameAvailability(db, logHandler)

	auth := api.Group("/auth")
	{
		auth.Post("/sign-up", authController.SignUp)
		auth.Post("/sign-in", authController.SignIn)
		auth.Get("/check-username", userNameCheck.CheckUsername)
		auth.Post("/logout", authController.Logout)
	}
}
