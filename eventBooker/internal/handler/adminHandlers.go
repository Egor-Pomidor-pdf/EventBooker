package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// UIAdminEventsList — админская страница со списком событий и бронирований
func (h *handler) UIAdminEventsList(c *ginext.Context) {
	ctx := c.Request.Context()

	events, err := h.service.GetAllEvents(ctx)
	if err != nil {
		zlog.Logger.Err(err).Msg("UIAdminEventsList: failed to fetch events")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
			"error": "could not fetch events",
		})
		return
	}

	type EventView struct {
		ID          int64
		Title       string
		StartTime   time.Time
		Capacity    int64
		FreePlaces  int64
		Bookings    []*models.Booking
		TotalBooked int64
	}

	var data []EventView
	for _, e := range events {
		free, err := h.service.CountFreePlaces(ctx, e.ID)
		if err != nil {
			zlog.Logger.Err(err).Int64("event_id", e.ID).Msg("UIAdminEventsList: failed to count free places")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
				"error": "error counting free places",
			})
			return
		}

		bookings, err := h.service.GetBookingsByEventID(ctx, e.ID)
		if err != nil {
			zlog.Logger.Err(err).Int64("event_id", e.ID).Msg("UIAdminEventsList: failed to fetch bookings")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{
				"error": "error fetching bookings",
			})
			return
		}

		data = append(data, EventView{
			ID:          e.ID,
			Title:       e.Title,
			StartTime:   e.StartTime,
			Capacity:    e.Capacity,
			FreePlaces:  free,
			Bookings:    bookings,
			TotalBooked: int64(len(bookings)),
		})
	}

	c.HTML(http.StatusOK, "admin_events.html", data)
}

// UIAdminEventPage — админская страница одного события с деталями бронирований
func (h *handler) UIAdminEventPage(c *ginext.Context) {
	ctx := c.Request.Context()
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid event id"})
		return
	}

	event, err := h.service.GetEventByID(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Msg("UIAdminEventPage: failed to fetch event")
		c.AbortWithStatusJSON(http.StatusNotFound, ginext.H{"error": "event not found"})
		return
	}

	freePlaces, err := h.service.CountFreePlaces(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Msg("UIAdminEventPage: failed to count free places")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error counting free places"})
		return
	}

	bookings, err := h.service.GetBookingsByEventID(ctx, eventID)
	if err != nil {
		zlog.Logger.Err(err).Int64("event_id", eventID).Msg("UIAdminEventPage: failed to fetch bookings")
		c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "error fetching bookings"})
		return
	}

	data := struct {
		Event      *models.Event
		FreePlaces int64
		Bookings   []*models.Booking
	}{
		Event:      event,
		FreePlaces: freePlaces,
		Bookings:   bookings,
	}

	c.HTML(http.StatusOK, "admin_event.html", data)
}

// UIAdminConfirmBooking — админ может подтвердить любую бронь
func (h *handler) UIAdminConfirmBooking(c *ginext.Context) {
	ctx := c.Request.Context()

	bookingIDStr := c.PostForm("booking_id")
	if bookingIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "missing booking id"})
		return
	}
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "invalid booking id"})
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

	user, _ := h.service.GetUserByID(c.Request.Context(), userID)
	err = h.service.UpdateBookingStatus(ctx, bookingID, "confirmed", user.IsAdmin, userID)
	if err != nil {
		zlog.Logger.Err(err).Int64("booking_id", bookingID).Msg("UIAdminConfirmBooking: failed to confirm booking")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": "could not confirm booking"})
		return
	}

	eventIDStr := c.PostForm("event_id")
	if eventIDStr != "" {
		c.Redirect(http.StatusSeeOther, "/ui/admin/events/"+eventIDStr)
	} else {
		c.Redirect(http.StatusSeeOther, "/ui/admin/events")
	}
}

// UIAdminCreateEvent — админ создаёт новое событие
func (h *handler) UIAdminCreateEvent(c *ginext.Context) {
	var input struct {
		Title     string `form:"title" binding:"required"`
		StartTime string `form:"start_time" binding:"required"` // ожидаем формат RFC3339
		Capacity  int64  `form:"capacity" binding:"required"`
	}

	// Парсим данные из формы
	if err := c.Bind(&input); err != nil {
		zlog.Logger.Err(err).Msg("failed to bind form data")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid form data: " + err.Error(),
		})
		return
	}

	// Парсим время
	startTime, err := time.Parse("2006-01-02T15:04", input.StartTime)
	if err != nil {
		zlog.Logger.Err(err).Msg("invalid start_time format")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "invalid start_time format, must be RFC3339",
		})
		return
	}

	event := &models.Event{
		Title:     input.Title,
		StartTime: startTime,
		Capacity:  input.Capacity,
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

	user, _ := h.service.GetUserByID(c.Request.Context(), userID)

	_, err = h.service.CreateEvent(c.Request.Context(), event, user.IsAdmin)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to create event")
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{
			"error": "could not create event: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/ui/admin/events")

}
