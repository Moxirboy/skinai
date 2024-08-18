package rest

import (
	"github.com/gin-gonic/gin"
)

func NewFrontend(
	g *gin.RouterGroup,
) {

	g.StaticFile("/create/fact", "internal/controller/templates/fact.html")
	g.StaticFile("/create/fact/question", "internal/controller/templates/questions.html")
}
