package rest

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type HealthController struct {
	db       *sql.DB
	botToken string
	geminiKey string
}

type serviceStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type healthResponse struct {
	Status   string                   `json:"status"`
	Uptime   string                   `json:"uptime"`
	Services map[string]serviceStatus `json:"services"`
}

var startTime = time.Now()

func NewHealthController(group *gin.RouterGroup, db *sql.DB, botToken string, geminiKey string) {
	h := &HealthController{
		db:        db,
		botToken:  botToken,
		geminiKey: geminiKey,
	}

	group.GET("/health", h.HealthCheck)
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns the health status of all services (Postgres, Telegram Bot, Gemini AI)
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Failure      503  {object}  healthResponse
// @Router       /health [get]
func (h *HealthController) HealthCheck(c *gin.Context) {
	services := make(map[string]serviceStatus)
	allHealthy := true

	// Check PostgreSQL
	services["postgres"] = h.checkPostgres()
	if services["postgres"].Status != "up" {
		allHealthy = false
	}

	// Check Telegram Bot
	services["telegram_bot"] = h.checkTelegramBot()
	if services["telegram_bot"].Status != "up" {
		allHealthy = false
	}

	// Check Gemini AI
	services["gemini_ai"] = h.checkGeminiAI()
	if services["gemini_ai"].Status != "up" {
		allHealthy = false
	}

	overallStatus := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		overallStatus = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, healthResponse{
		Status:   overallStatus,
		Uptime:   time.Since(startTime).Round(time.Second).String(),
		Services: services,
	})
}

func (h *HealthController) checkPostgres() serviceStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		return serviceStatus{
			Status:  "down",
			Latency: time.Since(start).Round(time.Millisecond).String(),
			Error:   err.Error(),
		}
	}
	return serviceStatus{
		Status:  "up",
		Latency: time.Since(start).Round(time.Millisecond).String(),
	}
}

func (h *HealthController) checkTelegramBot() serviceStatus {
	start := time.Now()

	resp, err := http.Get("https://api.telegram.org/bot" + h.botToken + "/getMe")
	if err != nil {
		return serviceStatus{
			Status:  "down",
			Latency: time.Since(start).Round(time.Millisecond).String(),
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return serviceStatus{
			Status:  "down",
			Latency: time.Since(start).Round(time.Millisecond).String(),
			Error:   "telegram API returned status " + resp.Status,
		}
	}
	return serviceStatus{
		Status:  "up",
		Latency: time.Since(start).Round(time.Millisecond).String(),
	}
}

func (h *HealthController) checkGeminiAI() serviceStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(h.geminiKey))
	if err != nil {
		return serviceStatus{
			Status:  "down",
			Latency: time.Since(start).Round(time.Millisecond).String(),
			Error:   err.Error(),
		}
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash-lite")
	_, err = model.CountTokens(ctx, genai.Text("health check"))
	if err != nil {
		return serviceStatus{
			Status:  "down",
			Latency: time.Since(start).Round(time.Millisecond).String(),
			Error:   err.Error(),
		}
	}

	return serviceStatus{
		Status:  "up",
		Latency: time.Since(start).Round(time.Millisecond).String(),
	}
}
