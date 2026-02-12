package Bot

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ──────────────────────────────────────────────
// Build info (set via ldflags or hardcoded)
// ──────────────────────────────────────────────

var (
	Version   = "1.2.0"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
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
const healthCheckInterval = 6 * time.Hour

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

// ──────────────────────────────────────────────
// Register Commands with Telegram (setMyCommands)
// ──────────────────────────────────────────────

func (b *bot) registerCommands() {
	type botCommand struct {
		Command     string `json:"command"`
		Description string `json:"description"`
	}

	commands := []botCommand{
		{Command: "health", Description: "🏥 Check all service statuses"},
		{Command: "stats", Description: "📊 Server statistics & metrics"},
		{Command: "uptime", Description: "⏱ Server uptime info"},
		{Command: "errors", Description: "🔴 Recent error log"},
		{Command: "dbstats", Description: "🗄 Database connection pool stats"},
		{Command: "ping", Description: "🏓 Quick latency check"},
		{Command: "version", Description: "📦 Build & version info"},
		{Command: "mem", Description: "💾 Memory usage statistics"},
		{Command: "help", Description: "❓ Show available commands"},
	}

	cmdJSON, err := json.Marshal(commands)
	if err != nil {
		log.Printf("Failed to marshal commands: %v", err)
		return
	}

	params := url.Values{}
	params.Set("commands", string(cmdJSON))
	resp, err := b.MakeRequest("setMyCommands", params)
	if err != nil {
		log.Printf("Failed to register bot commands: %v", err)
		return
	}
	if !resp.Ok {
		log.Printf("setMyCommands failed: %s", resp.Description)
		return
	}
	log.Println("✅ Bot commands registered with Telegram successfully")
}

// ──────────────────────────────────────────────
// Scheduled Health Check
// ──────────────────────────────────────────────

func (b *bot) startScheduledHealthCheck() {
	go func() {
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			b.runScheduledHealthCheck()
		}
	}()
	log.Printf("Scheduled health check every %v", healthCheckInterval)
}

func (b *bot) runScheduledHealthCheck() {
	pgUp, _, pgErr := b.checkPostgres()
	botUp, _, botErr := b.checkTelegramAPI()
	aiUp, _, aiErr := b.checkGeminiAI()
	srvUp, _, srvErr := b.checkHTTPServer()

	allUp := pgUp && botUp && aiUp && srvUp
	if allUp {
		return // all good, stay silent
	}

	// Something is down — alert the monitoring chat
	var sb strings.Builder
	sb.WriteString("⚠️ *Scheduled Health Alert*\n\n")
	if !pgUp {
		sb.WriteString(fmt.Sprintf("❌ PostgreSQL: `%s`\n", pgErr))
	}
	if !botUp {
		sb.WriteString(fmt.Sprintf("❌ Telegram API: `%s`\n", botErr))
	}
	if !aiUp {
		sb.WriteString(fmt.Sprintf("❌ Gemini AI: `%s`\n", aiErr))
	}
	if !srvUp {
		sb.WriteString(fmt.Sprintf("❌ HTTP Server: `%s`\n", srvErr))
	}
	sb.WriteString(fmt.Sprintf("\n_Auto-check at %s_", time.Now().Format("2006/01/02 15:04:05")))
	b.sendReply(chatID, sb.String())
}

// StartCommandListener starts polling for Telegram bot commands in a goroutine
func (b *bot) StartCommandListener() {
	if b.BotAPI == nil {
		log.Println("Bot API is nil, skipping command listener")
		return
	}

	// Register command menu with Telegram
	b.registerCommands()

	// Start scheduled health checks
	b.startScheduledHealthCheck()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Failed to get updates channel: %v", err)
		return
	}

	go func() {
		for update := range updates {
			// Handle callback queries (inline button presses)
			if update.CallbackQuery != nil {
				b.handleCallback(update.CallbackQuery)
				continue
			}

			if update.Message == nil {
				continue
			}

			// Handle commands
			if update.Message.IsCommand() {
				b.showTyping(update.Message.Chat.ID)
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
				case "ping":
					b.handlePing(update.Message.Chat.ID, update.Message.MessageID)
				case "version":
					b.handleVersion(update.Message.Chat.ID)
				case "mem":
					b.handleMem(update.Message.Chat.ID)
				case "help", "start":
					b.handleHelp(update.Message.Chat.ID)
				default:
					b.sendReply(update.Message.Chat.ID, "❓ Unknown command. Send /help to see available commands.")
				}
				continue
			}

			// Handle regular messages — only in private chat, give a hint
			if update.Message.Chat.IsPrivate() {
				b.sendReply(update.Message.Chat.ID,
					"💡 I'm a monitoring bot. Use /help to see what I can do!")
			}
		}
	}()

	log.Println("Bot command listener started (with callbacks + scheduled checks)")
}

// ──────────────────────────────────────────────
// Callback Query Handler (inline keyboard buttons)
// ──────────────────────────────────────────────

func (b *bot) handleCallback(cq *tgbotapi.CallbackQuery) {
	// Acknowledge the callback immediately
	callback := tgbotapi.NewCallback(cq.ID, "")
	b.AnswerCallbackQuery(callback)

	targetChatID := cq.Message.Chat.ID
	b.showTyping(targetChatID)

	switch cq.Data {
	case "cb_health":
		b.handleHealth(targetChatID)
	case "cb_stats":
		b.handleStats(targetChatID)
	case "cb_uptime":
		b.handleUptime(targetChatID)
	case "cb_errors":
		b.handleErrors(targetChatID)
	case "cb_dbstats":
		b.handleDBStats(targetChatID)
	case "cb_ping":
		b.handlePing(targetChatID, 0)
	case "cb_version":
		b.handleVersion(targetChatID)
	case "cb_mem":
		b.handleMem(targetChatID)
	case "cb_help":
		b.handleHelp(targetChatID)
	}
}

// ──────────────────────────────────────────────
// Inline Keyboards
// ──────────────────────────────────────────────

func healthKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "cb_health"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Stats", "cb_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗄 DB", "cb_dbstats"),
			tgbotapi.NewInlineKeyboardButtonData("💾 Memory", "cb_mem"),
		),
	)
}

func statsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏥 Health", "cb_health"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "cb_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 Errors", "cb_errors"),
			tgbotapi.NewInlineKeyboardButtonData("⏱ Uptime", "cb_uptime"),
		),
	)
}

func quickActionsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏥 Health", "cb_health"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Stats", "cb_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🏓 Ping", "cb_ping"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏱ Uptime", "cb_uptime"),
			tgbotapi.NewInlineKeyboardButtonData("🗄 DB", "cb_dbstats"),
			tgbotapi.NewInlineKeyboardButtonData("💾 Mem", "cb_mem"),
		),
	)
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

	b.sendReplyWithKeyboard(targetChatID, sb.String(), healthKeyboard())
}

func (b *bot) handleStats(targetChatID int64) {
	b.mu.RLock()
	reqs := b.reqCount
	errs := b.errCount
	b.mu.RUnlock()

	uptime := time.Since(b.startTime).Round(time.Second)
	errRate := float64(0)
	if reqs > 0 {
		errRate = float64(errs) / float64(reqs) * 100
	}

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

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	text := fmt.Sprintf(
		"📊 *Server Statistics*\n\n"+
			"⏱ *Uptime:* `%s`\n"+
			"🌐 *Requests:* `%d`\n"+
			"❌ *Errors:* `%d` (%.1f%%)\n"+
			"🔌 *DB Connections:*\n`%s`\n"+
			"💾 *Memory:* `%.1f MB`\n"+
			"🧵 *Goroutines:* `%d`\n"+
			"🚪 *Port:* `%s`\n\n"+
			"_Updated at %s_",
		uptime, reqs, errs, errRate, dbStats,
		float64(memStats.Alloc)/1024/1024,
		runtime.NumGoroutine(),
		b.port,
		time.Now().Format("2006/01/02 15:04:05"),
	)
	b.sendReplyWithKeyboard(targetChatID, text, statsKeyboard())
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

func (b *bot) handlePing(targetChatID int64, replyToMsgID int) {
	start := time.Now()
	msg := tgbotapi.NewMessage(targetChatID, "🏓")
	msg.ParseMode = "Markdown"
	sent, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending ping: %v", err)
		return
	}
	elapsed := time.Since(start)

	edit := tgbotapi.NewEditMessageText(
		targetChatID,
		sent.MessageID,
		fmt.Sprintf("🏓 *Pong!*\n\nBot latency: `%s`", elapsed.Round(time.Millisecond)),
	)
	edit.ParseMode = "Markdown"
	b.Send(edit)
}

func (b *bot) handleVersion(targetChatID int64) {
	text := fmt.Sprintf(
		"📦 *Skin AI — Version Info*\n\n"+
			"Version: `%s`\n"+
			"Go: `%s`\n"+
			"OS/Arch: `%s/%s`\n"+
			"Build: `%s`\n"+
			"Goroutines: `%d`\n\n"+
			"_Skin AI Backend_",
		Version, GoVersion,
		runtime.GOOS, runtime.GOARCH,
		BuildTime,
		runtime.NumGoroutine(),
	)
	b.sendReply(targetChatID, text)
}

func (b *bot) handleMem(targetChatID int64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	text := fmt.Sprintf(
		"💾 *Memory Statistics*\n\n"+
			"Alloc: `%.2f MB`\n"+
			"Total Alloc: `%.2f MB`\n"+
			"Sys: `%.2f MB`\n"+
			"Heap Alloc: `%.2f MB`\n"+
			"Heap Sys: `%.2f MB`\n"+
			"Heap Objects: `%d`\n"+
			"Stack Sys: `%.2f MB`\n"+
			"GC Cycles: `%d`\n"+
			"Goroutines: `%d`\n\n"+
			"_Updated at %s_",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		float64(m.HeapAlloc)/1024/1024,
		float64(m.HeapSys)/1024/1024,
		m.HeapObjects,
		float64(m.StackSys)/1024/1024,
		m.NumGC,
		runtime.NumGoroutine(),
		time.Now().Format("2006/01/02 15:04:05"),
	)
	b.sendReply(targetChatID, text)
}

func (b *bot) handleHelp(targetChatID int64) {
	text := "🤖 *Skin AI Bot — Available Commands*\n\n" +
		"*Monitoring:*\n" +
		"/health — Check all service statuses\n" +
		"/stats — Server statistics & metrics\n" +
		"/uptime — Server uptime info\n" +
		"/errors — Recent error log\n\n" +
		"*Diagnostics:*\n" +
		"/dbstats — Database connection pool stats\n" +
		"/ping — Quick latency check\n" +
		"/mem — Memory usage statistics\n" +
		"/version — Build & version info\n\n" +
		"/help — Show this help message\n\n" +
		"💡 _Tip: Use the inline buttons below for quick navigation!_"
	b.sendReplyWithKeyboard(targetChatID, text, quickActionsKeyboard())
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

func (b *bot) showTyping(targetChatID int64) {
	action := tgbotapi.NewChatAction(targetChatID, tgbotapi.ChatTyping)
	b.Send(action)
}

func (b *bot) sendReply(targetChatID int64, text string) {
	msg := tgbotapi.NewMessage(targetChatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending bot reply: %v", err)
	}
}

func (b *bot) sendReplyWithKeyboard(targetChatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(targetChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending bot reply with keyboard: %v", err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
