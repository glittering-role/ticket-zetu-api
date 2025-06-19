package routes

import (
	"github.com/redis/go-redis/v9"
	"ticket-zetu-api/logs/handler"
	authentication "ticket-zetu-api/modules/users/authentication/controllers"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"
	auth_service "ticket-zetu-api/modules/users/authentication/service"
	auth_utils "ticket-zetu-api/modules/users/authentication/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupAuthRoutes(api fiber.Router, db *gorm.DB, redisClient *redis.Client, logHandler *handler.LogHandler, emailService mail_service.EmailService) {
	userService := auth_service.NewUserService(db, redisClient, logHandler, emailService)
	authController, err := authentication.NewAuthController(db, emailService, userService, logHandler)
	if err != nil {
		panic(err)
	}
	userNameCheck := auth_utils.NewUsernameCheck(db, logHandler)

	auth := api.Group("/auth")
	{
		auth.Post("/signup", authController.SignUp)
		auth.Post("/signin", authController.SignIn)
		auth.Get("/check-username", userNameCheck.CheckUsername)
		auth.Post("/logout", authController.Logout)
		auth.Post("/verify-email", authController.VerifyEmail)
		auth.Post("/reset-password-request", authController.RequestPasswordReset)
		auth.Post("/reset-password", authController.SetNewPassword)
	}
}
