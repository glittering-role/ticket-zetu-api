package artist

import (
	"github.com/gofiber/fiber/v2"
	"ticket-zetu-api/modules/users/members/dto"
)

// UpdateArtistProfile godoc
// @Summary Update artist profile
// @Description Updates the authenticated user's artist profile
// @Tags Artist Profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.UpdateArtistProfileDTO true "Artist profile details to update"
// @Success 200 {object} map[string]interface{} "Artist profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Artist profile not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /artist-profile [patch]
func (c *ArtistController) UpdateArtistProfile(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	if userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusUnauthorized, "unauthorized: user ID not found"), fiber.StatusUnauthorized)
	}

	var artistDto dto.UpdateArtistProfileDTO
	if err := ctx.BodyParser(&artistDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body: "+err.Error()), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&artistDto); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	_, err := c.artistService.UpdateArtistProfile(userID, &artistDto)
	if err != nil {
		if err.Error() == "artist profile not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Artist profile updated successfully", true)
}

// DeleteArtistProfile godoc
// @Summary Delete artist profile
// @Description Deletes the authenticated user's artist profile
// @Tags Artist Profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Artist profile deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Artist profile not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /artist-profile [delete]
func (c *ArtistController) DeleteArtistProfile(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	if userID == "" {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusUnauthorized, "unauthorized: user ID not found"), fiber.StatusUnauthorized)
	}

	if err := c.artistService.DeleteArtistProfile(userID); err != nil {
		if err.Error() == "artist profile not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Artist profile deleted successfully", true)
}
