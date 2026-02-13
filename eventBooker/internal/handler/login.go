package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

// GET /ui/login — форма логина (вводим user_id)
func (h *handler) ShowLoginPage(c *ginext.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// POST /ui/login — авторизация по user_id
func (h *handler) UILogin(c *ginext.Context) {
	idStr := c.PostForm("user_id")
	if idStr == "" {
		c.HTML(http.StatusBadRequest, "login.html", ginext.H{
			"error": "User ID is required",
		})
		return
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		c.HTML(http.StatusBadRequest, "login.html", ginext.H{
			"error": "Invalid User ID",
		})
		return
	}

	// Получаем пользователя из сервиса
	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", ginext.H{
			"error": "User not found",
		})
		return
	}

	// Создаём cookie
	c.SetCookie("user_id", fmt.Sprintf("%d", user.ID), 3600, "/", "", false, true)

	// Редиректим в зависимости от роли
	if user.IsAdmin {
		c.Redirect(http.StatusSeeOther, "/ui/admin/events")
	} else {
		c.Redirect(http.StatusSeeOther, "/ui/user/events")
	}
}
