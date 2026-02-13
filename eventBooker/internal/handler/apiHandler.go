package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// CreateEvent — создаёт новое событие
func (h *handler) CreateEvent(c *ginext.Context) {
	ctx := c.Request.Context()

	var input struct {
		Title     string `json:"title"`
		StartTime string `json:"start_time"`
		Capacity  int64  `json:"capacity"`
	}

	if err := c.BindJSON(&input); err != nil {
		zlog.Logger.Err(err).Msg("CreateEvent: failed to bind JSON")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	t, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid start_time format, must be RFC3339"})
		return
	}

	isAdmin := c.Query("admin") == "1"

	event := &models.Event{
		Title:     input.Title,
		StartTime: t,
		Capacity:  input.Capacity,
	}

	created, err := h.service.CreateEvent(ctx, event, isAdmin)
	if err != nil {
		zlog.Logger.Err(err).Msg("CreateEvent: service failed")
		c.AbortWithStatusJSON(http.StatusForbidden, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetAllEvents — возвращает список всех событий с количеством свободных мест
func (h *handler) GetAllEvents(c *ginext.Context) {
	ctx := c.Request.Context()

	events, err := h.service.GetAllEvents(ctx)
	if err != nil {
		zlog.Logger.Err(err).Msg("GetAllEvents: failed to fetch events")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error fetching events"})
		return
	}

	type EventWithFree struct {
		*models.Event
		FreePlaces int64 `json:"free_places"`
	}

	var out []EventWithFree
	for _, e := range events {
		free, err := h.service.CountFreePlaces(ctx, e.ID)
		if err != nil {
			zlog.Logger.Err(err).Int64("event_id", e.ID).Msg("GetAllEvents: failed to count free places")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error counting free places"})
			return
		}
		out = append(out, EventWithFree{
			Event:      e,
			FreePlaces: free,
		})
	}

	c.JSON(http.StatusOK, out)
}

// GetEvent — возвращает информацию об одном событии
func (h *handler) GetEvent(c *ginext.Context) {
	ctx := c.Request.Context()
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid event id"})
		return
	}

	event, err := h.service.GetEventByID(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Msg("GetEvent: failed to fetch event")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error fetching event"})
		return
	}

	free, err := h.service.CountFreePlaces(ctx, event.ID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", event.ID).Msg("GetEvent: failed to count free places")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error counting free places"})
		return
	}

	type EventWithFree struct {
		*models.Event
		FreePlaces int64 `json:"free_places"`
	}

	c.JSON(http.StatusOK, EventWithFree{
		Event:      event,
		FreePlaces: free,
	})
}

// BookEvent — бронирование места пользователем
func (h *handler) BookEvent(c *ginext.Context) {
	ctx := c.Request.Context()
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid event id"})
		return
	}

	var input struct {
		UserID int64 `json:"user_id"`
	}

	// Поддержка JSON или query параметра
	if c.ContentType() == "application/json" {
		if err := c.BindJSON(&input); err != nil {
			zlog.Logger.Err(err).Int64("event_id", eventID).Msg("BookEvent: failed to bind JSON")
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid JSON: " + err.Error()})
			return
		}
	} else {
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "missing user id"})
			return
		}
		input.UserID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil || input.UserID <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid user id"})
			return
		}
	}

	booking := &models.Booking{
		EventID:   eventID,
		UserID:    input.UserID,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	created, err := h.service.CreateBooking(ctx, booking)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Int64("user_id", input.UserID).Msg("BookEvent: failed to create booking")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "error creating booking: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ConfirmBooking — подтверждение брони пользователем или админом
func (h *handler) ConfirmBooking(c *ginext.Context) {
	ctx := c.Request.Context()
	eventIDStr := c.Param("id")

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid event id"})
		return
	}

	var input struct {
		UserID int64 `json:"user_id"`
	}

	if c.ContentType() == "application/json" {
		if err := c.BindJSON(&input); err != nil {
			zlog.Logger.Err(err).Int64("event_id", eventID).Msg("ConfirmBooking: failed to bind JSON")
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid JSON: " + err.Error()})
			return
		}
	} else {
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "missing user id"})
			return
		}
		input.UserID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil || input.UserID <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid user id"})
			return
		}
	}

	isAdmin := c.Query("admin") == "1"

	booking, err := h.service.GetBooking(ctx, eventID, input.UserID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Int64("user_id", input.UserID).Msg("ConfirmBooking: failed to fetch booking")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "error fetching booking"})
		return
	}

	if booking.Status == "confirmed" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "booking already confirmed"})
		return
	}

	if err := h.service.UpdateBookingStatus(ctx, booking.ID, "confirmed", isAdmin, input.UserID); err != nil {
		zlog.Logger.Err(err).Int64("booking_id", booking.ID).Msg("ConfirmBooking: failed to confirm booking")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "error confirming booking"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "confirmed"})
}
