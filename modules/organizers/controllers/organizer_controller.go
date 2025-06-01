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
	return ctx.JSON(organizer)
}

func (c *OrganizerController) GetOrganizers(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizers, err := c.service.GetOrganizers(userID)
	if err != nil {
		if err.Error() == "user lacks view:organizers permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return ctx.JSON(organizers)
}

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
	return ctx.JSON(organizer)
}

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
