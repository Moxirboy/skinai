package rest

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewFrontend(
	g *gin.RouterGroup,
	log log.Logger,
) {

	g.StaticFile("/create/fact", "internal/controller/templates/fact.html")
	g.StaticFile("/create/fact/question", "internal/controller/templates/questions.html")
}
