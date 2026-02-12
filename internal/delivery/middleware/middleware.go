package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"testDeployment/pkg/jwt"
)

// ══════════════════════════════════════════════
// Helper: Get user ID from context (set by auth middleware)
// ══════════════════════════════════════════════

// GetUserID extracts the authenticated user ID from gin context.
// Checks context first (set by JWT middleware), then session fallback.
func GetUserID(c *gin.Context) int {
	if id, exists := c.Get("user_id"); exists {
		if intID, ok := id.(int); ok && intID > 0 {
			return intID
		}
	}
	// Fallback: session
	s := sessions.Default(c)
	if id := s.Get("userId"); id != nil {
		if intID, ok := id.(int); ok {
			return intID
		}
	}
	return 0
}

// GetRole returns the role of the current user from context
func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return "anonymous"
}

// ══════════════════════════════════════════════
// Auth extraction (JWT → Session fallback)
// ══════════════════════════════════════════════

func extractAuth(c *gin.Context) *jwt.Claims {
	// 1. Check Authorization: Bearer <token>
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.VerifyToken(tokenStr)
		if err == nil {
			return claims
		}
	}

	// 2. Fallback: session cookie
	session := sessions.Default(c)
	if userID := session.Get("userId"); userID != nil {
		if id, ok := userID.(int); ok && id > 0 {
			return &jwt.Claims{
				UserID: id,
				Role:   "user",
			}
		}
	}

	return nil
}

// ══════════════════════════════════════════════
// Middleware: OptionalAuth
// Sets user_id + role in context if auth is present.
// Does NOT block unauthenticated requests.
// ══════════════════════════════════════════════

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := extractAuth(c)
		if claims != nil {
			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)
			c.Set("guest_id", claims.GuestID)
		} else {
			c.Set("role", "anonymous")
		}
		c.Next()
	}
}

// ══════════════════════════════════════════════
// Middleware: AuthMiddleware
// Requires authenticated user (not guest, not anonymous).
// Blocks with 401 if no valid auth found.
// ══════════════════════════════════════════════

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := extractAuth(c)
		if claims == nil || claims.UserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Please login or signup to access this resource",
			})
			c.Abort()
			return
		}
		if claims.Role == "guest" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Guests cannot access this resource. Please register.",
			})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// ══════════════════════════════════════════════
// Guest Rate Limiter
// ══════════════════════════════════════════════

const (
	GuestAILimit     = 5 // max AI text requests per day
	GuestUploadLimit = 3 // max image uploads per day
)

type guestUsage struct {
	aiCount     int
	uploadCount int
	resetAt     time.Time
}

type RateLimiter struct {
	mu    sync.Mutex
	usage map[string]*guestUsage
}

// GuestLimiter is the singleton rate limiter for guest users
var GuestLimiter = newRateLimiter()

func newRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		usage: make(map[string]*guestUsage),
	}
	// Background cleanup every hour
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for key, u := range rl.usage {
		if now.After(u.resetAt) {
			delete(rl.usage, key)
		}
	}
}

func (rl *RateLimiter) allowAI(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	u, ok := rl.usage[key]
	if !ok || time.Now().After(u.resetAt) {
		u = &guestUsage{resetAt: time.Now().Add(24 * time.Hour)}
		rl.usage[key] = u
	}
	if u.aiCount >= GuestAILimit {
		return false
	}
	u.aiCount++
	return true
}

func (rl *RateLimiter) allowUpload(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	u, ok := rl.usage[key]
	if !ok || time.Now().After(u.resetAt) {
		u = &guestUsage{resetAt: time.Now().Add(24 * time.Hour)}
		rl.usage[key] = u
	}
	if u.uploadCount >= GuestUploadLimit {
		return false
	}
	u.uploadCount++
	return true
}

func (rl *RateLimiter) Remaining(key string) (ai int, upload int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	u, ok := rl.usage[key]
	if !ok || time.Now().After(u.resetAt) {
		return GuestAILimit, GuestUploadLimit
	}
	return GuestAILimit - u.aiCount, GuestUploadLimit - u.uploadCount
}

// ══════════════════════════════════════════════
// Helper: resolve guest rate-limit key from context
// ══════════════════════════════════════════════

// GuestKey returns the rate-limiter key for the current guest.
// Prefers guest_id (from JWT) falling back to client IP.
func GuestKey(c *gin.Context) string {
	if guestID, exists := c.Get("guest_id"); exists {
		if gid, ok := guestID.(string); ok && gid != "" {
			return gid
		}
	}
	return c.ClientIP()
}

// ══════════════════════════════════════════════
// Middleware: AIRateLimit
// Blocks anonymous users entirely.
// Limits guest AI usage per day.
// Registered users pass through freely.
// ══════════════════════════════════════════════

func AIRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)

		// Anonymous: must get at least a guest token
		if role == "anonymous" || role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Please login, signup, or continue as guest to use AI",
			})
			c.Abort()
			return
		}

		// Registered users: no limit
		if role == "user" || role == "doctor" {
			c.Next()
			return
		}

		// Guest: rate limited
		if role == "guest" {
			key := GuestKey(c)

			isUpload := strings.Contains(c.Request.URL.Path, "/upload")

			if isUpload {
				if !GuestLimiter.allowUpload(key) {
					aiLeft, uploadLeft := GuestLimiter.Remaining(key)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":              "rate_limit_exceeded",
						"message":            "Guest upload limit reached. Register for unlimited access!",
						"daily_upload_limit":  GuestUploadLimit,
						"remaining_ai":        aiLeft,
						"remaining_uploads":   uploadLeft,
					})
					c.Abort()
					return
				}
			} else {
				if !GuestLimiter.allowAI(key) {
					aiLeft, uploadLeft := GuestLimiter.Remaining(key)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":            "rate_limit_exceeded",
						"message":          "Guest AI limit reached. Register for unlimited access!",
						"daily_ai_limit":   GuestAILimit,
						"remaining_ai":     aiLeft,
						"remaining_uploads": uploadLeft,
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// ══════════════════════════════════════════════
// Middleware: GuestInfo
// Returns remaining AI/upload quota for guest users.
// Attach to a GET endpoint like /auth/guest/status
// ══════════════════════════════════════════════

func GuestRemainingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role != "guest" {
			c.JSON(200, gin.H{
				"role":    role,
				"limited": false,
			})
			return
		}

		key := GuestKey(c)

		aiLeft, uploadLeft := GuestLimiter.Remaining(key)
		c.JSON(200, gin.H{
			"role":               "guest",
			"limited":            true,
			"remaining_ai":       aiLeft,
			"remaining_uploads":  uploadLeft,
			"daily_ai_limit":     GuestAILimit,
			"daily_upload_limit": GuestUploadLimit,
		})
	}
}
