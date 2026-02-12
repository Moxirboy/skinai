package server

import (
	"fmt"
	"strings"

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
	r.SetTrustedProxies(nil) // Trust all proxies (Railway reverse proxy) for real client IP
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

		// ── Real client IP (proxy-aware) ──
		// With SetTrustedProxies(nil), Gin's ClientIP() reads X-Forwarded-For automatically.
		// We still check extra headers as fallback for CDNs like Cloudflare.
		clientIP := c.ClientIP()
		if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
			clientIP = cfIP
		}

		// ── Parse browser & OS from User-Agent ──
		userAgent := c.Request.UserAgent()
		browserName, osName := parseUserAgent(userAgent)

		// ── Additional request info ──
		referer := c.Request.Referer()
		method := c.Request.Method
		contentType := c.ContentType()
		queryStr := c.Request.URL.RawQuery
		acceptLang := c.GetHeader("Accept-Language")
		origin := c.GetHeader("Origin")
		reqSize := c.Request.ContentLength
		respSize := c.Writer.Size()

		// ── User registration status via session ──
		userInfo := "👤 Guest"
		session := sessions.Default(c)
		if userID := session.Get("userId"); userID != nil {
			userInfo = fmt.Sprintf("👤 Registered (ID: %v)", userID)
		}

		// ── Status emoji ──
		statusEmoji := "🟢"
		if statusCode >= 400 && statusCode < 500 {
			statusEmoji = "🟡"
		} else if statusCode >= 500 {
			statusEmoji = "🔴"
		}

		// ── Build the log message ──
		// Escape underscores in dynamic strings to prevent Telegram Markdown from breaking
		escPath := escapeMD(path)
		escIP := escapeMD(clientIP)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"%s *%s* `%s`\n"+
				"├ Status: `%d` | Duration: `%v`\n"+
				"├ 🌐 IP: `%s`\n"+
				"├ %s\n"+
				"├ 🖥 Browser: `%s`\n"+
				"├ 💻 OS: `%s`",
			statusEmoji, method, escPath,
			statusCode, duration.Round(time.Millisecond),
			escIP,
			userInfo,
			browserName,
			osName,
		))

		// Language
		if acceptLang != "" {
			lang := parseLanguage(acceptLang)
			sb.WriteString(fmt.Sprintf("\n├ 🌍 Language: `%s`", lang))
		}

		// Origin
		if origin != "" {
			sb.WriteString(fmt.Sprintf("\n├ 🔗 Origin: `%s`", escapeMD(origin)))
		}

		// Referer
		if referer != "" {
			sb.WriteString(fmt.Sprintf("\n├ ↩️ Referer: `%s`", escapeMD(referer)))
		}

		// Query string
		if queryStr != "" {
			sb.WriteString(fmt.Sprintf("\n├ 🔎 Query: `%s`", escapeMD(queryStr)))
		}

		// Content-Type & body sizes
		if contentType != "" && method != "GET" {
			sb.WriteString(fmt.Sprintf("\n├ 📄 Content-Type: `%s`", escapeMD(contentType)))
		}
		if reqSize > 0 {
			sb.WriteString(fmt.Sprintf("\n├ 📤 Request Size: `%s`", formatBytes(reqSize)))
		}
		if respSize > 0 {
			sb.WriteString(fmt.Sprintf("\n├ 📥 Response Size: `%s`", formatBytes(int64(respSize))))
		}

		sb.WriteString(fmt.Sprintf("\n└ 🕐 `%s`", time.Now().Format("2006/01/02 15:04:05")))

		b.SendRequestLog(sb.String())
	}
}

// parseUserAgent extracts browser and OS from the User-Agent string
func parseUserAgent(ua string) (browser string, os string) {
	browser = "Unknown"
	os = "Unknown"
	ual := strings.ToLower(ua)

	// ── OS detection ──
	switch {
	case strings.Contains(ual, "iphone"):
		os = "iOS (iPhone)"
	case strings.Contains(ual, "ipad"):
		os = "iOS (iPad)"
	case strings.Contains(ual, "android"):
		os = "Android"
	case strings.Contains(ual, "windows nt 10"):
		os = "Windows 10/11"
	case strings.Contains(ual, "windows nt"):
		os = "Windows"
	case strings.Contains(ual, "macintosh") || strings.Contains(ual, "mac os x"):
		os = "macOS"
	case strings.Contains(ual, "linux"):
		os = "Linux"
	case strings.Contains(ual, "cros"):
		os = "ChromeOS"
	case strings.Contains(ual, "bot") || strings.Contains(ual, "crawler") || strings.Contains(ual, "spider"):
		os = "Bot/Crawler"
	}

	// ── Browser detection (order matters) ──
	switch {
	case strings.Contains(ual, "edg/"):
		browser = "Edge"
	case strings.Contains(ual, "opr/") || strings.Contains(ual, "opera"):
		browser = "Opera"
	case strings.Contains(ual, "brave"):
		browser = "Brave"
	case strings.Contains(ual, "vivaldi"):
		browser = "Vivaldi"
	case strings.Contains(ual, "yabrowser"):
		browser = "Yandex"
	case strings.Contains(ual, "samsungbrowser"):
		browser = "Samsung Browser"
	case strings.Contains(ual, "ucbrowser"):
		browser = "UC Browser"
	case strings.Contains(ual, "chrome") && !strings.Contains(ual, "chromium"):
		browser = "Chrome"
	case strings.Contains(ual, "firefox"):
		browser = "Firefox"
	case strings.Contains(ual, "safari") && !strings.Contains(ual, "chrome"):
		browser = "Safari"
	case strings.Contains(ual, "msie") || strings.Contains(ual, "trident"):
		browser = "Internet Explorer"
	case strings.Contains(ual, "postman"):
		browser = "Postman"
	case strings.Contains(ual, "curl"):
		browser = "cURL"
	case strings.Contains(ual, "python"):
		browser = "Python"
	case strings.Contains(ual, "go-http-client"):
		browser = "Go HTTP Client"
	}

	if ua == "" {
		return "No UA", "Unknown"
	}
	return
}

// parseLanguage extracts the primary language from Accept-Language header
func parseLanguage(al string) string {
	// Accept-Language: en-US,en;q=0.9,uz;q=0.8  →  "en-US"
	parts := strings.SplitN(al, ",", 2)
	lang := strings.TrimSpace(parts[0])
	if idx := strings.Index(lang, ";"); idx > 0 {
		lang = lang[:idx]
	}
	if len(lang) > 10 {
		lang = lang[:10]
	}
	return lang
}

// escapeMD escapes Telegram Markdown special characters in dynamic values
func escapeMD(s string) string {
	r := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"`", "\\`",
		"[", "\\[",
	)
	return r.Replace(s)
}

// formatBytes converts bytes to human-readable string
func formatBytes(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
