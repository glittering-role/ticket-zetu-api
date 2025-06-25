package controller

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/seat_allocation/dto"
	"ticket-zetu-api/modules/events/seat_allocation/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SeatReservationController struct {
	service    services.SeatReservationService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

func NewSeatReservationController(service services.SeatReservationService, logHandler *handler.LogHandler) *SeatReservationController {
	return &SeatReservationController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

// CreateSeatReservation godoc
// @Summary Create a new Seat Reservation
// @Description Creates a new seat reservation for an event.
// @Tags Seat Reservation Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateSeatReservationDTO true "Seat reservation details"
// @Success 200 {object} map[string]interface{} "Seat reservation created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks create permission"
// @Failure 404 {object} map[string]interface{} "User, event, or seat not found"
// @Failure 409 {object} map[string]interface{} "Seat already reserved"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seat-reservations [post]
func (c *SeatReservationController) CreateSeatReservation(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var input dto.CreateSeatReservationDTO
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	reservation, err := c.service.CreateSeatReservation(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks create:seat_reservations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "seat is already reserved for this event":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "invalid user ID format", "invalid event ID format", "invalid seat ID format", "invalid expires_at format", "expires_at must be in the future", "seat is not available for reservation":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user not found", "event not found", "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, reservation, "Seat reservation created successfully", true)
}

// GetSeatReservations godoc
// @Summary Get all Seat Reservations
// @Description Retrieves all seat reservations with optional filters.
// @Tags Seat Reservation Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param filter query dto.SeatReservationFilterDTO true "Seat reservation filter parameters"
// @Success 200 {object} map[string]interface{} "Seat reservations retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 403 {object} map[string]interface{} "User lacks read permission"
// @Failure 404 {object} map[string]interface{} "User, event, or seat not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seat-reservations [get]
func (c *SeatReservationController) GetSeatReservations(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	var filter dto.SeatReservationFilterDTO
	if err := ctx.QueryParser(&filter); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(filter); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	reservations, err := c.service.GetSeatReservations(userID, filter)
	if err != nil {
		switch err.Error() {
		case "user lacks read:seat_reservations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid user ID format", "invalid event ID format", "invalid seat ID format":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "user not found", "event not found", "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, reservations, "Seat reservations retrieved successfully", true)
}

// UpdateSeatReservation godoc
// @Summary Update a Seat Reservation
// @Description Updates an existing seat reservation.
// @Tags Seat Reservation Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reservation ID"
// @Param input body dto.UpdateSeatReservationDTO true "Seat reservation details"
// @Success 200 {object} map[string]interface{} "Seat reservation updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Reservation, user, event, or seat not found"
// @Failure 409 {object} map[string]interface{} "Seat already reserved"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seat-reservations/{id} [put]
func (c *SeatReservationController) UpdateSeatReservation(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	input := dto.UpdateSeatReservationDTO{
		ID: ctx.Params("id"),
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	reservation, err := c.service.UpdateSeatReservation(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks update:seat_reservations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "seat is already reserved for this event":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		case "invalid reservation ID format", "invalid user ID format", "invalid event ID format", "invalid seat ID format", "invalid expires_at format", "expires_at must be in the future":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "reservation not found", "user not found", "event not found", "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, reservation, "Seat reservation updated successfully", true)
}

// DeleteSeatReservation godoc
// @Summary Delete a Seat Reservation
// @Description Deletes a seat reservation by its ID.
// @Tags Seat Reservation Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} map[string]interface{} "Seat reservation deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid reservation ID format or reservation confirmed"
// @Failure 403 {object} map[string]interface{} "User lacks delete permission"
// @Failure 404 {object} map[string]interface{} "Reservation not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seat-reservations/{id} [delete]
func (c *SeatReservationController) DeleteSeatReservation(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	id := ctx.Params("id")

	err := c.service.DeleteSeatReservation(userID, id)
	if err != nil {
		switch err.Error() {
		case "user lacks delete:seat_reservations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid reservation ID format", "cannot delete a confirmed reservation":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "reservation not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Seat reservation deleted successfully", true)
}

// ToggleSeatReservationStatus godoc
// @Summary Toggle Seat Reservation Status
// @Description Toggles the status of a seat reservation (held, confirmed, released).
// @Tags Seat Reservation Group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Reservation ID"
// @Param input body dto.ToggleSeatReservationStatusDTO true "New status"
// @Success 200 {object} map[string]interface{} "Seat reservation status toggled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or reservation ID"
// @Failure 403 {object} map[string]interface{} "User lacks update permission"
// @Failure 404 {object} map[string]interface{} "Reservation or seat not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /seat-reservations/{id}/toggle-status [put]
func (c *SeatReservationController) ToggleSeatReservationStatus(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
	input := dto.ToggleSeatReservationStatusDTO{
		ID: ctx.Params("id"),
	}

	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
	}

	err := c.service.ToggleSeatReservationStatus(userID, input)
	if err != nil {
		switch err.Error() {
		case "user lacks update:seat_reservations permission":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		case "invalid reservation ID format", "reservation status already set":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, err.Error()), fiber.StatusBadRequest)
		case "reservation not found", "seat not found":
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		default:
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
	}
	return c.logHandler.LogSuccess(ctx, nil, "Seat reservation status toggled successfully", true)
}
