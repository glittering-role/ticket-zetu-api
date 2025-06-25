package services

import (
	"errors"
	"ticket-zetu-api/modules/events/models/events"
	"ticket-zetu-api/modules/events/models/seats"
	"ticket-zetu-api/modules/events/seat_allocation/dto"
	"ticket-zetu-api/modules/tickets/models/tickets"
	"ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SeatService interface {
	GetSeats(userID string, filter dto.SeatFilterDTO) ([]dto.GetSeatDTO, error)
	CreateSeat(userID string, input dto.CreateSeatDTO) (*dto.GetSeatDTO, error)
	UpdateSeat(userID string, input dto.UpdateSeatDTO) (*dto.GetSeatDTO, error)
	DeleteSeat(userID, id string) error
	ToggleSeatStatus(userID string, input dto.ToggleSeatStatusDTO) error
}

type seatService struct {
	db          *gorm.DB
	authService authorization_service.PermissionService
}

func NewSeatService(db *gorm.DB, authService authorization_service.PermissionService) SeatService {
	return &seatService{
		db:          db,
		authService: authService,
	}
}

func (s *seatService) toDTO(seat *seats.Seat) *dto.GetSeatDTO {
	var venueName string
	var priceTier *tickets.PriceTier
	if seat.Venue.ID != "" {
		venueName = seat.Venue.Name
	}
	if seat.PriceTierID != "" {
		priceTier = &seat.PriceTier
	}

	var deletedAt *string
	if seat.DeletedAt.Valid {
		da := seat.DeletedAt.Time.Format(time.RFC3339)
		deletedAt = &da
	}

	return &dto.GetSeatDTO{
		ID:          seat.ID,
		VenueID:     seat.VenueID,
		SeatNumber:  seat.SeatNumber,
		SeatSection: seat.SeatSection,
		Status:      seat.Status,
		PriceTierID: seat.PriceTierID,
		CreatedAt:   seat.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   seat.UpdatedAt.Format(time.RFC3339),
		DeletedAt:   deletedAt,
		VenueName:   venueName,
		PriceTier:   priceTier,
	}
}

func (s *seatService) GetSeats(userID string, filter dto.SeatFilterDTO) ([]dto.GetSeatDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "read:seats")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks read:seats permission")
	}

	if filter.VenueID != "" {
		if _, err := uuid.Parse(filter.VenueID); err != nil {
			return nil, errors.New("invalid venue ID format")
		}
	}

	if filter.PriceTierID != "" {
		if _, err := uuid.Parse(filter.PriceTierID); err != nil {
			return nil, errors.New("invalid price tier ID format")
		}
	}

	// Validate pagination parameters
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	// Build query
	query := s.db.Model(&seats.Seat{}).
		Preload("Venue").
		Preload("PriceTier").
		Where("deleted_at IS NULL")

	if filter.VenueID != "" {
		query = query.Where("venue_id = ?", filter.VenueID)
		var venue events.Venue
		if err := s.db.Where("id = ? AND deleted_at IS NULL", filter.VenueID).First(&venue).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("venue not found")
			}
			return nil, err
		}
	}

	if filter.SeatNumber != "" {
		query = query.Where("seat_number = ?", filter.SeatNumber)
	}

	if filter.SeatSection != "" {
		query = query.Where("seat_section = ?", filter.SeatSection)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.PriceTierID != "" {
		query = query.Where("price_tier_id = ?", filter.PriceTierID)
		var priceTier tickets.PriceTier
		if err := s.db.Where("id = ? AND deleted_at IS NULL", filter.PriceTierID).First(&priceTier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("price tier not found")
			}
			return nil, err
		}
	}

	var seats []seats.Seat
	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&seats).Error; err != nil {
		return nil, err
	}

	if len(seats) == 0 {
		return []dto.GetSeatDTO{}, nil
	}

	var seatsDTO []dto.GetSeatDTO
	for _, seat := range seats {
		seatsDTO = append(seatsDTO, *s.toDTO(&seat))
	}
	return seatsDTO, nil
}

func (s *seatService) CreateSeat(userID string, input dto.CreateSeatDTO) (*dto.GetSeatDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "create:seats")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:seats permission")
	}

	if _, err := uuid.Parse(input.VenueID); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	if input.PriceTierID != "" {
		if _, err := uuid.Parse(input.PriceTierID); err != nil {
			return nil, errors.New("invalid price tier ID format")
		}
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.VenueID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	if input.PriceTierID != "" {
		var priceTier tickets.PriceTier
		if err := s.db.Where("id = ? AND deleted_at IS NULL", input.PriceTierID).First(&priceTier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("price tier not found")
			}
			return nil, err
		}
	}

	var existingSeat seats.Seat
	if err := s.db.Where("venue_id = ? AND seat_number = ? AND deleted_at IS NULL", input.VenueID, input.SeatNumber).First(&existingSeat).Error; err == nil {
		return nil, errors.New("seat number already exists in this venue")
	}

	seat := seats.Seat{
		VenueID:     input.VenueID,
		SeatNumber:  input.SeatNumber,
		SeatSection: input.SeatSection,
		Status:      input.Status,
		PriceTierID: input.PriceTierID,
	}

	if err := s.db.Create(&seat).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Venue").Preload("PriceTier").First(&seat, "id = ?", seat.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&seat), nil
}

func (s *seatService) UpdateSeat(userID string, input dto.UpdateSeatDTO) (*dto.GetSeatDTO, error) {
	hasPerm, err := s.authService.HasPermission(userID, "update:seats")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:seats permission")
	}

	if _, err := uuid.Parse(input.ID); err != nil {
		return nil, errors.New("invalid seat ID format")
	}

	if _, err := uuid.Parse(input.VenueID); err != nil {
		return nil, errors.New("invalid venue ID format")
	}

	if input.PriceTierID != "" {
		if _, err := uuid.Parse(input.PriceTierID); err != nil {
			return nil, errors.New("invalid price tier ID format")
		}
	}

	var seat seats.Seat
	if err := s.db.Preload("Venue").Where("id = ? AND deleted_at IS NULL", input.ID).First(&seat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seat not found")
		}
		return nil, err
	}

	var venue events.Venue
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.VenueID).First(&venue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}

	var existingSeat seats.Seat
	if err := s.db.Where("venue_id = ? AND seat_number = ? AND id != ? AND deleted_at IS NULL", input.VenueID, input.SeatNumber, input.ID).First(&existingSeat).Error; err == nil {
		return nil, errors.New("seat number already exists in this venue")
	}

	if input.PriceTierID != "" {
		var priceTier tickets.PriceTier
		if err := s.db.Where("id = ? AND deleted_at IS NULL", input.PriceTierID).First(&priceTier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("price tier not found")
			}
			return nil, err
		}
	}

	seat.VenueID = input.VenueID
	seat.SeatNumber = input.SeatNumber
	seat.SeatSection = input.SeatSection
	seat.Status = input.Status
	seat.PriceTierID = input.PriceTierID

	if err := s.db.Save(&seat).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&seat), nil
}

func (s *seatService) DeleteSeat(userID, id string) error {
	hasPerm, err := s.authService.HasPermission(userID, "delete:seats")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:seats permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid seat ID format")
	}

	var seat seats.Seat
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&seat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("seat not found")
		}
		return err
	}

	if seat.Status != "available" {
		return errors.New("cannot delete a seat that is held or booked")
	}

	if err := s.db.Model(&seat).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}

func (s *seatService) ToggleSeatStatus(userID string, input dto.ToggleSeatStatusDTO) error {
	hasPerm, err := s.authService.HasPermission(userID, "update:seats")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks update:seats permission")
	}

	if _, err := uuid.Parse(input.ID); err != nil {
		return errors.New("invalid seat ID format")
	}

	var seat seats.Seat
	if err := s.db.Where("id = ? AND deleted_at IS NULL", input.ID).First(&seat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("seat not found")
		}
		return err
	}

	if seat.Status == input.Status {
		return errors.New("seat status already set")
	}

	seat.Status = input.Status
	if err := s.db.Save(&seat).Error; err != nil {
		return err
	}

	return nil
}
