package utils

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/techmaster-vietnam/dd_goshare/config"
)

// extractUserIDFromToken extracts user_id from JWT token
func ExtractUserIDFromToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "" // No Bearer prefix found
	}

	// Load config to get JWT secret
	cfg := config.LoadAllConfigs()

	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		// Return secret key from config
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		log.Printf("DEBUG: JWT parse error: %v", err)
		return ""
	}

	// Extract user_id from claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, exists := claims["user_id"]; exists {
			if userIDStr, ok := userID.(string); ok {
				return userIDStr
			}
		}
	}

	return ""
}