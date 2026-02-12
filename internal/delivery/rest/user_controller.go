package rest

import (
	"fmt"
	"net/http"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/delivery/middleware"
	"testDeployment/pkg/jwt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SignUp godoc
// @Summary      Register a new user
// @Description  Create a new account with email, username and password (min 6 chars). Returns JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      dto.SignupRequest  true  "Signup credentials"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      409   {object}  map[string]interface{}
// @Router       /signup [post]
func (c controller) SignUp(ctx *gin.Context) {
	var req dto.SignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Email, username (min 3 chars), and password (min 6 chars) are required",
		})
		return
	}

	newUser := dto.User{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	exist, err := c.usecase.Exist(newUser)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "check_failed",
			"message": "Could not verify username availability",
		})
		return
	}
	if exist {
		ctx.JSON(http.StatusConflict, gin.H{
			"error":   "user_exists",
			"message": "Username already registered",
		})
		return
	}

	id, err := c.usecase.RegisterUser(&newUser)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "registration_failed",
			"message": "Could not create account",
		})
		return
	}

	token, err := jwt.CreateToken(id, "user")
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_error",
			"message": "Account created but could not generate token",
		})
		return
	}

	s := sessions.Default(ctx)
	s.Set("Token", token)
	s.Set("userId", id)
	s.Save()

	ctx.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken: token,
		UserID:      id,
		Role:        "user",
		ExpiresIn:   86400,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate with username and password. Returns JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Router       /login [post]
func (c controller) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Username and password are required",
		})
		return
	}

	user := dto.User{
		Username: req.Username,
		Password: req.Password,
	}

	exist, err := c.usecase.Exist(user)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "check_failed",
			"message": "Could not verify account",
		})
		return
	}
	if !exist {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "No account found with that username",
		})
		return
	}

	match, id, err := c.usecase.Login(user)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "login_error",
			"message": "Could not process login",
		})
		return
	}

	if !match {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_credentials",
			"message": "Incorrect password",
		})
		return
	}

	token, err := jwt.CreateToken(id, "user")
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_error",
			"message": "Login successful but could not generate token",
		})
		return
	}

	s := sessions.Default(ctx)
	s.Set("Token", token)
	s.Set("userId", id)
	s.Save()

	ctx.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: token,
		UserID:      id,
		Role:        "user",
		ExpiresIn:   86400,
	})
}

// GuestLogin godoc
// @Summary      Continue as guest
// @Description  Get a temporary guest token with limited AI access (5 text + 3 image per day). No registration needed.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.GuestResponse
// @Router       /guest [post]
func (c controller) GuestLogin(ctx *gin.Context) {
	guestID := fmt.Sprintf("guest_%s_%d", ctx.ClientIP(), time.Now().UnixNano())

	token, err := jwt.CreateGuestToken(guestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_error",
			"message": "Could not create guest token",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.GuestResponse{
		AccessToken: token,
		Role:        "guest",
		AILimit:     middleware.GuestAILimit,
		UploadLimit: middleware.GuestUploadLimit,
		ExpiresIn:   7200,
		Message:     "Welcome! You have limited AI access. Register for unlimited use.",
	})
}

// AuthStatus godoc
// @Summary      Check authentication status
// @Description  Returns the current user auth status, role, and remaining quotas for guests
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/status [get]
func (c controller) AuthStatus(ctx *gin.Context) {
	role := middleware.GetRole(ctx)
	userID := middleware.GetUserID(ctx)

	response := gin.H{
		"authenticated": role != "anonymous",
		"role":          role,
	}

	if userID > 0 {
		response["user_id"] = userID
	}

	if role == "guest" {
		key := middleware.GuestKey(ctx)
		aiLeft, uploadLeft := middleware.GuestLimiter.Remaining(key)
		response["remaining_ai"] = aiLeft
		response["remaining_uploads"] = uploadLeft
		response["daily_ai_limit"] = middleware.GuestAILimit
		response["daily_upload_limit"] = middleware.GuestUploadLimit
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary      Logout user
// @Description  Clear session and logout user
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /dashboard/middle/logout [get]
func (c controller) Logout(ctx *gin.Context) {
	s := sessions.Default(ctx)
	s.Clear()
	s.Save()
	ctx.JSON(200, gin.H{
		"message": "successfully logged out",
	})
}

// DeleteAccount godoc
// @Summary      Delete user account
// @Description  Delete the current user account and clear session
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /dashboard/middle/deleteAccount [get]
func (c controller) DeleteAccount(ctx *gin.Context) {
	id := middleware.GetUserID(ctx)
	if id == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	err := c.usecase.DeleteUser(id)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := sessions.Default(ctx)
	s.Clear()
	s.Save()
	ctx.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

// GetPremium godoc
// @Summary      Get premium status
// @Description  Check if the current user has premium
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /dashboard/middle/get_premium [get]
func (c controller) GetPremium(ctx *gin.Context) {
	id := middleware.GetUserID(ctx)
	if id == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	isPremium, err := c.usecase.IsPremium(id)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"premium": isPremium})
}

// BuyPremium godoc
// @Summary      Buy premium
// @Description  Upgrade the current user to premium
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /dashboard/middle/buy_premium [get]
func (c controller) BuyPremium(ctx *gin.Context) {
	id := middleware.GetUserID(ctx)
	if id == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	err := c.usecase.UpdatePremium(id)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// GetPoint godoc
// @Summary      Get user points
// @Description  Get the current user point balance
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /dashboard/middle/get-point [get]
func (c controller) GetPoint(ctx *gin.Context) {
	id := middleware.GetUserID(ctx)
	if id == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	point, err := c.usecase.GetPoint(id)
	if err != nil {
		c.bot.SendErrorNotification(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"point": point})
}
