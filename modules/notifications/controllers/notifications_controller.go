package notification_controllers

import (
	"strconv"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/notifications/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type NotificationController struct {
	service    notification_service.NotificationService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

func NewNotificationController(service notification_service.NotificationService, logHandler *handler.LogHandler) *NotificationController {
	return &NotificationController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

// GetUserNotifications godoc
// @Summary Get user notifications
// @Description Retrieves notifications for the logged-in user, with optional filters for unread only or module.
// @Tags Notification Group
// @Accept application/json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User ID"
// @Param unread_only query boolean false "Filter for unread notifications only"
// @Param module query string false "Filter by module (e.g., payments, events)"
// @Param limit query int false "Limit (default: 10)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} map[string]interface{} "Notifications retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID or query parameters"
// @Failure 403 {object} map[string]interface{} "User not authorized to view notifications"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{user_id}/notifications [get]
func (c *NotificationController) GetUserNotifications(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	targetUserID := ctx.Params("user_id")

	// Ensure user can only view their own notifications
	if userID != targetUserID {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user not authorized to view notifications"), fiber.StatusForbidden)
	}

	// Parse query parameters
	unreadOnly, _ := strconv.ParseBool(ctx.Query("unread_only", "false"))
	module := ctx.Query("module")
	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get notifications
	notifications, totalCount, err := c.service.GetUserNotifications(userID, unreadOnly, module, limit, offset)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		default:
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
		}
	}

	response := map[string]interface{}{
		"notifications": notifications,
		"total_count":   totalCount,
	}

	return c.logHandler.LogSuccess(ctx, response, "Notifications retrieved successfully", true)
}
