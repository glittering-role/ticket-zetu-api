package category

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/category/services"

	"github.com/gofiber/fiber/v2"
)

type ImageController struct {
	service    services.ImageService
	logHandler *handler.LogHandler
}

func NewImageController(service services.ImageService, logHandler *handler.LogHandler) *ImageController {
	return &ImageController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *ImageController) AddImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	entityType := ctx.Params("entity_type") // "category" or "subcategory"
	entityID := ctx.Params("entity_id")

	form, err := ctx.MultipartForm()
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid form data"), fiber.StatusBadRequest)
	}

	files := form.File["image"]
	if len(files) != 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Exactly one image must be provided"), fiber.StatusBadRequest)
	}

	url, err := c.service.AddImage(userID, entityType, entityID, files[0])
	if err != nil {
		switch err.Error() {
		case "invalid entity type":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user lacks update:categories permission", "user lacks update:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "category not found", "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "file size exceeds 10MB limit", "invalid file type. Only images are allowed":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, fiber.Map{"image_url": url}, "Image added successfully", true)
}

func (c *ImageController) DeleteImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	entityType := ctx.Params("entity_type") // "category" or "subcategory"
	entityID := ctx.Params("entity_id")

	err := c.service.DeleteImage(userID, entityType, entityID)
	if err != nil {
		switch err.Error() {
		case "invalid entity type":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user lacks update:categories permission", "user lacks update:subcategories permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "category not found", "subcategory not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Image deleted successfully", true)
}
