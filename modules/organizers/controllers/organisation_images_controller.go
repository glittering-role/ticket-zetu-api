package organizers

import (
	"ticket-zetu-api/logs/handler"
	organizers_services "ticket-zetu-api/modules/organizers/services"

	"github.com/gofiber/fiber/v2"
)

type OrganizationImageController struct {
	service    organizers_services.OrganizationImageService
	logHandler *handler.LogHandler
}

func NewOrganizationImageController(service organizers_services.OrganizationImageService, logHandler *handler.LogHandler) *OrganizationImageController {
	return &OrganizationImageController{
		service:    service,
		logHandler: logHandler,
	}
}

func (c *OrganizationImageController) AddOrganizationImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizationID := ctx.Params("organization_id")

	form, err := ctx.MultipartForm()
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid form data"), fiber.StatusBadRequest)
	}

	files := form.File["image"]
	if len(files) != 1 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Exactly one image must be provided"), fiber.StatusBadRequest)
	}

	url, err := c.service.AddImage(userID, organizationID, files[0])
	if err != nil {
		switch err.Error() {
		case "user lacks update:organizations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organization not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "file size exceeds 10MB limit", "invalid file type. Only images are allowed":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, fiber.Map{"image_url": url}, "Organization image added successfully", true)
}

func (c *OrganizationImageController) DeleteOrganizationImage(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizationID := ctx.Params("organization_id")

	err := c.service.DeleteImage(userID, organizationID)
	if err != nil {
		switch err.Error() {
		case "user lacks update:organizations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "organization not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Organization image deleted successfully", true)
}
