package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testDeployment/internal/delivery/dto"
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

// CreateUserHandler godoc
// @Summary Get all news
// @Description Get all news with pagination
// @ID get-all-news
// @tags news
// @Produce json
// @Param page query int true "Page number"
// @Success 200 {object} dto.Response
// @Router /news/getall [get]
func (cr news) GetAll(c *gin.Context){
	pq,err:=utils.GetPaginationFromCtx(c)
	if err!=nil{
		cr.bot.SendErrorNotification(err)
		c.JSON(200,gin.H{
			"message":"No news yet",
		})
	}
	
	url := fmt.Sprintf("https://api-portal.gov.uz/news/category?code_name=news&page=%d",pq.GetPage())

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Print the response body for debugging
	fmt.Println("Response body:", string(body))

	// Unmarshal the JSON data into the Response struct
	var response dto.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Print the details of each item
	for i, item := range response.Data {
		response.Data[i].UrlToWeb=fmt.Sprintf("https://gov.uz/news/view/%d/",item.ID)
	}
	
	c.JSON(200,response)
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
