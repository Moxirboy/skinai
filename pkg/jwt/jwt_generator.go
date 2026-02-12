package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

// ── JWT Secret ──

var jwtSecret []byte

func init() {
	key := os.Getenv("SIGNING")
	if key == "" {
		key = "skinai-secret-key-2026"
	}
	jwtSecret = []byte(key)
}

// ── Claims ──

type Claims struct {
	UserID  int
	GuestID string
	Role    string // "user", "guest", "doctor"
}

// ── Token Creation ──

// CreateToken creates a JWT for an authenticated user (24h TTL)
func CreateToken(userID int, role string) (string, error) {
	if role == "" {
		role = "user"
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// CreateGuestToken creates a short-lived JWT for guest users (2h TTL)
func CreateGuestToken(guestID string) (string, error) {
	claims := jwt.MapClaims{
		"guest_id": guestID,
		"role":     "guest",
		"exp":      time.Now().Add(2 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ── Token Verification ──

func VerifyToken(tokenStr string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	}

	jwtToken, err := jwt.Parse(tokenStr, keyFunc)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token expired")
			}
		}
		return nil, err
	}

	mapClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return nil, errors.New("invalid token claims")
	}

	claims := &Claims{}

	if role, ok := mapClaims["role"].(string); ok {
		claims.Role = role
	}
	if userID, ok := mapClaims["user_id"].(float64); ok {
		claims.UserID = int(userID)
	}
	if guestID, ok := mapClaims["guest_id"].(string); ok {
		claims.GuestID = guestID
	}

	// Backward compat: old tokens used "sub" for user id
	if sub, ok := mapClaims["sub"].(float64); ok && claims.UserID == 0 {
		claims.UserID = int(sub)
		if claims.Role == "" {
			claims.Role = "user"
		}
	}

	return claims, nil
}
