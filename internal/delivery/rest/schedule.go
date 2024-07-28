package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testDeployment/internal/delivery/http"
	"testDeployment/internal/domain"
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type schedule struct{
	uc usecase.IScheduleUseCase
	bot Bot.Bot
	http request.CustomJSONRequester

}
func NewSchedule(g *gin.RouterGroup,uc usecase.IScheduleUseCase,bot Bot.Bot,request request.CustomJSONRequester){
	handler:=schedule{
		uc:uc,
		bot:bot,
		http: request,
	}
	r:=g.Group("/schedule")
	r.POST("/create",handler.Create)
	r.GET("/getall",handler.GetAll)
}
func (h *schedule) Create(c *gin.Context){
	s:=sessions.Default(c)
	req:=&domain.Schedule{}
	c.ShouldBindJSON(&req)
	if req.Agree=="1"{
		req.IsAgree=true
	}
	if s.Get("userId")==nil{
		c.JSON(http.StatusNonAuthoritativeInfo, gin.H{"error": "auth first"})
		return
	}
	req.UserId=strconv.Itoa(s.Get("userId").(int))
	if req.IsAgree{
		var NewMessage domain.NewMessage=domain.NewMessage{
			Request: "give me summary",
			UserId: req.Id,
			Name: "mirvohidkuch",
		}
		// Make the JSON request
	resp, err := h.http.Post("https://aihookah.up.railway.app/generate_response", NewMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
		return
	}
	defer resp.Body.Close()
	// Read the response body

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal(body, &responseMap)
	if err == nil {
		// Extract the value associated with the "response" key
	value, ok := responseMap["response"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format"})
		return
	}
	req.Overview=value
	}		
	}
	if req.Overview==""{
		req.Overview="no overview"}
	
	
	
	err:=h.uc.Create(c,req)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			c.JSON(200,gin.H{
				"message": "no schedules",
			})
			return
		}
		c.JSON(200,gin.H{
			"message":err,
		})
	}
	c.JSON(200,gin.H{
		"message":"success",
	})
}
func (h *schedule) GetAll(c *gin.Context){
	s:=sessions.Default(c)
userId,ok:=s.Get("userId").(int)	
if !ok{
	c.JSON(500,gin.H{
		"message":"no user id",
	})
}

res,err:=h.uc.GetAll(c,userId)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			c.JSON(200,gin.H{
				"message": "no schedules",
			})
			return
		}
		c.JSON(200,gin.H{
			"message":err,
		})
	}
	c.JSON(200,res)
}
