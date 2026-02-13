package handlers

import "github.com/wb-go/wbf/ginext"

func NewRouter(h * handler) *ginext.Engine {
	router := ginext.New("release")

	tmpl := LoadTemplates()
	router.SetHTMLTemplate(tmpl)

	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())

	router.GET("/", h.Home)

	router.GET("/ui/login", h.ShowLoginPage)
	router.POST("/ui/login", h.UILogin)

	router.GET("/ui/register", h.ShowRegisterPage)
	router.POST("/ui/register", h.RegisterUser)

    api := router.Group("/events")
    {
        api.GET("", h.GetAllEvents)
        api.GET("/:id", h.GetEvent)
        api.POST("", h.CreateEvent)
        api.POST("/:id/book", h.BookEvent)
        api.POST("/:id/confirm", h.ConfirmBooking)
    }

	    uiUser := router.Group("/ui/user")
		uiUser.Use(AuthMiddleware())
    {
        uiUser.GET("/events", h.UIUserEventsList)
        uiUser.GET("/events/:id", h.UIUserEventPage)
        uiUser.POST("/events/:id/book", h.UIUserBook)
        uiUser.POST("/events/:id/confirm", h.UIUserConfirm)
    }

	uiAdmin := router.Group("/ui/admin")
	uiAdmin.Use(AuthMiddleware(), h.AdminMiddleware())
    {
        uiAdmin.GET("/events", h.UIAdminEventsList)
        uiAdmin.GET("/events/:id", h.UIAdminEventPage)
        uiAdmin.POST("/events/:id/confirm", h.UIAdminConfirmBooking)
        uiAdmin.POST("/events/new", h.UIAdminCreateEvent)
    }

	return router
}
