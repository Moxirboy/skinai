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
	r := g.Group("/")

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})
	r.POST("/signup", controller.SignUp)
	r.POST("/login", controller.Login)

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
		{
		}
		dash.POST("/fillUserInfo", controller.FillUserInfo)
	}
}
