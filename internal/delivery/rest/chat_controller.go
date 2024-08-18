package rest

import (
	"context"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"testDeployment/internal/domain"
)

// CreateFactHandler godoc
// @Summary Create a fact question
// @Description Creates a new fact question and returns the created fact questions.
// @ID message
// @tags message
// @Produce json
// @Param fact body domain.NewMessage true "List of fact questions to be created"
// @Success 201 {array} domain.NewMessage
// @Router /dashboard/middle/send-request [post]
func (c controller) SendMessage(ctx *gin.Context) {
	var err error
	var NewMessage domain.NewMessage
	if err := ctx.ShouldBindJSON(&NewMessage); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	//body, err := generateResponse(NewMessage.Request)
	if err != nil {
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}
	body := ""
	// Send the response back to the client
	ctx.String(http.StatusOK, string(body))
}

func (c controller) GetAllMessages(ctx *gin.Context) {
	s := sessions.Default(ctx)
	UserID := s.Get("userId").(int)
	messages, err := c.usecase.GetAllMessages(UserID)
	if err != nil {
		ctx.JSON(http.StatusOK, messages)
	}
	ctx.JSON(http.StatusOK, messages)
}

// Define the API endpoint and your API key
const apiURL = "https://api.googleapis.com/v1beta2/gemini:generate"
const apiKey = "AIzaSyAdYxh1y6hn490ZFYzX3BeJRINu5XzTjh0"

// Define the request and response structures
type GenerateRequest struct {
	GenerationConfig  map[string]interface{} `json:"generation_config"`
	SafetySettings    []map[string]string    `json:"safety_settings"`
	SystemInstruction string                 `json:"system_instruction"`
}

type GenerateResponse struct {
	Text string `json:"text"`
}

func generateResponse(as string) string {
	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	key := "AIzaSyAdYxh1y6hn490ZFYzX3BeJRINu5XzTjh0"
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(as))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
	return "hi"
}
