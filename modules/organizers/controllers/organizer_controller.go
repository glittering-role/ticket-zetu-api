package organizers

import (
	"math"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/organizers/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrganizerController struct {
	service    organizers_services.OrganizerService
	logHandler *handler.LogHandler
}

func NewOrganizerController(service organizers_services.OrganizerService, logHandler *handler.LogHandler) *OrganizerController {
	return &OrganizerController{
		service:    service,
		logHandler: logHandler,
	}
}

// GetOrganizer godoc
// @Summary Get organizer details
// @Description Retrieves details of a specific organizer by its ID
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Organizer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks view permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/{id} [get]
func (c *OrganizerController) GetOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid organizer ID format"), fiber.StatusBadRequest)
	}

	organizer, err := c.service.GetOrganizer(userID, id)
	if err != nil {
		if err.Error() == "invalid organizer ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks view:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, organizer, "Organizer retrieved successfully", true)

}

// GetOrganizers godoc
// @Summary List all organizations
// @Description Retrieves a paginated list of organizations with limited or full details based on user permissions
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Organizations retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers [get]
func (c *OrganizerController) GetOrganizers(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	page := ctx.QueryInt("page", 1)
	pageSize := ctx.QueryInt("page_size", 20)

	organizers, total, err := c.service.GetOrganizers(userID, page, pageSize)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return ctx.JSON(fiber.Map{
		"data": organizers,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	})
}

// GetMyOrganizer godoc
// @Summary Get user's own organizer
// @Description Retrieves the organizer associated with the authenticated user
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "My organizer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID format"
// @Failure 403 {object} map[string]interface{} "User lacks view permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/my-organization [get]
func (c *OrganizerController) GetMyOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	organizer, err := c.service.GetMyOrganizer(userID)
	if err != nil {
		if err.Error() == "invalid user ID format" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		if err.Error() == "organizer not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "user lacks view:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, organizer, "My organizer retrieved successfully", true)
}

// SearchOrganizers godoc
// @Summary Search organizers
// @Description Searches organizers by name or creator, with pagination
// @Tags Organizers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param search query string false "Search term for organizer name"
// @Param created_by query string false "Creator ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Organizers retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid creator ID format"
// @Failure 403 {object} map[string]interface{} "User lacks view permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /organizers/search [get]
func (c *OrganizerController) SearchOrganizers(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Get query parameters
	searchTerm := ctx.Query("search", "")
	createdByStr := ctx.Query("created_by", "")
	page := ctx.QueryInt("page", 1)
	pageSize := ctx.QueryInt("page_size", 10)

	// Parse createdBy if provided
	var createdBy uuid.UUID
	var err error
	if createdByStr != "" {
		createdBy, err = uuid.Parse(createdByStr)
		if err != nil {
			return c.logHandler.LogError(ctx,
				fiber.NewError(fiber.StatusBadRequest, "Invalid creator ID format"),
				fiber.StatusBadRequest)
		}
	}

	organizers, total, err := c.service.SearchOrganizers(
		userID,
		searchTerm,
		createdBy,
		page,
		pageSize,
	)
	if err != nil {
		if err.Error() == "user lacks view:organizers permission" {
			return c.logHandler.LogError(ctx,
				fiber.NewError(fiber.StatusForbidden, err.Error()),
				fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	// Return response with pagination metadata
	return c.logHandler.LogSuccess(ctx, fiber.Map{
		"organizers": organizers,
		"pagination": fiber.Map{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	}, "Organizers retrieved successfully", true)
}
