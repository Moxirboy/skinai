package rest

import (
	"io"
	"net/http"
	config "testDeployment/internal/common/config"
	"testDeployment/internal/delivery/middleware"
	"testDeployment/internal/domain"
	"testDeployment/pkg/ai"

	"github.com/gin-gonic/gin"
)

type chat struct {
	gin    *gin.RouterGroup
	model  *ai.Dermato
	config config.Config
}

func NewChat(
	gin *gin.RouterGroup,
	model *ai.Dermato,
	config config.Config,
) {
	h := &chat{
		gin:    gin,
		model:  model,
		config: config,
	}
	r := gin.Group("/chat")
	// Apply auth detection + rate limiting for guests
	r.Use(middleware.OptionalAuth())
	r.Use(middleware.AIRateLimit())
	r.POST("/generate", h.SendMessage)
	r.POST("/upload", h.Upload)
}

// AiHandler godoc
// @Summary send message to ai
// @Description send message to ai
// @ID message
// @tags message
// @Produce json
// @Param ai body domain.NewMessage true "List of fact questions to be created"
// @Success 201 {array} domain.NewMessage
// @Router /chat/generate  [post]
func (c *chat) SendMessage(ctx *gin.Context) {
	var newMessage domain.NewMessage
	if err := ctx.ShouldBindJSON(&newMessage); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := c.model.GenerateResponse(ctx.Request.Context(), newMessage.Request)
	if err != nil {
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"response": res,
	})
}

// GetAllMessages godoc
// @Summary      Get all messages
// @Description  Returns all chat messages for the current user
// @Tags         message
// @Produce      json
// @Success      200  {array}   domain.Message
// @Router       /dashboard/middle/get-all-messages [get]
func (c controller) GetAllMessages(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	messages, err := c.usecase.GetAllMessages(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve messages"})
		return
	}
	ctx.JSON(http.StatusOK, messages)
}
// Upload godoc
// @Summary Upload an image and generate a response
// @Description This endpoint allows you to upload an image and optionally provide a prompt for AI image generation.
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image to upload"
// @Param prompt formData string false "Prompt for the image generation"
// @Success 200 {object} map[string]interface{} "response: generated image response"
// @Failure 400 {object} map[string]interface{} "error: Invalid form data or no image uploaded"
// @Failure 500 {object} map[string]interface{} "error: Could not open or read file / AI generation error"
// @Router /chat/upload [post]
func (c *chat) Upload(ctx *gin.Context) {
	
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	
	files := form.File["image"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	
	file, err := files[0].Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open file"})
		return
	}
	defer file.Close()

	
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read file"})
		return
	}

	
	prompt := ctx.PostForm("prompt")
	if prompt == "" {
		prompt = c.config.Ai.Prompt
	}



	res, err := c.model.GenerateImageResponse(ctx.Request.Context(), fileBytes, prompt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"response": res,
	})
}