package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/techmaster-vietnam/dd_goshare/config"
)

// ClaimsStringID represents JWT claims with string user ID
type ClaimsStringID struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateTokenStringID generates a new JWT token for string user ID
func GenerateTokenStringID(userID string, email string, cfg *config.Config) (string, error) {
	claims := ClaimsStringID{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.Expiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// Claims represents JWT claims
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
				"message": "Please provide a valid authorization token",
			})
		}

		// Check if token starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token format",
				"message": "Token must be in format: Bearer <token>",
			})
		}

		// Parse and validate token (dùng ClaimsStringID)
		token, err := jwt.ParseWithClaims(tokenString, &ClaimsStringID{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token",
				"message": "Token is invalid or expired",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*ClaimsStringID)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token claims",
				"message": "Token claims are invalid",
			})
		}

		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token expired",
				"message": "Token has expired",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)

		return c.Next()
	}
}

// GenerateToken generates a new JWT token
func GenerateToken(userID uint, email string, cfg *config.Config) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.Expiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *fiber.Ctx) string {
	if userID, ok := c.Locals("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(c *fiber.Ctx) string {
	if userEmail, ok := c.Locals("user_email").(string); ok {
		return userEmail
	}
	return ""
}
