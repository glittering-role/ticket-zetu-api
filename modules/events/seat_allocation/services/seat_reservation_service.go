package services

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/events/models/seats"
	"ticket-zetu-api/modules/events/seat_allocation/dto"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SeatReservationService interface {
	GetSeatReservations(userID string, filter dto.SeatReservationFilterDTO) ([]dto.GetSeatReservationDTO, error)
	CreateSeatReservation(userID string, input dto.CreateSeatReservationDTO) (*dto.GetSeatReservationDTO, error)
	UpdateSeatReservation(userID string, input dto.UpdateSeatReservationDTO) (*dto.GetSeatReservationDTO, error)
	DeleteSeatReservation(userID, id string) error
	ToggleSeatReservationStatus(userID string, input dto.ToggleSeatReservationStatusDTO) error
}

type seatReservationService struct {
	db          *gorm.DB
	authService authorization_service.PermissionService
}

func NewSeatReservationService(db *gorm.DB, authService authorization_service.PermissionService) SeatReservationService {
	return &seatReservationService{
		db:          db,
		authService: authService,
	}
}

func (s *seatReservationService) toDTO(reservation *seats.SeatReservation) *dto.GetSeatReservationDTO {
	var userName, eventName string
	var seatInfo *seats.Seat
	if reservation.User.ID != "" {
		userName = reservation.User.Username
	}
	if reservation.Event.ID != "" {
		eventName = reservation.Event.Title
	}
	if reservation.Seat.ID != "" {
		seatInfo = &reservation.Seat
	}

	var deletedAt *string
	if reservation.DeletedAt.Valid {
		da := reservation.DeletedAt.Time.Format(time.RFC3339)
		deletedAt = &da
	}

	return &dto.GetSeatReservationDTO{
		ID:        reservation.ID,
		UserID:    reservation.UserID,
		EventID:   reservation.EventID,
		SeatID:    reservation.SeatID,
		Status:    reservation.Status,
		ExpiresAt: reservation.ExpiresAt.Format(time.RFC3339),
		CreatedAt: reservation.CreatedAt.Format(time.RFC3339),
		UpdatedAt: reservation.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
		UserName:  userName,
		EventName: eventName,
		SeatInfo:  seatInfo,
	}
}

func (s *seatReservationService) GetSeatReservations(userID string, filter dto.SeatReservationFilterDTO) ([]dto.GetSeatReservationDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "read:seat_reservations")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:seat_reservations permission")
	}

	if filter.UserID != "" {
		if _, err := uuid.Parse(filter.UserID); err != nil {
			return nil, errors.New("invalid user ID format")
		}
	}
	if filter.EventID != "" {
		if _, err := uuid.Parse(filter.EventID); err != nil {
			return nil, errors.New("invalid event ID format")
		}
	}
	if filter.SeatID != "" {
		if _, err := uuid.Parse(filter.SeatID); err != nil {
			return nil, errors.New("invalid seat ID format")
		}
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	query := s.db.Model(&seats.SeatReservation{}).
		Preload("User").
		Preload("Event").
		Preload("Seat").
		Where("deleted_at IS NULL")

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
		var user members.User
		if err := s.db.Where("id = ? AND deleted_at IS NULL", filter.UserID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("user not found")
			}
			return nil, err
		}
	}

	if filter.EventID != "" {
		query = query.Where("event_id = ?", filter.EventID)
		var event events.Event
		if err := s.db.Where("id = ? AND deleted_at IS NULL", filter.EventID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("event not found")
			}
			return nil, err
		}
	}

	if filter.SeatID != "" {
		query = query.Where("seat_id = ?", filter.SeatID)
		var seat seats.Seat
		if err := s.db.Where("id = ? AND deleted_at IS NULL", filter.SeatID).First(&seat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("seat not found")
			}
			return nil, err
		}
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var reservations []seats.SeatReservation
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&reservations).Error; err != nil {
		return nil, err
	}

	if len(reservations) == 0 {
		return []dto.GetSeatReservationDTO{}, nil
	}

	var reservationsDTO []dto.GetSeatReservationDTO
	for _, reservation := range reservations {
		reservationsDTO = append(reservationsDTO, *s.toDTO(&reservation))
	}
	return reservationsDTO, nil
}

func (s *seatReservationService) CreateSeatReservation(userID string, input dto.CreateSeatReservationDTO) (*dto.GetSeatReservationDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "create:seat_reservations")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:seat_reservations permission")
	}

	expiresAt, err := time.Parse(time.RFC3339, input.ExpiresAt)
	if err != nil {
		return nil, errors.New("invalid expires_at format")
	}
	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expires_at must be in the future")
	}

	var user members.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	var event events.Event
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.EventID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	var seat seats.Seat
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.SeatID).First(&seat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seat not found")
		}
		return nil, err
	}
	if seat.Status != "available" {
		return nil, errors.New("seat is not available for reservation")
	}

	var existingReservation seats.SeatReservation
	if err := s.db.Where("seat_id = ? AND event_id = ? AND status IN ('held', 'confirmed') AND deleted_at IS NULL", input.SeatID, input.EventID).First(&existingReservation).Error; err == nil {
		return nil, errors.New("seat is already reserved for this event")
	}

	reservation := seats.SeatReservation{
		UserID:    input.UserID,
		EventID:   input.EventID,
		SeatID:    input.SeatID,
		Status:    "held",
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(&reservation).Error; err != nil {
		return nil, err
	}

	// Update seat status to held
	if err := s.db.Model(&seat).Update("status", "held").Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("User").Preload("Event").Preload("Seat").First(&reservation, "id = ?", reservation.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&reservation), nil
}

func (s *seatReservationService) UpdateSeatReservation(userID string, input dto.UpdateSeatReservationDTO) (*dto.GetSeatReservationDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "update:seat_reservations")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:seat_reservations permission")
	}

	expiresAt, err := time.Parse(time.RFC3339, input.ExpiresAt)
	if err != nil {
		return nil, errors.New("invalid expires_at format")
	}
	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expires_at must be in the future")
	}

	var reservation seats.SeatReservation
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.ID).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("reservation not found")
		}
		return nil, err
	}

	var user members.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	var event events.Event
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.EventID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	var seat seats.Seat
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.SeatID).First(&seat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seat not found")
		}
		return nil, err
	}

	if input.SeatID != reservation.SeatID || input.EventID != reservation.EventID {
		var existingReservation seats.SeatReservation
		if err := s.db.Where("seat_id = ? AND event_id = ? AND status IN ('held', 'confirmed') AND id != ? AND deleted_at IS NULL", input.SeatID, input.EventID, input.ID).First(&existingReservation).Error; err == nil {
			return nil, errors.New("seat is already reserved for this event")
		}
	}

	// Update old seat status if seat_id changes
	if input.SeatID != reservation.SeatID {
		var oldSeat seats.Seat
		if err := s.db.Where("id = ?", reservation.SeatID).First(&oldSeat).Error; err == nil {
			if err := s.db.Model(&oldSeat).Update("status", "available").Error; err != nil {
				return nil, err
			}
		}
	}

	reservation.UserID = input.UserID
	reservation.EventID = input.EventID
	reservation.SeatID = input.SeatID
	reservation.Status = input.Status
	reservation.ExpiresAt = expiresAt

	if err := s.db.Save(&reservation).Error; err != nil {
		return nil, err
	}

	// Update new seat status
	if err := s.db.Model(&seat).Update("status", input.Status).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("User").Preload("Event").Preload("Seat").First(&reservation, "id = ?", reservation.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&reservation), nil
}

func (s *seatReservationService) DeleteSeatReservation(userID, id string) error {
	hasPerm, err := s.authService.HasPermission(userID, "delete:seat_reservations")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:seat_reservations permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid reservation ID format")
	}

	var reservation seats.SeatReservation
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("reservation not found")
		}
		return err
	}

	if reservation.Status == "confirmed" {
		return errors.New("cannot delete a confirmed reservation")
	}

	// Update seat status to available
	var seat seats.Seat
	if err := s.db.Where("id = ?", reservation.SeatID).First(&seat).Error; err == nil {
		if err := s.db.Model(&seat).Update("status", "available").Error; err != nil {
			return err
		}
	}

	if err := s.db.Model(&reservation).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}

func (s *seatReservationService) ToggleSeatReservationStatus(userID string, input dto.ToggleSeatReservationStatusDTO) error {
	hasPerm, err := s.authService.HasPermission(userID, "update:seat_reservations")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks update:seat_reservations permission")
	}

	if _, err := uuid.Parse(input.ID); err != nil {
		return errors.New("invalid reservation ID format")
	}

	var reservation seats.SeatReservation
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.ID).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("reservation not found")
		}
		return err
	}

	if reservation.Status == input.Status {
		return errors.New("reservation status already set")
	}

	var seat seats.Seat
	if err := s.db.Where("id = ?", reservation.SeatID).First(&seat).Error; err != nil {
		return errors.New("seat not found")
	}

	reservation.Status = input.Status
	if err := s.db.Save(&reservation).Error; err != nil {
		return err
	}

	// Update seat status to match reservation status
	if err := s.db.Model(&seat).Update("status", input.Status).Error; err != nil {
		return err
	}

	return nil
}
