package rest

import (
	"net/http"
	"testDeployment/internal/delivery/dto"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CreateUserHandler godoc
// @Summary User info
// @Description User Info with the input attributes
// @Tags users
// @Accept  json
// @Produce  json
// @Param user_info body dto.UserInfo true "User Info"
// @Success 201 {object} dto.UserInfo
// @Router /dashboard/fillUserInfo [post]
func (c controller) FillUserInfo(ctx *gin.Context) {
	var UserInfo dto.UserInfo
	err := ctx.ShouldBindJSON(&UserInfo)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(406, gin.H{
			"Message": "Invalid credentials",
		})
		return
	}
	s := sessions.Default(ctx)
	UserInfo.Id = s.Get("userId").(int)
	id, err := c.usecase.FillInfo(UserInfo)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(400, gin.H{
			"Message": "Bad request",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(201, gin.H{
		"Message": "success",
		"Info id": id,
	})
}

// CreateUserHandler godoc
// @Summary User info
// @Description update User email
// @Tags users
// @Accept  json
// @Produce  json
// @Params user_email body dto.UserEmail true "User email"
// @Success 201 {object} dto.UserEmail
// @Router /dashboard/middle/update-email [post]
func (c controller) UpdateEmail(ctx *gin.Context) {
	var User dto.UserEmail
	s := sessions.Default(ctx)
	User.ID = s.Get("userId").(int)
	err := ctx.ShouldBindJSON(&User)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(406, gin.H{
			"Message": "Invalid credentials",
		})
		return
	}
	id, err := c.usecase.UpdateEmail(User)

	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.String(400, "internal error")
		return
	}
	ctx.String(200, "id: ", id)
}

func (c controller) UpdateUserInfo(ctx *gin.Context) {
	var User dto.UserInfo
	s := sessions.Default(ctx)
	User.Id = s.Get("userId").(int)
	err := ctx.ShouldBindJSON(&User)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(406, gin.H{
			"Message": "Invalid credentials",
		})
		return
	}

	id, err := c.usecase.UpdateInfo(User)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.String(400, "internal error")
		return
	}
	ctx.String(200, "id: ", id)
}

// CreateUserHandler godoc
// @Summary User info
// @Description Get User Info
// @Tags users
// @Accept  json
// @Produce  json
// @Success 201 {object} dto.UserInfo
// @Router /dashboard/middle/showUserInfo [get]
func (c controller) ShowUserInfo(ctx *gin.Context) {
	var User dto.UserInfo
	s := sessions.Default(ctx)
	User.Id = s.Get("userId").(int)
	if User.Id == 0 {
		ctx.String(http.StatusUnauthorized, "Not registered")
		return
	}
	User, err := c.usecase.GetUserInfo(User.Id)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(400, gin.H{
			"message": err})
		return

	}
	ctx.JSON(200, User)
}
