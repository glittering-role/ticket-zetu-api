package category

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/category/dto"
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

// AddImage godoc
// @Summary Add an image to a category or subcategory
// @Description Uploads an image for a specified category or subcategory
// @Tags Category Images
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param entity_type path string true "Entity type (category or subcategory)" Enums(category, subcategory)
// @Param entity_id path string true "Entity ID (UUID)"
// @Param image formData file true "Image file to upload"
// @Success 200 {object} map[string]interface{} "Category image added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category image request"
// @Failure 403 {object} map[string]interface{} "User lacks permission to modify category image"
// @Failure 404 {object} map[string]interface{} "Category or subcategory for image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error while adding category image"
// @Router /category_images/{entity_type}/{entity_id} [post]
func (c *ImageController) AddImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	entityType := ctx.Params("entity_type")
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

	return c.logHandler.LogSuccess(ctx, dto.ImageResponse{ImageURL: url}, "Image added successfully", true)
}

// DeleteImage godoc
// @Summary Delete an image from a category or subcategory
// @Description Removes the image associated with a specified category or subcategory
// @Tags Category Images
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param entity_type path string true "Entity type (category or subcategory)" Enums(category, subcategory)
// @Param entity_id path string true "Entity ID (UUID)"
// @Success 200 {object} map[string]interface{} "Category image deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category image request"
// @Failure 403 {object} map[string]interface{} "User lacks permission to delete category image"
// @Failure 404 {object} map[string]interface{} "Category or subcategory for image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error while deleting category image"
// @Router /category_images/{entity_type}/{entity_id} [delete]
func (c *ImageController) DeleteImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	entityType := ctx.Params("entity_type")
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
