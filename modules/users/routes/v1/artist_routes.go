package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/artist"
	"ticket-zetu-api/modules/users/members/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ArtistRoutes registers routes for artist profile operations
func ArtistRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db, logHandler)

	artistService := service.NewArtistService(db)
	artistController := artist.NewArtistController(artistService, logHandler)

	artistGroup := router.Group("/artist-profile", authMiddleware)
	{
		artistGroup.Get("/", artistController.GetArtistProfile)
		artistGroup.Post("", artistController.CreateArtistProfile)
		artistGroup.Patch("/", artistController.UpdateArtistProfile)
		artistGroup.Delete("/", artistController.DeleteArtistProfile)
	}
}
