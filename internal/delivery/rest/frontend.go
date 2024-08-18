package rest

import (
	"github.com/gin-gonic/gin"
)

func NewFrontend(
	g *gin.RouterGroup,
) {

	g.StaticFile("/create/fact", "internal/delivery/html/fact.html")
	g.StaticFile("/create/fact/question", "internal/delivery/html/questions.html")
}
