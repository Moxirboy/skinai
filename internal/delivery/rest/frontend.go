package rest

import (
	"github.com/gin-gonic/gin"
)

func NewFrontend(
	g *gin.RouterGroup,
) {
	g.Static("/file", "internal/delivery/html")
	g.StaticFile("/create/fact", "internal/delivery/html/fact.html")
	g.StaticFile("/api/v1/create/fact/question", "internal/delivery/html/questions.html")

}
