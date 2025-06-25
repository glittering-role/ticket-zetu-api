package dto

import (
	"ticket-zetu-api/modules/events/models/seats"
)

type GetSeatReservationDTO struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	EventID   string            `json:"event_id"`
	SeatID    string            `json:"seat_id"`
	Status    string            `json:"status"`
	ExpiresAt string            `json:"expires_at"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	DeletedAt *string           `json:"deleted_at,omitempty"`
	UserName  string            `json:"user_name,omitempty"`
	EventName string            `json:"event_name,omitempty"`
	SeatInfo  *seats.Seat       `json:"seat_info,omitempty"`
}

type CreateSeatReservationDTO struct {
	UserID    string `json:"user_id" binding:"required,uuid"`
	EventID   string `json:"event_id" binding:"required,uuid"`
	SeatID    string `json:"seat_id" binding:"required,uuid"`
	ExpiresAt string `json:"expires_at" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type UpdateSeatReservationDTO struct {
	ID        string `json:"id" binding:"required,uuid"`
	UserID    string `json:"user_id" binding:"required,uuid"`
	EventID   string `json:"event_id" binding:"required,uuid"`
	SeatID    string `json:"seat_id" binding:"required,uuid"`
	Status    string `json:"status" binding:"required,oneof=held confirmed released"`
	ExpiresAt string `json:"expires_at" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type ToggleSeatReservationStatusDTO struct {
	ID     string `json:"id" binding:"required,uuid"`
	Status string `json:"status" binding:"required,oneof=held confirmed released"`
}

type SeatReservationFilterDTO struct {
	UserID    string `json:"user_id,omitempty" binding:"omitempty,uuid"`
	EventID   string `json:"event_id,omitempty" binding:"omitempty,uuid"`
	SeatID    string `json:"seat_id,omitempty" binding:"omitempty,uuid"`
	Status    string `json:"status,omitempty" binding:"omitempty,oneof=held confirmed released"`
	Page      int    `json:"page,omitempty" binding:"omitempty,min=1"`
	PageSize  int    `json:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}