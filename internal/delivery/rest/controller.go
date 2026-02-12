package rest

import (
	request "testDeployment/internal/delivery/http"
	"testDeployment/internal/delivery/middleware"
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"

	"github.com/gin-gonic/gin"
)

type controller struct {
	usecase usecase.Usecase
	bot     Bot.Bot
	http    request.CustomJSONRequester
}

func NewController(g *gin.RouterGroup, usecase usecase.Usecase, bot Bot.Bot, request request.CustomJSONRequester) {
	controller := controller{
		usecase: usecase,
		bot:     bot,
		http:    request,
	}

	// Apply OptionalAuth to all routes so user info is available everywhere
	r := g.Group("/")
	r.Use(middleware.OptionalAuth())

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	// ── Public auth routes ──
	r.POST("/signup", controller.SignUp)
	r.POST("/login", controller.Login)
	r.POST("/guest", controller.GuestLogin)
	r.GET("/auth/status", controller.AuthStatus)
	r.GET("/auth/guest/remaining", middleware.GuestRemainingHandler())

	// ── Dashboard (protected) ──
	dash := r.Group("/dashboard")
	{
		dash.GET("/", func(c *gin.Context) {
			c.String(200, "Hello from dashboard")
		})

		middle := dash.Group("/middle")
		middle.Use(middleware.AuthMiddleware())
		{
			middle.GET("/get_premium", controller.GetPremium)
			middle.GET("/buy_premium", controller.BuyPremium)
			middle.GET("/get-all-messages", controller.GetAllMessages)
			middle.POST("/updateuserinfo", controller.UpdateUserInfo)
			middle.GET("/showUserInfo", controller.ShowUserInfo)
			middle.GET("/get-point", controller.GetPoint)
			middle.GET("/logout", controller.Logout)
			middle.GET("/deleteAccount", controller.DeleteAccount)
			middle.POST("/update-email", controller.UpdateEmail)
		}
		dash.POST("/fillUserInfo", controller.FillUserInfo)
	}
}
