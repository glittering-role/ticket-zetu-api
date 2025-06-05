package organizers

import (
	"math"
	"ticket-zetu-api/logs/handler"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers_services "ticket-zetu-api/modules/organizers/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SubscriptionController struct {
	service    organizers_services.SubscriptionService
	logHandler *handler.LogHandler
}

func NewSubscriptionController(service organizers_services.SubscriptionService, logHandler *handler.LogHandler) *SubscriptionController {
	return &SubscriptionController{
		service:    service,
		logHandler: logHandler,
	}
}

// Subscribe godoc
// @Summary Subscribe to an organizer
// @Description Subscribes the authenticated user to an organizer
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Successfully subscribed to organizer"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 404 {object} map[string]interface{} "Organizer not found"
// @Failure 409 {object} map[string]interface{} "Already subscribed to this organizer"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions/{organizer_id}/subscribe [post]
func (c *SubscriptionController) Subscribe(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizerID := ctx.Params("organizer_id")

	if _, err := uuid.Parse(organizerID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid organizer ID format"), fiber.StatusBadRequest)
	}

	_, err := c.service.SubscribeToOrganization(userID, organizerID)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format", "invalid organizer ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "organizer not found or not accepting subscriptions":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "already subscribed to this organizer":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, nil, "Successfully subscribed to organizer", true)
}

// UnsubscribeFromOrganization godoc
// @Summary Unsubscribe from an organizer
// @Description Unsubscribes the authenticated user from an organizer
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} map[string]interface{} "Successfully unsubscribed from organizer"
// @Failure 400 {object} map[string]interface{} "Invalid organizer ID format"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 404 {object} map[string]interface{} "Subscription not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions/{organizer_id}/subscribe [delete]
func (c *SubscriptionController) UnsubscribeFromOrganization(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizerID := ctx.Params("organizer_id")

	if _, err := uuid.Parse(organizerID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid organizer ID format"), fiber.StatusBadRequest)
	}

	err := c.service.UnsubscribeFromOrganization(userID, organizerID)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format", "invalid organizer ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subscription not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Successfully unsubscribed from organizer", true)
}

// GetSubscriptionsForUser godoc
// @Summary Get user subscriptions
// @Description Retrieves a paginated list of organizers the user is subscribed to
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Subscriptions retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID format"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions [get]
func (c *SubscriptionController) GetSubscriptionsForUser(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	page := ctx.QueryInt("page", 1)
	pageSize := ctx.QueryInt("page_size", 20)

	subscriptions, total, err := c.service.GetSubscriptionsForUser(userID, page, pageSize)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, fiber.Map{
		"data": subscriptions,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	}, "Subscriptions retrieved successfully", true)
}

// GetSubscriptionsForOrganizer godoc
// @Summary Get organizer subscribers
// @Description Retrieves a paginated list of subscribers for the user's organizer
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Subscribers retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID format"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 404 {object} map[string]interface{} "No organizer found for this user"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions/my-subscribers [get]
func (c *SubscriptionController) GetSubscriptionsForOrganizer(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	page := ctx.QueryInt("page", 1)
	pageSize := ctx.QueryInt("page_size", 20)

	subscribers, total, err := c.service.GetSubscriptionsForOrganizer(userID, page, pageSize)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "no organizer found for this user":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, fiber.Map{
		"data": subscribers,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	}, "Subscribers retrieved successfully", true)
}

// UpdatePreferences godoc
// @Summary Update subscription preferences
// @Description Updates the subscription preferences for an organizer
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param organizer_id path string true "Organizer ID"
// @Param preferences body organizer_dto.SubscriptionInfo true "Subscription preferences"
// @Success 200 {object} map[string]interface{} "Subscription preferences updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or organizer ID"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 404 {object} map[string]interface{} "Subscription not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions/{organizer_id}/preferences [patch]
func (c *SubscriptionController) UpdatePreferences(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	organizerID := ctx.Params("organizer_id")

	if _, err := uuid.Parse(organizerID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid organizer ID format"), fiber.StatusBadRequest)
	}

	var prefs organizer_dto.SubscriptionInfo
	if err := ctx.BodyParser(&prefs); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	subscriptionInfo, err := c.service.UpdateSubscriptionPreferences(userID, organizerID, prefs)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format", "invalid organizer ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "subscription not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}

	return c.logHandler.LogSuccess(ctx, fiber.Map{
		"data": subscriptionInfo,
	}, "Subscription preferences updated successfully", true)
}

// BanSubscriber godoc
// @Summary Ban a subscriber
// @Description Bans a subscriber from an organizer
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param subscriber_id path string true "Subscriber ID"
// @Param reason body organizers.BanSubscriberInput true "Ban reason"
// @Success 200 {object} map[string]interface{} "Subscriber banned successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or subscriber ID"
// @Failure 403 {object} map[string]interface{} "User lacks required permission"
// @Failure 404 {object} map[string]interface{} "Subscription or organizer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscriptions/subscribers/{subscriber_id}/ban [patch]
func (c *SubscriptionController) BanSubscriber(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	subscriberID := ctx.Params("subscriber_id")

	var requestBody struct {
		Reason string `json:"reason"`
	}
	if err := ctx.BodyParser(&requestBody); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(subscriberID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid subscriber ID format"), fiber.StatusBadRequest)
	}

	err := c.service.BanSubscriber(userID, subscriberID, requestBody.Reason)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format", "invalid subscriber ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "no organizer found for this user", "subscription not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		case "user lacks required permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Subscriber banned successfully", true)
}

// BanSubscriberInput struct for OpenAPI Swagger schema
type BanSubscriberInput struct {
	Reason string `json:"reason"`
}
