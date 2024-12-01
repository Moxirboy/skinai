package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	_ "testDeployment/docs"
	configs "testDeployment/internal/common/config"
	"testDeployment/internal/delivery"
	request "testDeployment/internal/delivery/http"
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"
	ai2 "testDeployment/pkg/ai"
	"time"
)

type Server struct {
	cfg *configs.Config
}

func NewServer(
	cfg *configs.Config,
) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s Server) Run() error {

	r := gin.New()
	conf := configs.Configuration()
	store := cookie.NewStore([]byte(conf.Token))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   conf.Age,
		HttpOnly: true,
		Secure:   true,
	})
	r.Use(sessions.Sessions(conf.Sessions, store))
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},                                       // Set allowed origins, "*" allows all origins
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Allowed methods
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"}, // Allowed headers
		ExposeHeaders: []string{"Content-Length"},
	}))
	url := ginSwagger.URL("swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	bot, err := configs.BotConfi(conf.BotToken)
	NewBot := Bot.NewBot(bot)
	r.Use(ginLogger(NewBot))
	if err != nil {
		NewBot.SendErrorNotification(err)
		return err

	}

	httpClient := request.NewCustomHTTPClient()
	jsonRequester := request.NewCustomJSONRequester(httpClient)
	pg, err := configs.NewPostgresConfig(conf)
	if err != nil {
		NewBot.SendErrorNotification(err)
		fmt.Println(err)
		return err
	}
	uc := usecase.New(pg, NewBot)
	conf.Instruction = os.Getenv("INSTRUCTION")
	ai, err := ai2.NewDermato(os.Getenv("apikey"))
	if err != nil {
		NewBot.SendErrorNotification(err)
		fmt.Println(err)
		return err
	}
	ai.Configure(conf.Instruction, 0.7, 0.95, 40, 300)
	conf.Ai.Prompt=os.Getenv("PROMPT")
	delivery.SetUp(r, uc, NewBot, *jsonRequester, ai, *conf)
	conf.Port = os.Getenv("PORT")
	if conf.Port == "" {
		conf.Port = "8080"
	}
	NewBot.SendNotification("Runnung on : " + conf.Port)
	return r.Run(fmt.Sprintf(":%s", conf.Port))
}

func ginLogger(b Bot.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		logMessage := fmt.Sprintf("Method: %s, Path: %s, Status: %d, Duration: %v",
			c.Request.Method, c.Request.URL.Path, statusCode, duration)

		b.SendNotification(logMessage)
	}
}
