package routes

import (
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/logs/handler"
	mail_service "ticket-zetu-api/modules/users/authentication/mail"
	auth_utils "ticket-zetu-api/modules/users/authentication/utils"
	"ticket-zetu-api/modules/users/members/account"
	"ticket-zetu-api/modules/users/members/preference"
	members_service "ticket-zetu-api/modules/users/members/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UserRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler, cloudinary *cloudinary.CloudinaryService, emailService mail_service.EmailService) {
	authMiddleware := middleware.IsAuthenticated(db)
	userNameCheck := auth_utils.NewUsernameCheck(db, logHandler)

	userService := members_service.NewUserService(db, emailService, userNameCheck)
	userController := account.NewUserController(userService, logHandler)

	preferencesService := members_service.NewUserPreferencesService(db)
	preferencesController := preference.NewUserPreferencesController(preferencesService, logHandler)

	profileImageService := members_service.NewImageService(db, cloudinary)
	profileImageController := account.NewImageController(profileImageService, logHandler)

	userGroup := router.Group("/users", authMiddleware)
	{
		userGroup.Get("/me", userController.GetMyProfile)
		userGroup.Get("/:identifier", userController.GetUserProfile)
		userGroup.Post("/me/details", userController.UpdateDetails)
		userGroup.Post("/me/location", userController.UpdateLocation)
		userGroup.Post("/me/phone", userController.UpdatePhone)
		userGroup.Post("/me/email", userController.UpdateEmail)
		userGroup.Post("/me/username", userController.UpdateUsername)
		userGroup.Post("/me/password", userController.SetNewPassword)
	}

	userGroupPr := router.Group("/users/me", authMiddleware)
	{
		userGroupPr.Get("/preferences", preferencesController.GetUserPreferences)
		userGroupPr.Post("/preferences", preferencesController.UpdateUserPreferences)
	}

	userGroupAvt := router.Group("/users/me", authMiddleware)
	{
		userGroupAvt.Post("/image", profileImageController.UploadProfileImage)
		userGroupAvt.Delete("/image", profileImageController.DeleteProfileImage)
	}
}
