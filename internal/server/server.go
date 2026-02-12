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
		AllowOriginFunc: func(origin string) bool {
			return true // Allow all origins
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // Cache preflight requests for 12 hours
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
	geminiKey := os.Getenv("GEMINI_API_KEY")
	ai, err := ai2.NewDermato(geminiKey)
	if err != nil {
		NewBot.SendErrorNotification(err)
		fmt.Println(err)
		return err
	}
	ai.Configure(conf.Instruction, 0.7, 0.95, 40, 300)
	conf.Ai.Prompt=os.Getenv("PROMPT")
	conf.Port = os.Getenv("PORT")
	if conf.Port == "" {
		conf.Port = "8080"
	}

	// Inject dependencies into bot for health checks and stats
	NewBot.SetDependencies(pg, geminiKey, conf.Port)
	NewBot.StartCommandListener()

	delivery.SetUp(r, uc, NewBot, *jsonRequester, ai, *conf, pg, geminiKey)
	NewBot.SendNotification("Running on : " + conf.Port)
	return r.Run(fmt.Sprintf(":%s", conf.Port))
}

func ginLogger(b Bot.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Track request and error counts for /stats command
		b.IncrementRequests()
		if statusCode >= 400 {
			b.IncrementErrors()
		}

		// Skip logging health check pings and swagger to reduce noise
		path := c.Request.URL.Path
		if path == "/api/v1/health" || len(path) >= 8 && path[:8] == "/swagger" {
			return
		}

		// Collect client info
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		method := c.Request.Method
		contentType := c.ContentType()
		queryStr := c.Request.URL.RawQuery

		// Check if user is registered via session
		userInfo := "👤 Guest"
		session := sessions.Default(c)
		if userID := session.Get("userId"); userID != nil {
			userInfo = fmt.Sprintf("👤 Registered (ID: %v)", userID)
		}

		// Build status emoji
		statusEmoji := "🟢"
		if statusCode >= 400 && statusCode < 500 {
			statusEmoji = "🟡"
		} else if statusCode >= 500 {
			statusEmoji = "🔴"
		}

		// Format log message
		logMessage := fmt.Sprintf(
			"%s *%s* `%s`\n"+
				"Status: `%d` | Duration: `%v`\n"+
				"🌐 IP: `%s`\n"+
				"%s\n"+
				"📱 UA: `%s`",
			statusEmoji, method, path,
			statusCode, duration.Round(time.Millisecond),
			clientIP,
			userInfo,
			truncateUA(userAgent, 120),
		)

		// Append optional fields only if present
		if queryStr != "" {
			logMessage += fmt.Sprintf("\n🔎 Query: `%s`", queryStr)
		}
		if referer != "" {
			logMessage += fmt.Sprintf("\n↩️ Referer: `%s`", referer)
		}
		if contentType != "" && method != "GET" {
			logMessage += fmt.Sprintf("\n📄 Content-Type: `%s`", contentType)
		}

		b.SendNotification(logMessage)
	}
}

func truncateUA(ua string, maxLen int) string {
	if len(ua) <= maxLen {
		return ua
	}
	return ua[:maxLen-3] + "..."
}
