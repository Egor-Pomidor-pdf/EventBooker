package handlers

import (
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)
func AuthMiddleware() ginext.HandlerFunc {
    return func(c *ginext.Context) {
        userIDStr, err := c.Cookie("user_id")
        if err != nil {
            c.Redirect(http.StatusSeeOther, "/ui/login")
            c.Abort()
            return
        }

        userID, _ := strconv.ParseInt(userIDStr, 10, 64)
        c.Set("user_id", userID)

        c.Next()
    }
}

func (h *handler) AdminMiddleware() ginext.HandlerFunc {
    return func(c *ginext.Context) {
        userIDRaw, exists := c.Get("user_id")
        if !exists {
            c.Redirect(http.StatusSeeOther, "/ui/login")
            c.Abort()
            return
        }

        userID := userIDRaw.(int64)
        user, _ := h.service.GetUserByID(c.Request.Context(), userID)

        if !user.IsAdmin {
            c.String(http.StatusForbidden, "access denied")
            c.Abort()
            return
        }

        c.Next()
    }
}