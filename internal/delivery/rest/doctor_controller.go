package rest

import (
	"testDeployment/internal/delivery/dto"

	"github.com/gin-gonic/gin"
)
// CreateDoctor godoc
// @Summary      Create a doctor
// @Description  Register a new doctor user
// @Tags         doctors
// @Accept       json
// @Produce      json
// @Param        user  body  dto.User  true  "Doctor user info"
// @Success      200  {object}  map[string]interface{}
// @Failure      303  {object}  map[string]interface{}
// @Router       /doctor/create [post]
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

