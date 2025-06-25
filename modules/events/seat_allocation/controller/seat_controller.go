package controller

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/seat_allocation/dto"
	"ticket-zetu-api/modules/events/seat_allocation/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SeatController struct {
	service    services.SeatService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

func NewSeatController(service services.SeatService, logHandler *handler.LogHandler) *SeatController {
	return &SeatController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

// CreateSeat godoc
// @Summary Create a new Seat
// @Description Creates a new seat for a venue.
// @Tags Seat Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateSeatDTO true "Seat details"
// @Success 200 {object} map[string]interface{} "Seat created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "Venue or price tier not found"
// @Failure 409 {object} map[string]interface{} "Seat number already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seats [post]
func (c *SeatController) CreateSeat(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreateSeatDTO
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	if len(input.SeatNumber) > 10 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Seat number must be 10 characters or less"), fiber.StatusBadRequest)
	}

	seat, err := c.service.CreateSeat(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks create:seats permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "seat number already exists in this venue":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "invalid venue ID format", "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "venue not found", "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, seat, "Seat created successfully", true)
}

// GetSeats godoc
// @Summary Get all Seats
// @Description Retrieves all seats for a venue with optional filters.
// @Tags Seat Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param filter query dto.SeatFilterDTO true "Seat filter parameters"
// @Success 200 {object} map[string]interface{} "Seats retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "Venue or price tier not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seats [get]
func (c *SeatController) GetSeats(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var filter dto.SeatFilterDTO
	if err := ctx.QueryParser(&filter); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(filter); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	seats, err := c.service.GetSeats(userID, filter)
	if err != nil {
		switch err.Error() {
		case "user lacks read:seats permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid venue ID format", "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "venue not found", "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, seats, "Seats retrieved successfully", true)
}

// UpdateSeat godoc
// @Summary Update a Seat
// @Description Updates an existing seat.
// @Tags Seat Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Seat ID"
// @Param input body dto.UpdateSeatDTO true "Seat details"
// @Success 200 {object} map[string]interface{} "Seat updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Seat or price tier not found"
// @Failure 409 {object} map[string]interface{} "Seat number already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seats/{id} [put]
func (c *SeatController) UpdateSeat(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	input := dto.UpdateSeatDTO{
		ID: ctx.Params("id"),
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	seat, err := c.service.UpdateSeat(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks update:seats permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "seat number already exists in this venue":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "invalid seat ID format", "invalid venue ID format", "invalid price tier ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "seat not found", "venue not found", "price tier not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, seat, "Seat updated successfully", true)
}

// DeleteSeat godoc
// @Summary Delete a Seat
// @Description Deletes a seat by its ID.
// @Tags Seat Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Seat ID"
// @Success 200 {object} map[string]interface{} "Seat deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid seat ID format or seat not available"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Seat not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seats/{id} [delete]
func (c *SeatController) DeleteSeat(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteSeat(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:seats permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid seat ID format", "cannot delete a seat that is held or booked":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Seat deleted successfully", true)
}

// ToggleSeatStatus godoc
// @Summary Toggle Seat Status
// @Description Toggles the status of a seat (available, held, booked).
// @Tags Seat Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Seat ID"
// @Param input body dto.ToggleSeatStatusDTO true "New status"
// @Success 200 {object} map[string]interface{} "Seat status toggled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or seat ID"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Seat not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seats/{id}/toggle-status [put]
func (c *SeatController) ToggleSeatStatus(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	input := dto.ToggleSeatStatusDTO{
		ID: ctx.Params("id"),
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	err := c.service.ToggleSeatStatus(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks update:seats permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid seat ID format", "seat status already set":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Seat status toggled successfully", true)
}
