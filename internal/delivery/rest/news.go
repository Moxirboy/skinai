package rest

import (
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"
	"testDeployment/pkg/utils"

	"github.com/gin-gonic/gin"
)

type news struct {
	uc  usecase.INewsUseCase
	bot Bot.Bot
}

func NewNewsController(g *gin.RouterGroup, bot Bot.Bot, uc usecase.INewsUseCase) {
	controller := news{
		uc:  uc,
		bot: bot,
	}
	r := g.Group("/news")
	r.GET("/getall", controller.GetAll)
	r.GET("/getone", controller.GetOneById)
}

// GetAll godoc
// @Summary      Get all medical news
// @Description  Get medical news from PubMed & Europe PMC — topics: dermatology, AI in medicine, skincare, digital health, clinical trials
// @ID           get-all-news
// @Tags         news
// @Produce      json
// @Param        page  query  int  false  "Page number (default 1)"
// @Success      200  {object}  domain.NewsList
// @Failure      500  {object}  map[string]interface{}
// @Router       /news/getall [get]
func (cr news) GetAll(c *gin.Context) {
	pq, err := utils.GetPaginationFromCtx(c)
	if err != nil {
		cr.bot.SendErrorNotification(err)
		c.JSON(200, gin.H{"message": "Invalid pagination", "news": []interface{}{}})
		return
	}

	newsList, err := cr.uc.GetAll(c, *pq)
	if err != nil {
		cr.bot.SendErrorNotification(err)
		c.JSON(500, gin.H{"message": "Failed to fetch news"})
		return
	}

	c.JSON(200, newsList)
}

// GetOneById godoc
// @Summary      Get one medical news article
// @Description  Get a single medical article by its ID (e.g. pubmed-12345678 or epmc-12345678)
// @ID           get-one-news
// @Tags         news
// @Produce      json
// @Param        id  query  string  true  "Article ID"
// @Success      200  {object}  domain.NewWithSinglePhoto
// @Failure      404  {object}  map[string]interface{}
// @Router       /news/getone [get]
func (cr news) GetOneById(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{"message": "id parameter is required"})
		return
	}

	article, err := cr.uc.GetOneById(c, id)
	if err != nil {
		cr.bot.SendErrorNotification(err)
		c.JSON(500, gin.H{"message": "Failed to fetch article"})
		return
	}
	if article == nil {
		c.JSON(404, gin.H{"message": "Article not found"})
		return
	}

	c.JSON(200, article)
}
