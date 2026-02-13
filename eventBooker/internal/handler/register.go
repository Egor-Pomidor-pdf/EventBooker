package handlers

import (
	"fmt"
	"net/http"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/wb-go/wbf/ginext"
)

// GET /ui/register — показать форму регистрации
func (h *handler) ShowRegisterPage(c *ginext.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

// POST /ui/register — обработка регистрации
func (h *handler) RegisterUser(c *ginext.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.HTML(http.StatusBadRequest, "register.html", ginext.H{
			"error": "Имя обязательно",
		})
		return
	}

	// чекбокс для админа
	isAdmin := c.PostForm("is_admin") == "1"

	user := &models.User{
		Name:    name,
		IsAdmin: isAdmin,
	}

	createdUser, err := h.service.CreateUser(c.Request.Context(), user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", ginext.H{
			"error": fmt.Sprintf("не удалось создать пользователя: %v", err),
		})
		return
	}

	// ставим cookie с user_id
	c.SetCookie("user_id", fmt.Sprintf("%d", createdUser.ID), 3600*24*7, "/", "", false, true) // живет 7 дней

	// редирект в зависимости от роли
	if createdUser.IsAdmin {
		c.Redirect(http.StatusSeeOther, "/ui/admin/events")
	} else {
		c.Redirect(http.StatusSeeOther, "/ui/user/events")
	}
}
