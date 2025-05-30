package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (c *EventController) UpdateEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	var input struct {
		Title         string    `form:"title" validate:"required,min=2,max=255"`
		CategoryID    string    `form:"category_id" validate:"required,uuid"`
		SubcategoryID string    `form:"subcategory_id" validate:"required,uuid"`
		Description   string    `form:"description" validate:"max=1000"`
		VenueID       string    `form:"venue_id" validate:"required,uuid"`
		TotalSeats    int       `form:"total_seats" validate:"required,gt=0"`
		BasePrice     float64   `form:"base_price" validate:"required,gte=0"`
		StartTime     time.Time `form:"start_time" validate:"required"`
		EndTime       time.Time `form:"end_time" validate:"required,gtfield=StartTime"`
		PriceTierID   string    `form:"price_tier_id" validate:"required,uuid"`
		Status        string    `form:"status" validate:"required,oneof=active inactive"`
		IsFeatured    bool      `form:"is_featured"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	var imageURLs []string
	form, err := ctx.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for _, file := range files {
			if file.Size > 10*1024*1024 {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit"), fiber.StatusBadRequest)
			}
			if !isValidFileType(file.Header.Get("Content-Type")) {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid file type. Only images are allowed"), fiber.StatusBadRequest)
			}

			f, err := file.Open()
			if err != nil {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to open file"), fiber.StatusInternalServerError)
			}
			defer f.Close()

			url, err := c.cloudinary.UploadFile(ctx.Context(), f, "event_images")
			if err != nil {
				return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file to Cloudinary"), fiber.StatusInternalServerError)
			}
			imageURLs = append(imageURLs, url)
		}
	}

	event, err := c.service.UpdateEvent(
		userID,
		id,
		input.Title,
		input.Description,
		input.VenueID,
		input.CategoryID,
		input.SubcategoryID,
		input.TotalSeats,
		input.BasePrice,
		input.StartTime,
		input.EndTime,
		input.IsFeatured,
		input.Status,
		imageURLs,
	)
	if err != nil {
		switch err.Error() {
		case "userID lacks update:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "venue not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, event, "Event updated successfully", true)
}

func (c *EventController) DeleteEvent(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteEvent(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:events permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "event not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "organizer not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "cannot delete an active event":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Event deleted successfully", true)
}
