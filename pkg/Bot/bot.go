package Bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type bot struct {
	*tgbotapi.BotAPI
	db        *sql.DB
	geminiKey string
	botToken  string
	port      string
	startTime time.Time
	mu        sync.RWMutex
	reqCount  int64
	errCount  int64
	lastErrs  []errorEntry
}

type errorEntry struct {
	Time    time.Time
	Message string
}

type Bot interface {
	SendErrorNotification(err error)
	SendNotification(mess string)
	StartCommandListener()
	SetDependencies(db *sql.DB, geminiKey string, port string)
	IncrementRequests()
	IncrementErrors()
}

const chatID = int64(-4103413678)
const maxRecentErrors = 10

func NewBot(botAPI *tgbotapi.BotAPI) Bot {
	return &bot{
		BotAPI:    botAPI,
		startTime: time.Now(),
		lastErrs:  make([]errorEntry, 0, maxRecentErrors),
	}
}

// SetDependencies injects runtime dependencies needed for health checks and stats
func (b *bot) SetDependencies(db *sql.DB, geminiKey string, port string) {
	b.db = db
	b.geminiKey = geminiKey
	b.port = port
	if b.BotAPI != nil {
		b.botToken = b.BotAPI.Token
	}
}

func (b *bot) IncrementRequests() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.reqCount++
}

func (b *bot) IncrementErrors() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.errCount++
}

func (b *bot) SendErrorNotification(err error) {
	if err == nil {
		return
	}
	b.mu.Lock()
	b.errCount++
	if len(b.lastErrs) >= maxRecentErrors {
		b.lastErrs = b.lastErrs[1:]
	}
	b.lastErrs = append(b.lastErrs, errorEntry{Time: time.Now(), Message: err.Error()})
	b.mu.Unlock()

	_, file, line, _ := runtime.Caller(1)
	message := fmt.Sprintf("🔴 *Error*\n`%s:%d`\n```\n%v\n```\n_%s_",
		file, line, err, time.Now().Format("2006/01/02 15:04:05"))
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	_, sendErr := b.Send(msg)
	if sendErr != nil {
		log.Printf("Error sending notification: %v", sendErr)
	}
}

func (b *bot) SendNotification(message string) {
	_, file, line, _ := runtime.Caller(1)
	logEntry := fmt.Sprintf("ℹ️ `[%s:%d]`\n%s\n_%s_",
		file, line, message, time.Now().Format("2006/01/02 15:04:05"))
	msg := tgbotapi.NewMessage(chatID, logEntry)
	msg.ParseMode = "Markdown"
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending notification: %v", err)
	}
}

// StartCommandListener starts polling for Telegram bot commands in a goroutine
func (b *bot) StartCommandListener() {
	if b.BotAPI == nil {
		log.Println("Bot API is nil, skipping command listener")
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Failed to get updates channel: %v", err)
		return
	}

	go func() {
		for update := range updates {
			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}
			switch update.Message.Command() {
			case "health":
				b.handleHealth(update.Message.Chat.ID)
			case "stats":
				b.handleStats(update.Message.Chat.ID)
			case "uptime":
				b.handleUptime(update.Message.Chat.ID)
			case "errors":
				b.handleErrors(update.Message.Chat.ID)
			case "dbstats":
				b.handleDBStats(update.Message.Chat.ID)
			case "help", "start":
				b.handleHelp(update.Message.Chat.ID)
			default:
				b.sendReply(update.Message.Chat.ID, "❓ Unknown command. Send /help to see available commands.")
			}
		}
	}()

	log.Println("Bot command listener started")
}

// ──────────────────────────────────────────────
// Command Handlers
// ──────────────────────────────────────────────

func (b *bot) handleHealth(targetChatID int64) {
	var sb strings.Builder
	sb.WriteString("🏥 *System Health Check*\n\n")
	allUp := true

	// 1. PostgreSQL
	pgUp, pgLat, pgErr := b.checkPostgres()
	if pgUp {
		sb.WriteString(fmt.Sprintf("✅ *PostgreSQL* — UP (%s)\n", pgLat))
	} else {
		sb.WriteString(fmt.Sprintf("❌ *PostgreSQL* — DOWN (%s)\n   └ `%s`\n", pgLat, pgErr))
		allUp = false
	}

	// 2. Telegram Bot API
	botUp, botLat, botErr := b.checkTelegramAPI()
	if botUp {
		sb.WriteString(fmt.Sprintf("✅ *Telegram Bot API* — UP (%s)\n", botLat))
	} else {
		sb.WriteString(fmt.Sprintf("❌ *Telegram Bot API* — DOWN (%s)\n   └ `%s`\n", botLat, botErr))
		allUp = false
	}

	// 3. Gemini AI
	aiUp, aiLat, aiErr := b.checkGeminiAI()
	if aiUp {
		sb.WriteString(fmt.Sprintf("✅ *Gemini AI* — UP (%s)\n", aiLat))
	} else {
		sb.WriteString(fmt.Sprintf("❌ *Gemini AI* — DOWN (%s)\n   └ `%s`\n", aiLat, aiErr))
		allUp = false
	}

	// 4. HTTP Server
	srvUp, srvLat, srvErr := b.checkHTTPServer()
	if srvUp {
		sb.WriteString(fmt.Sprintf("✅ *HTTP Server* — UP (%s)\n", srvLat))
	} else {
		sb.WriteString(fmt.Sprintf("❌ *HTTP Server* — DOWN (%s)\n   └ `%s`\n", srvLat, srvErr))
		allUp = false
	}

	sb.WriteString("\n")
	if allUp {
		sb.WriteString("🟢 *Overall: All systems operational*")
	} else {
		sb.WriteString("🔴 *Overall: Some systems are degraded*")
	}
	sb.WriteString(fmt.Sprintf("\n\n_Checked at %s_", time.Now().Format("2006/01/02 15:04:05")))

	b.sendReply(targetChatID, sb.String())
}

func (b *bot) handleStats(targetChatID int64) {
	b.mu.RLock()
	reqs := b.reqCount
	errs := b.errCount
	b.mu.RUnlock()

	uptime := time.Since(b.startTime).Round(time.Second)

	var dbStats string
	if b.db != nil {
		stats := b.db.Stats()
		dbStats = fmt.Sprintf(
			"   Open: %d | InUse: %d | Idle: %d",
			stats.OpenConnections, stats.InUse, stats.Idle,
		)
	} else {
		dbStats = "   N/A"
	}

	text := fmt.Sprintf(
		"📊 *Server Statistics*\n\n"+
			"⏱ *Uptime:* `%s`\n"+
			"🌐 *Requests:* `%d`\n"+
			"❌ *Errors:* `%d`\n"+
			"🔌 *DB Connections:*\n`%s`\n"+
			"🚪 *Port:* `%s`\n\n"+
			"_Updated at %s_",
		uptime, reqs, errs, dbStats, b.port,
		time.Now().Format("2006/01/02 15:04:05"),
	)
	b.sendReply(targetChatID, text)
}

func (b *bot) handleUptime(targetChatID int64) {
	uptime := time.Since(b.startTime)
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	text := fmt.Sprintf(
		"⏱ *Server Uptime*\n\n"+
			"`%dd %dh %dm %ds`\n\n"+
			"🕐 Started: `%s`\n"+
			"🕐 Now:     `%s`",
		days, hours, minutes, seconds,
		b.startTime.Format("2006/01/02 15:04:05"),
		time.Now().Format("2006/01/02 15:04:05"),
	)
	b.sendReply(targetChatID, text)
}

func (b *bot) handleErrors(targetChatID int64) {
	b.mu.RLock()
	errs := make([]errorEntry, len(b.lastErrs))
	copy(errs, b.lastErrs)
	total := b.errCount
	b.mu.RUnlock()

	if len(errs) == 0 {
		b.sendReply(targetChatID, "✅ *No errors recorded*\n\nThe system has been running without errors.")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔴 *Recent Errors* (total: %d)\n\n", total))
	for i := len(errs) - 1; i >= 0; i-- {
		e := errs[i]
		sb.WriteString(fmt.Sprintf("• `%s`\n  %s\n\n",
			e.Time.Format("15:04:05"),
			truncate(e.Message, 100),
		))
	}
	b.sendReply(targetChatID, sb.String())
}

func (b *bot) handleDBStats(targetChatID int64) {
	if b.db == nil {
		b.sendReply(targetChatID, "❌ Database connection not available")
		return
	}

	stats := b.db.Stats()
	text := fmt.Sprintf(
		"🗄 *Database Statistics*\n\n"+
			"Open Connections: `%d`\n"+
			"In Use: `%d`\n"+
			"Idle: `%d`\n"+
			"Max Open: `%d`\n"+
			"Wait Count: `%d`\n"+
			"Wait Duration: `%s`\n"+
			"Max Idle Closed: `%d`\n"+
			"Max Lifetime Closed: `%d`\n\n"+
			"_Updated at %s_",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.MaxOpenConnections,
		stats.WaitCount,
		stats.WaitDuration.Round(time.Millisecond),
		stats.MaxIdleClosed,
		stats.MaxLifetimeClosed,
		time.Now().Format("2006/01/02 15:04:05"),
	)
	b.sendReply(targetChatID, text)
}

func (b *bot) handleHelp(targetChatID int64) {
	text := "🤖 *Skin AI Bot — Available Commands*\n\n" +
		"/health — Check all service statuses\n" +
		"/stats — Server statistics & metrics\n" +
		"/uptime — Server uptime info\n" +
		"/errors — Recent error log\n" +
		"/dbstats — Database connection pool stats\n" +
		"/help — Show this help message\n\n" +
		"_Skin AI Monitoring Bot_"
	b.sendReply(targetChatID, text)
}

// ──────────────────────────────────────────────
// Service Checks
// ──────────────────────────────────────────────

func (b *bot) checkPostgres() (up bool, latency string, errMsg string) {
	if b.db == nil {
		return false, "0ms", "database not configured"
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := b.db.PingContext(ctx); err != nil {
		return false, time.Since(start).Round(time.Millisecond).String(), err.Error()
	}
	return true, time.Since(start).Round(time.Millisecond).String(), ""
}

func (b *bot) checkTelegramAPI() (up bool, latency string, errMsg string) {
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.telegram.org/bot" + b.botToken + "/getMe")
	elapsed := time.Since(start).Round(time.Millisecond).String()
	if err != nil {
		return false, elapsed, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, elapsed, fmt.Sprintf("status %d", resp.StatusCode)
	}
	return true, elapsed, ""
}

func (b *bot) checkGeminiAI() (up bool, latency string, errMsg string) {
	if b.geminiKey == "" {
		return false, "0ms", "API key not configured"
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(b.geminiKey))
	if err != nil {
		return false, time.Since(start).Round(time.Millisecond).String(), err.Error()
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash-lite")
	_, err = model.CountTokens(ctx, genai.Text("ping"))
	if err != nil {
		return false, time.Since(start).Round(time.Millisecond).String(), err.Error()
	}
	return true, time.Since(start).Round(time.Millisecond).String(), ""
}

func (b *bot) checkHTTPServer() (up bool, latency string, errMsg string) {
	if b.port == "" {
		return false, "0ms", "port not configured"
	}
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/health", b.port))
	elapsed := time.Since(start).Round(time.Millisecond).String()
	if err != nil {
		return false, elapsed, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return false, elapsed, fmt.Sprintf("status %d", resp.StatusCode)
	}
	return true, elapsed, ""
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func (b *bot) sendReply(targetChatID int64, text string) {
	msg := tgbotapi.NewMessage(targetChatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending bot reply: %v", err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
