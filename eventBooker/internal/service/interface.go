package service

import (
	"context"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
)

var validBookingStatuses = map[string]bool{
	"booked":    true,
	"confirmed": true,
	"cancelled": true,
}

type Service interface {
	CreateEvent(ctx context.Context, event *models.Event, isAdmin bool) (*models.Event, error) // только админ
	GetEventByID(ctx context.Context, eventID int64) (*models.Event, error)
	GetAllEvents(ctx context.Context) ([]*models.Event, error)
	GetAllEventsWithFreePlaces(ctx context.Context) ([]models.EventWithFreePlaces, error)
	CountFreePlaces(ctx context.Context, eventID int64) (int64, error) // возвращает количество свободных мест

	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)

	// IsBookedByUserID(ctx context.Context, eventID, userID int64) (bool, *models.Booking, error)
	CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetBooking(ctx context.Context, eventID, userID int64) (*models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string, isAdmin bool, userID int64) error
	GetBookingsByEventID(ctx context.Context, eventID int64) ([]*models.Booking, error)

	CancelExpiredBookings(ctx context.Context) ([]*models.Booking, error) // удаляет устаревшие неоплаченные брони
}
