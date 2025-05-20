package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authentication"

	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(api fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	userService := authentication.NewUserService(db)
	authController := authentication.NewAuthController(db, logHandler, userService)
	userNameCheck := authentication.CheckUsernameAvailability(db, logHandler)

	auth := api.Group("/auth")
	{
		auth.Post("/sign-up", authController.SignUp)
		auth.Post("/sign-in", authController.SignIn)
		auth.Get("/check-username", userNameCheck.CheckUsername)
		auth.Post("/logout", authController.Logout)
	}
}
