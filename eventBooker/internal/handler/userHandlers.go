package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)



func (h *handler) UIUserEventsList(c *ginext.Context) {
	events, err := h.service.GetAllEvents(c.Request.Context())
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to fetch events")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
			"error": "error fetching events ",
		})
		return
	}

	type EventView struct {
		ID         int64
		Title      string
		StartTime  time.Time
		Capacity   int64
		FreePlaces int64
	}

	var data []EventView
	for _, e := range events {
		free, err := h.service.CountFreePlaces(c.Request.Context(), e.ID)
		if err != nil {
			zlog.Logger.Err(err).Msg("failed to fetch events")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
				"error": "error counting free places:",
			})
			return
		}

		data = append(data, EventView{
			ID:         e.ID,
			Title:      e.Title,
			StartTime:  e.StartTime,
			Capacity:   e.Capacity,
			FreePlaces: free,
		})
	}

	c.HTML(http.StatusOK, "user_events.html", data)
}

func (h *handler) UIUserEventPage(c *ginext.Context) {
	ctx := c.Request.Context()

	// --- Получаем ID события из URL ---
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid event id",
		})
		return
	}

	// --- Получаем событие ---
	event, err := h.service.GetEventByID(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).
			Int64("event_id", eventID).
			Msg("UIUserEventPage: failed to fetch event")

		c.AbortWithStatusJSON(http.StatusNotFound, ginext.H{
			"error": "event not found",
		})
		return
	}

	// --- Считаем свободные места ---
	freePlaces, err := h.service.CountFreePlaces(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).
			Int64("event_id", eventID).
			Msg("UIUserEventPage: failed to count free places")

		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
			"error": "internal server error",
		})
		return
	}

	// --- Данные для шаблона ---
	data := ginext.H{
		"Event":      event,
		"FreePlaces": freePlaces,
	}

	// --- Рендер HTML ---
	c.HTML(http.StatusOK, "user_event.html", data)
}

func (h *handler) UIUserBook(c *ginext.Context) {
	ctx := c.Request.Context()

	// --- Получаем eventID из URL ---
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid event id",
		})
		return
	}

	// --- Получаем user_id  ---
	userIDStr, err := c.Cookie("user_id")
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "missing user id",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid user id",
		})
		return
	}

	// --- Создаём бронь ---
	booking := &models.Booking{
		EventID:   eventID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	_, err = h.service.CreateBooking(ctx, booking)
	if err != nil {
		zlog.Logger.Err(err).
			Int64("event_id", eventID).
			Int64("user_id", userID).
			Msg("UIUserBookEvent: failed to create booking")

		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "could not create booking",
		})
		return
	}

	// --- ВАЖНО: Redirect после POST ---
	c.Redirect(http.StatusSeeOther, "/ui/user/events/"+strconv.FormatInt(eventID, 10))
}

func (h *handler) UIUserConfirm(c *ginext.Context) {
	ctx := c.Request.Context()

	// --- Получаем eventID ---
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid event id",
		})
		return
	}

	// --- Получаем user_id ---
	userIDStr, err := c.Cookie("user_id")
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "missing user id",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid user id",
		})
		return
	}

	// --- Получаем бронь ---
	booking, err := h.service.GetBooking(ctx, eventID, userID)
	if err != nil {
		zlog.Logger.Err(err).
			Int64("event_id", eventID).
			Int64("user_id", userID).
			Msg("UIUserConfirmBooking: failed to fetch booking")

		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "could not fetch booking",
		})
		return
	}

	if booking == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "booking not found",
		})
		return
	}

	if booking.Status == "confirmed" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "booking already confirmed",
		})
		return
	}

	// --- Подтверждаем бронь ---
	err = h.service.UpdateBookingStatus(ctx, booking.ID, "confirmed", false, userID)
	if err != nil {
		zlog.Logger.Err(err).
			Int64("booking_id", booking.ID).
			Msg("UIUserConfirmBooking: failed to confirm booking")

		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "could not confirm booking",
		})
		return
	}

	// --- PRG ---
	c.Redirect(http.StatusSeeOther, "/ui/user/events/"+eventIDStr)
}
