package handlers

import (
	"net/http"
	"strconv"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type handler struct {
	service service.Service
}

func NewHandker(s service.Service) *handler {
	return &handler{
		service: s,
	}
}

func (h *handler) Home(c *ginext.Context) {
	userIDStr, err := c.Cookie("user_id")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/ui/login")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/ui/login")
		return
	}
	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/ui/login")
		return
	}

	if user.IsAdmin {
		c.Redirect(http.StatusSeeOther, "/ui/admin/events")
	} else {
		c.Redirect(http.StatusSeeOther, "/ui/user/events")
	}
}
