package controller

import (
	"ticket-zetu-api/modules/events/events/dto"

	"github.com/gofiber/fiber/v2"
)

// ToggleFavorite godoc
// @Summary Toggle favorite status for an event
// @Description Add or remove an event from user's favorites
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Success 200 {object} map[string]interface{} "Favorite status updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/favorite [post]
func (c *EventController) ToggleFavorite(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	event_id := ctx.Params("event_id")

	if err := c.service.ToggleFavorite(userID, event_id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Favorite status updated", false)
}

// ToggleUpvote godoc
// @Summary Toggle upvote for an event
// @Description Toggle upvote status for the specified event (adds if not exists, removes if already upvoted, switches if downvoted)
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Success 200 {object} map[string]interface{} "Upvote status updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/upvote [post]
func (c *EventController) ToggleUpvote(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	event_id := ctx.Params("event_id")

	_, err := c.service.ToggleUpvote(userID, event_id)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Upvote status updated", false)
}

// ToggleDownvote godoc
// @Summary Toggle downvote for an event
// @Description Toggle downvote status for the specified event (adds if not exists, removes if already downvoted, switches if upvoted)
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Success 200 {object} map[string]interface{} "Downvote status updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/downvote [post]
func (c *EventController) ToggleDownvote(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	event_id := ctx.Params("event_id")

	_, err := c.service.ToggleDownvote(userID, event_id)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Downvote status updated", false)
}

// AddComment godoc
// @Summary Add comment to an event
// @Description Add a comment to the specified event
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Param input body dto.AddCommentInput true "Comment details"
// @Success 200 {object} map[string]interface{} "Comment added"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/comments [post]
func (c *EventController) AddComment(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("event_id")

	var input dto.AddCommentInput
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	comment, err := c.service.AddComment(userID, eventID, input.Content)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, comment, "Comment added", false)
}

// AddReply godoc
// @Summary Add reply to a comment
// @Description Add a reply to a specified comment
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event_id path string true "Event ID"
// @Param comment_id path string true "Comment ID"
// @Param input body dto.AddCommentInput true "Reply details"
// @Success 200 {object} map[string]interface{} "Reply added"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Comment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /events/{event_id}/comments/{comment_id}/replies [post]
func (c *EventController) AddReply(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	eventID := ctx.Params("event_id")
	commentID := ctx.Params("comment_id")

	var input dto.AddCommentInput
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	reply, err := c.service.AddReply(userID, eventID, commentID, input.Content)
	if err != nil {
		if err.Error() == "parent comment not found or doesn't belong to this event" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, reply, "Reply added", false)
}

// EditComment godoc
// @Summary Edit a comment or reply
// @Description Edit an existing comment or reply (within 10 minutes of creation)
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param comment_id path string true "Comment or Reply ID"
// @Param input body dto.EditCommentInput true "Updated content"
// @Success 200 {object} map[string]interface{} "Comment updated"
// @Failure 400 {object} map[string]interface{} "Invalid request or edit window expired"
// @Failure 404 {object} map[string]interface{} "Comment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /comments/{comment_id} [put]
func (c *EventController) EditComment(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	commentID := ctx.Params("comment_id")

	var input dto.EditCommentInput
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	comment, err := c.service.EditComment(userID, commentID, input.Content)
	if err != nil {
		if err.Error() == "comment not found or not owned by user" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "comment can only be edited within 10 minutes of creation" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		}
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, comment, "Comment updated", false)
}

// DeleteComment godoc
// @Summary Delete a comment or reply
// @Description Delete a comment or reply and its replies
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param comment_id path string true "Comment or Reply ID"
// @Success 200 {object} map[string]interface{} "Comment deleted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Comment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /comments/{comment_id} [delete]
func (c *EventController) DeleteComment(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	commentID := ctx.Params("comment_id")

	err := c.service.DeleteComment(userID, commentID)
	if err != nil {
		if err.Error() == "comment not found or not owned by user" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Comment deleted", false)
}

// GetUserFavorites godoc
// @Summary Get user's favorite events
// @Description Get all events favorited by the current user
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "List of favorites"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /me/favorites [get]
func (c *EventController) GetUserFavorites(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	favorites, err := c.service.GetUserFavorites(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, favorites, "Favorites retrieved", false)
}

// GetUserComments godoc
// @Summary Get user's comments
// @Description Get all comments made by the current user (excluding replies)
// @Tags Event Interactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "List of comments"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /me/comments [get]
func (c *EventController) GetUserComments(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	comments, err := c.service.GetUserComments(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusInternalServerError, err.Error()), fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, comments, "Comments retrieved", false)
}
