package rest

import (
	"testDeployment/internal/delivery/dto"

	"github.com/gin-gonic/gin"
)
func(c controller) CreateDoctor(ctx *gin.Context){
	var newUser dto.User
	ctx.ShouldBindJSON(&newUser)
	id,err:=c.usecase.RegisterDoctor(&newUser)
	if err!=nil{
		if err != nil {
			c.bot.SendErrorNotification(err)
			ctx.JSON(303, gin.H{
				"message":"user is registered",
			})
			return
		}
	}
	ctx.JSON(200,gin.H{
		"id":id,
	})
}

