package artist

import (
	"strings"
	"ticket-zetu-api/modules/users/members/dto"

	"github.com/gofiber/fiber/v2"
)

// CreateArtistProfile godoc
// @Summary Create a new artist profile
// @Description Creates a new artist profile for the authenticated user. Fields like genres, skills, and featured_works can be a single string or an array of strings.
// @Tags Artist Profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateArtistProfileDTO true "Artist profile details"
// @Success 200 {object} map[string]interface{} "Artist profile created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized: user ID not found"
// @Failure 403 {object} map[string]interface{} "Unauthorized: can only create profile for self"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /artist-profile [post]
func (c *ArtistController) CreateArtistProfile(ctx *fiber.Ctx) error {
	var artistDto dto.CreateArtistProfileDTO
	if err := ctx.BodyParser(&artistDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error()), fiber.StatusBadRequest)
	}

	// Validate DTO
	if err := c.validator.Struct(&artistDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	// Get requesterID from context
	requesterID := ctx.Locals("user_id")
	if requesterID == nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusUnauthorized, "unauthorized: user ID not found"), fiber.StatusUnauthorized)
	}

	_, err := c.artistService.CreateArtistProfile(&artistDto, requesterID.(string))
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "user already has an artist profile" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if strings.HasPrefix(err.Error(), "validation failed") {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Artist profile created successfully", true)
}
