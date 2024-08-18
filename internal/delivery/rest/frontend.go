package rest

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewFrontend(
	g *gin.RouterGroup,
	log log.Logger,
) {

	g.Static("/static", "internal/controller/templates")
	g.StaticFile("/", "internal/controller/templates/login.html")
	g.StaticFile("/sign", "internal/controller/templates/signUp.html")

}
