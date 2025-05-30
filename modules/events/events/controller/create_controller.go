package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (c *EventController) CreateEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input struct {
		Title         string    `json:"title" validate:"required,min=2,max=255"`
		SubcategoryID string    `json:"subcategory_id" validate:"required,uuid"`
		Description   string    `json:"description" validate:"max=1000"`
		VenueID       string    `json:"venue_id" validate:"required,uuid"`
		TotalSeats    int       `json:"total_seats" validate:"required,gt=0"`
		BasePrice     float64   `json:"base_price" validate:"required,gte=0"`
		StartTime     time.Time `json:"start_time" validate:"required"`
		EndTime       time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
		IsFeatured    bool      `json:"is_featured"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	_, err := c.service.CreateEvent(
		userID,
		input.Title,
		input.SubcategoryID,
		input.Description,
		input.VenueID,
		input.TotalSeats,
		input.BasePrice,
		input.StartTime,
		input.EndTime,
		input.IsFeatured,
	)
	if err != nil {
		switch err.Error() {
		case "user lacks create:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "organizer is not active":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Event created successfully", true)
}
