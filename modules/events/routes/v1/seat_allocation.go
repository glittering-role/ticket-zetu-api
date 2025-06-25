package routes

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/events/seat_allocation/controller"
	"ticket-zetu-api/modules/events/seat_allocation/services"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SeatRoutes(router fiber.Router, db *gorm.DB, logHandler *handler.LogHandler) {
	authMiddleware := middleware.IsAuthenticated(db, logHandler)
	authService := authorization_service.NewPermissionService(db)

	seatService := services.NewSeatService(db, authService)
	seatReservationService := services.NewSeatReservationService(db, authService)

	seatController := controller.NewSeatController(seatService, logHandler)
	seatReservationController := controller.NewSeatReservationController(seatReservationService, logHandler)

	seatGroup := router.Group("/seats", authMiddleware)
	{
		seatGroup.Get("/:venue_id", seatController.GetSeats)
		seatGroup.Post("/", seatController.CreateSeat)
		seatGroup.Put("/:id", seatController.UpdateSeat)
		seatGroup.Delete("/:id", seatController.DeleteSeat)
		seatGroup.Put("/:id/toggle-status", seatController.ToggleSeatStatus)
	}

	reservationGroup := router.Group("/seat-reservations", authMiddleware)
	{
		reservationGroup.Get("/", seatReservationController.GetSeatReservations)
		reservationGroup.Post("/", seatReservationController.CreateSeatReservation)
		reservationGroup.Put("/:id", seatReservationController.UpdateSeatReservation)
		reservationGroup.Delete("/:id", seatReservationController.DeleteSeatReservation)
		reservationGroup.Put("/:id/toggle-status", seatReservationController.ToggleSeatReservationStatus)
	}
}
