package Bot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

// ──────────────────────────────────────────────
// Build info (set via ldflags or hardcoded)
// ──────────────────────────────────────────────

var (
	Version   = "1.3.0"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// ──────────────────────────────────────────────
// Constants
// ──────────────────────────────────────────────

const (
	chatID              = int64(-4103413678)
	maxRecentErrors     = 10
	healthCheckInterval = 6 * time.Hour
)

// ──────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────

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

// Bot is the public interface for the Telegram monitoring bot.
type Bot interface {
	SendErrorNotification(err error)
	SendNotification(mess string)
	SendRequestLog(mess string)
	StartCommandListener()
	SetDependencies(db *sql.DB, geminiKey string, port string)
	IncrementRequests()
	IncrementErrors()
}

// ──────────────────────────────────────────────
// Constructor & Dependency Injection
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Notification Methods
// ──────────────────────────────────────────────

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
	text := fmt.Sprintf("🔴 *Error*\n`%s:%d`\n```\n%v\n```\n_%s_",
		file, line, err, time.Now().Format("2006/01/02 15:04:05"))
	b.sendToChat(text)
}

func (b *bot) SendNotification(message string) {
	text := fmt.Sprintf("ℹ️ %s\n_%s_",
		message, time.Now().Format("2006/01/02 15:04:05"))
	b.sendToChat(text)
}

// SendRequestLog sends a clean request log to the monitoring chat.
// Falls back to plain text if Markdown parsing fails (e.g. special chars in paths).
func (b *bot) SendRequestLog(message string) {
	if b.BotAPI == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	_, err := b.Send(msg)
	if err != nil {
		// Markdown failed — retry without formatting so the message still arrives
		msg.ParseMode = ""
		if _, retryErr := b.Send(msg); retryErr != nil {
			log.Printf("Error sending request log: %v (markdown err: %v)", retryErr, err)
		}
	}
}

// sendToChat is a shared helper for sending Markdown messages to the monitoring chat.
func (b *bot) sendToChat(text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.Send(msg); err != nil {
		log.Printf("Error sending bot message: %v", err)
	}
}

// ──────────────────────────────────────────────
// Register Commands with Telegram (setMyCommands)
// ──────────────────────────────────────────────

func (b *bot) registerCommands() {
	if b.botToken == "" {
		log.Println("Bot token is empty, skipping command registration")
		return
	}

	type botCommand struct {
		Command     string `json:"command"`
		Description string `json:"description"`
	}
	type setCommandsPayload struct {
		Commands []botCommand   `json:"commands"`
		Scope    map[string]string `json:"scope,omitempty"`
	}

	commands := []botCommand{
		{Command: "health", Description: "Check all service statuses"},
		{Command: "stats", Description: "Server statistics and metrics"},
		{Command: "uptime", Description: "Server uptime info"},
		{Command: "errors", Description: "Recent error log"},
		{Command: "dbstats", Description: "Database connection pool stats"},
		{Command: "ping", Description: "Quick latency check"},
		{Command: "version", Description: "Build and version info"},
		{Command: "mem", Description: "Memory usage statistics"},
		{Command: "help", Description: "Show available commands"},
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", b.botToken)

	// Register for each scope via direct HTTP POST with JSON body
	scopes := []map[string]string{
		nil,                           // default (private chats) — no scope field
		{"type": "all_group_chats"},   // groups & supergroups
		{"type": "all_chat_administrators"}, // admins
	}

	for _, scope := range scopes {
		payload := setCommandsPayload{Commands: commands, Scope: scope}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal setMyCommands payload: %v", err)
			continue
		}

		resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
		if err != nil {
			scopeName := "default"
			if scope != nil {
				scopeName = scope["type"]
			}
			log.Printf("Failed to register commands (scope: %s): %v", scopeName, err)
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		scopeName := "default"
		if scope != nil {
			scopeName = scope["type"]
		}
		log.Printf("setMyCommands (scope: %s) → %s", scopeName, string(respBody))
	}
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
	log.Printf("⏰ Scheduled health check every %v", healthCheckInterval)
}

func (b *bot) runScheduledHealthCheck() {
	checks := []struct {
		name  string
		check func() (bool, string, string)
	}{
		{"PostgreSQL", b.checkPostgres},
		{"Telegram API", b.checkTelegramAPI},
		{"Gemini AI", b.checkGeminiAI},
		{"HTTP Server", b.checkHTTPServer},
	}

	var down []string
	for _, c := range checks {
		up, _, errMsg := c.check()
		if !up {
			down = append(down, fmt.Sprintf("❌ %s: `%s`", c.name, errMsg))
		}
	}

	if len(down) == 0 {
		return // all good, stay silent
	}

	var sb strings.Builder
	sb.WriteString("⚠️ *Scheduled Health Alert*\n\n")
	for _, d := range down {
		sb.WriteString(d + "\n")
	}
	sb.WriteString(fmt.Sprintf("\n_Auto-check at %s_", time.Now().Format("2006/01/02 15:04:05")))
	b.sendReply(chatID, sb.String())
}

// ──────────────────────────────────────────────
// Command Listener & Dispatch
// ──────────────────────────────────────────────

// commandHandler maps command names to their handler functions.
type commandHandler func(chatID int64)

func (b *bot) commandHandlers() map[string]commandHandler {
	return map[string]commandHandler{
		"health":  b.handleHealth,
		"stats":   b.handleStats,
		"uptime":  b.handleUptime,
		"errors":  b.handleErrors,
		"dbstats": b.handleDBStats,
		"version": b.handleVersion,
		"mem":     b.handleMem,
		"help":    b.handleHelp,
		"start":   b.handleHelp,
	}
}

// StartCommandListener starts polling for Telegram bot commands in a goroutine.
func (b *bot) StartCommandListener() {
	if b.BotAPI == nil {
		log.Println("Bot API is nil, skipping command listener")
		return
	}

	b.registerCommands()
	b.startScheduledHealthCheck()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Failed to get updates channel: %v", err)
		return
	}

	handlers := b.commandHandlers()

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

			if update.Message.IsCommand() {
				b.showTyping(update.Message.Chat.ID)
				cmd := update.Message.Command()

				// Special case: /ping needs messageID for edit trick
				if cmd == "ping" {
					b.handlePing(update.Message.Chat.ID, update.Message.MessageID)
					continue
				}

				if handler, ok := handlers[cmd]; ok {
					handler(update.Message.Chat.ID)
				} else {
					b.sendReply(update.Message.Chat.ID,
						"❓ Unknown command. Send /help to see available commands.")
				}
				continue
			}

			// Regular messages — hint in private chat only
			if update.Message.Chat.IsPrivate() {
				b.sendReply(update.Message.Chat.ID,
					"💡 I'm a monitoring bot. Use /help to see what I can do!")
			}
		}
	}()

	log.Println("🤖 Bot command listener started")
}

// ──────────────────────────────────────────────
// Callback Query Handler (inline keyboard buttons)
// ──────────────────────────────────────────────

var callbackMap = map[string]string{
	"cb_health":  "health",
	"cb_stats":   "stats",
	"cb_uptime":  "uptime",
	"cb_errors":  "errors",
	"cb_dbstats": "dbstats",
	"cb_ping":    "ping",
	"cb_version": "version",
	"cb_mem":     "mem",
	"cb_help":    "help",
}

func (b *bot) handleCallback(cq *tgbotapi.CallbackQuery) {
	// Acknowledge immediately
	b.AnswerCallbackQuery(tgbotapi.NewCallback(cq.ID, ""))

	targetChatID := cq.Message.Chat.ID
	b.showTyping(targetChatID)

	cmd, ok := callbackMap[cq.Data]
	if !ok {
		return
	}

	if cmd == "ping" {
		b.handlePing(targetChatID, 0)
		return
	}

	handlers := b.commandHandlers()
	if handler, ok := handlers[cmd]; ok {
		handler(targetChatID)
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
	type checkResult struct {
		name    string
		up      bool
		latency string
		errMsg  string
	}

	checks := []checkResult{
		{"PostgreSQL", false, "", ""},
		{"Telegram Bot API", false, "", ""},
		{"Gemini AI", false, "", ""},
		{"HTTP Server", false, "", ""},
	}
	checks[0].up, checks[0].latency, checks[0].errMsg = b.checkPostgres()
	checks[1].up, checks[1].latency, checks[1].errMsg = b.checkTelegramAPI()
	checks[2].up, checks[2].latency, checks[2].errMsg = b.checkGeminiAI()
	checks[3].up, checks[3].latency, checks[3].errMsg = b.checkHTTPServer()

	var sb strings.Builder
	sb.WriteString("🏥 *System Health Check*\n\n")
	allUp := true
	for _, c := range checks {
		if c.up {
			sb.WriteString(fmt.Sprintf("✅ *%s* — UP (%s)\n", c.name, c.latency))
		} else {
			sb.WriteString(fmt.Sprintf("❌ *%s* — DOWN (%s)\n   └ `%s`\n", c.name, c.latency, c.errMsg))
			allUp = false
		}
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
