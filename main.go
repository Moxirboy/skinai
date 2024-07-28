package main

import (
	"log"
	configs "testDeployment/internal/common/config"
	"testDeployment/internal/server"
)

// @title Skin Ai Swagger
// @version 1.0
// @description This is a  server skin ai server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host skinai.up.railway.app
// @BasePath /api/v1


func main() {

	var (
		config = configs.Configuration()
	)

	s := server.NewServer(config)
	log.Fatal(s.Run())
}
