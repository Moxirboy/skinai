package rest

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"testDeployment/internal/domain"
	ai2 "testDeployment/pkg/ai"
)

type chat struct {
	gin   *gin.RouterGroup
	model *ai2.Dermato
}

func NewChat(
	gin *gin.RouterGroup,
	model *ai2.Dermato,
) {
	h := &chat{
		gin:   gin,
		model: model,
	}
	r := gin.Group("/chat")
	r.POST("/generate", h.SendMessage)
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
	var err error
	var NewMessage domain.NewMessage
	if err := ctx.ShouldBindJSON(&NewMessage); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if err != nil {
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}
	fmt.Println(NewMessage)
	res, err := c.model.GenerateResponse(ctx.Request.Context(), NewMessage.Request)
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

func (c controller) GetAllMessages(ctx *gin.Context) {
	s := sessions.Default(ctx)
	UserID := s.Get("userId").(int)
	messages, err := c.usecase.GetAllMessages(UserID)
	if err != nil {
		ctx.JSON(http.StatusOK, messages)
	}
	ctx.JSON(http.StatusOK, messages)
}
