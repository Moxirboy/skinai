package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testDeployment/internal/domain"
         
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)


func (c controller) SendMessage(ctx *gin.Context){
	s:=sessions.Default(ctx)
	var err error
	var NewMessage domain.NewMessage
	ctx.ShouldBindJSON(&NewMessage)
	NewMessage.UserId=strconv.Itoa(s.Get("userId").(int))
	NewMessage.Name,err= c.usecase.GetName(strconv.Atoi(NewMessage.UserId))
	if errors.Is(err,sql.ErrNoRows){
		if NewMessage.Name==""{
			NewMessage.Name="mirvohidKuch"
		}
	}
	if NewMessage.Name==""{
			NewMessage.Name="mirvohidKuch"
		}
	
	jsonData := NewMessage

	// Make the JSON request
	resp, err := c.http.Post("https://sololekum.up.railway.app/generate_response", jsonData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
		return
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal JSON"})
		return
	}

	// Extract the value associated with the "response" key
	responseValue, ok := responseMap["response"].(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format"})
		return
	}
	
	_,err=c.usecase.SaveMessage(NewMessage.UserId,false,NewMessage.Request)
	if err!=nil{
		
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	_,err=c.usecase.SaveMessage(NewMessage.UserId,true,responseValue)
	if err!=nil{
		
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// Send the response back to the client
	ctx.String(resp.StatusCode, string(body))
}
func (c controller) GetAllMessages(ctx *gin.Context){
	s:=sessions.Default(ctx)
	UserID:=s.Get("userId").(int)
	messages,err:=c.usecase.GetAllMessages(UserID)
	if err!=nil{
		ctx.JSON(http.StatusOK, messages)
	}
	ctx.JSON(http.StatusOK, messages)
}
