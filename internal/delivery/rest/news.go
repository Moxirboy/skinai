package rest

import (
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"
	"testDeployment/pkg/utils"

	"github.com/gin-gonic/gin"
)

type news struct{
	uc usecase.INewsUseCase
	bot Bot.Bot
}
func NewNewsController(g *gin.RouterGroup,bot Bot.Bot,	uc usecase.INewsUseCase){
	controller:=news{
		uc:uc,
		bot: bot,
	}
	r:=g.Group("/news")
	r.GET("/getall",controller.GetAll)
	r.GET("/getone",controller.GetOneById)
}
func (cr news) GetAll(c *gin.Context){
	pq,err:=utils.GetPaginationFromCtx(c)
	if err!=nil{
		cr.bot.SendErrorNotification(err)
		c.JSON(200,gin.H{
			"message":"No news yet",
		})
	}
	
	NewsList,err:=cr.uc.GetAll(c,*pq)
	if err!=nil{
		cr.bot.SendErrorNotification(err)
		c.JSON(200,gin.H{
			"message":"No news yet",
		})
		return
	}
	if NewsList==nil{
		c.JSON(200,gin.H{
			"message":"No news yet",
		})
		return
	}
	if NewsList.TotalPages<pq.Page{
		c.JSON(200,gin.H{
			"message":"No more pages",
		})
		return
	}
	c.JSON(200,NewsList)
}
func (cr news) GetOneById(c *gin.Context){
	id:=c.Query("id")
	news,err:=cr.uc.GetOneById(c,id)
	if err!=nil{
		cr.bot.SendErrorNotification(err)
		c.JSON(200,gin.H{
			"message":"No news yet",
		})
		return
	}
	if news==nil{
		c.JSON(200,gin.H{
			"message":"no such news",
		})
		return
	}
	c.JSON(200,news)
}