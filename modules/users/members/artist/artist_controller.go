package artist

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/members/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ArtistController handles HTTP requests for artist profile operations
type ArtistController struct {
	artistService service.ArtistService
	logHandler    *handler.LogHandler
	validator     *validator.Validate
}

// NewArtistController initializes a new ArtistController
func NewArtistController(artistService service.ArtistService, logHandler *handler.LogHandler) *ArtistController {
	return &ArtistController{
		artistService: artistService,
		logHandler:    logHandler,
		validator:     validator.New(),
	}
}

// GetArtistProfile godoc
// @Summary Get an artist profile
// @Description Retrieves an artist profile by ID or UserID.
// @Tags Artist Profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Artist profile retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid identifier"
// @Failure 401 {object} map[string]interface{} "Unauthorized: user ID not found"
// @Failure 404 {object} map[string]interface{} "Artist profile not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router  /artist-profile [get]
func (c *ArtistController) GetArtistProfile(ctx *fiber.Ctx) error {
	// Get userID from context (logged-in user)
	userID := ctx.Locals("user_id").(string)
	if userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusUnauthorized, "unauthorized: user ID not found"), fiber.StatusUnauthorized)
	}

	result, err := c.artistService.GetArtistProfileByUserID(userID)
	if err != nil {
		if err.Error() == "artist profile not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, result, "Artist profile retrieved successfully", true)
}
