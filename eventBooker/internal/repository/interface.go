package repository

import (
	"context"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
)

type Repository interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	GetEventByID(ctx context.Context, eventID int64) (*models.Event, error)
	GetAllEvents(ctx context.Context) ([]*models.Event, error)
	GetAllEventsWithFreePlaces(ctx context.Context) ([]models.EventWithFreePlaces, error)

	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)

	CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetBooking(ctx context.Context, eventID, userID int64) (*models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error
	GetBookingsByEventID(ctx context.Context, eventID int64) ([]*models.Booking, error)
	GetBookingByID(ctx context.Context, bookingID int64) (*models.Booking, error)

	CountActiveBookings(ctx context.Context, eventID int64) (int64, error)
	FindExpiredBookings(ctx context.Context) ([]*models.Booking, error)
	CancelExpiredBookings(ctx context.Context) ([]*models.Booking, error)
}
