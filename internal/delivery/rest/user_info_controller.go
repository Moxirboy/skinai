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

// CreateUserEmailHandler godoc
// @Summary User email
// @Description Update user email
// @Tags users
// @Accept  json
// @Produce  json
// @Param UserEmail body dto.UserEmail true "User email"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]string "Error response"
// @Failure 406 {object} map[string]string "Invalid request payload"
// @Router /dashboard/middle/update-email [post]
func (c controller) UpdateEmail(ctx *gin.Context) {
	var user dto.UserEmail
	s := sessions.Default(ctx)
	user.ID = s.Get("userId").(int)

	// Bind JSON input to the struct
	if err := ctx.ShouldBindJSON(&user); err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(406, gin.H{
			"Message": "Invalid request payload",
		})
		return
	}

	// Update email
	id, err := c.usecase.UpdateEmail(user)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(400, gin.H{
			"Message": "Internal error",
		})
		return
	}

	// Return the ID of the updated email
	ctx.JSON(200, gin.H{
		"id": id,
	})
}

// UpdateUserInfo godoc
// @Summary      Update user info
// @Description  Update user information fields
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_info  body  dto.UserInfo  true  "Updated user info"
// @Success      200  {string}  string  "id"
// @Failure      400  {string}  string  "internal error"
// @Failure      406  {object}  map[string]interface{}
// @Router       /dashboard/middle/updateuserinfo [post]
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
