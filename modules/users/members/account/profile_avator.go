package account

import (
	"ticket-zetu-api/logs/handler"
	members_service "ticket-zetu-api/modules/users/members/service"

	"github.com/gofiber/fiber/v2"
)

type ImageController struct {
	service    members_service.ImageService
	logHandler *handler.LogHandler
}

func NewImageController(service members_service.ImageService, logHandler *handler.LogHandler) *ImageController {
	return &ImageController{
		service:    service,
		logHandler: logHandler,
	}
}

// AddProfileImage godoc
// @Summary Add or update a user's profile image
// @Description Uploads a new profile image for the authenticated user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param image formData file true "Image file to upload"
// @Success 200 {object} map[string]interface{} "Profile image added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid profile image request"
// @Failure 403 {object} map[string]interface{} "User lacks permission to modify profile image"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error while adding profile image"
// @Router /users/me/image [post]
func (c *ImageController) UploadProfileImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	form, err := ctx.MultipartForm()
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid form data"), fiber.StatusBadRequest)
	}

	files := form.File["image"]
	if len(files) != 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Exactly one image must be provided"), fiber.StatusBadRequest)
	}

	file := files[0]
	if file.Size > 10*1024*1024 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit"), fiber.StatusBadRequest)
	}

	f, err := file.Open()
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, "Failed to open file"), fiber.StatusInternalServerError)
	}
	defer f.Close()

	_, err = c.service.UploadProfileImage(ctx.Context(), userID, f, file.Header.Get("Content-Type"))
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "invalid file type. Only images are allowed":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "failed to upload file to Cloudinary":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Profile image uploaded successfully", true)
}

// DeleteProfileImage godoc
// @Summary Delete a user's profile image
// @Description Removes the profile image for the authenticated user
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Profile image deleted successfully"
// @Failure 400 {object} map[string]interface{} "No profile image to delete"
// @Failure 403 {object} map[string]interface{} "User lacks permission to delete profile image"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error while deleting profile image"
// @Router /users/me/image [delete]
func (c *ImageController) DeleteProfileImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	err := c.service.DeleteProfileImage(userID)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "no profile image to delete":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "failed to delete file from Cloudinary":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Profile image deleted successfully", true)
}
