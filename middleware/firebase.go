package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/config"
	"google.golang.org/api/option"
)

var firebaseAuthClient *auth.Client

// InitFirebaseAuth khởi tạo Firebase Auth client
func InitFirebaseAuth(firebaseConfig *config.FirebaseConfig) error {
	ctx := context.Background()

	// Validate config
	if err := firebaseConfig.Validate(); err != nil {
		return fmt.Errorf("invalid firebase config: %v", err)
	}

	// Parse service account key
	var serviceAccountKey map[string]interface{}
	if err := json.Unmarshal([]byte(firebaseConfig.ServiceAccountKey), &serviceAccountKey); err != nil {
		return fmt.Errorf("failed to parse service account key: %v", err)
	}

	// Initialize Firebase app
	opt := option.WithCredentialsJSON([]byte(firebaseConfig.ServiceAccountKey))
	firebaseAppConfig := &firebase.Config{
		ProjectID: firebaseConfig.ProjectID,
	}

	app, err := firebase.NewApp(ctx, firebaseAppConfig, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	firebaseAuthClient = authClient
	return nil
}

// FirebaseAuthMiddleware validates Firebase ID token
func FirebaseAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if Firebase Auth client is initialized
		if firebaseAuthClient == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Firebase Auth not initialized",
			})
		}

		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
				"message": "Please provide a Firebase ID token",
			})
		}

		// Check if token starts with "Bearer "
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token format",
				"message": "Token must be in format: Bearer <firebase_id_token>",
			})
		}

		// Debug: Log token details
		fmt.Printf("🔍 DEBUG Firebase Middleware:\n")
		fmt.Printf("  - Token length: %d\n", len(idToken))
		fmt.Printf("  - Token starts with 'eyJ': %t\n", len(idToken) >= 3 && idToken[:3] == "eyJ")
		fmt.Printf("  - Token first 50 chars: %.50s...\n", idToken)
		if len(idToken) > 50 {
			fmt.Printf("  - Token last 20 chars: ...%s\n", idToken[len(idToken)-20:])
		}

		// Count JWT segments (should be 3: header.payload.signature)
		segments := strings.Split(idToken, ".")
		fmt.Printf("  - JWT segments count: %d\n", len(segments))
		if len(segments) != 3 {
			fmt.Printf("  - ❌ Invalid JWT format - expected 3 segments, got %d\n", len(segments))
			for i, segment := range segments {
				fmt.Printf("    Segment %d length: %d\n", i, len(segment))
			}
		}

		// Verify Firebase ID token
		token, err := firebaseAuthClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			fmt.Printf("  - ❌ Firebase verification error: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid Firebase token",
				"message": fmt.Sprintf("Token verification failed: %v", err),
			})
		} // Extract user information from token
		var email, name, picture, phone, provider string

		if emailClaim, ok := token.Claims["email"].(string); ok {
			email = emailClaim
		}
		if nameClaim, ok := token.Claims["name"].(string); ok {
			name = nameClaim
		}
		if pictureClaim, ok := token.Claims["picture"].(string); ok {
			picture = pictureClaim
		}
		if phoneClaim, ok := token.Claims["phone_number"].(string); ok {
			phone = phoneClaim
		}

		// Extract provider from Firebase claims
		if firebaseClaims, ok := token.Claims["firebase"].(map[string]interface{}); ok {
			if signInProvider, ok := firebaseClaims["sign_in_provider"].(string); ok {
				switch signInProvider {
				case "google.com":
					provider = "google"
				case "apple.com":
					provider = "apple"
				case "facebook.com":
					provider = "facebook"
				case "phone":
					provider = "phone"
				case "password":
					provider = "email"
				default:
					provider = signInProvider
				}
			}
		}

		if provider == "" {
			provider = "firebase"
		}

		// Store Firebase user info in context
		c.Locals("firebase_uid", token.UID)
		c.Locals("firebase_email", email)
		c.Locals("firebase_name", name)
		c.Locals("firebase_picture", picture)
		c.Locals("firebase_phone", phone)
		c.Locals("firebase_provider", provider)
		c.Locals("firebase_token", token)

		// Also store compatible user info for existing code
		c.Locals("user_id", token.UID)
		c.Locals("user_email", email)

		return c.Next()
	}
}

// GetFirebaseUIDFromContext extracts Firebase UID from context
func GetFirebaseUIDFromContext(c *fiber.Ctx) string {
	if uid, ok := c.Locals("firebase_uid").(string); ok {
		return uid
	}
	return ""
}

// GetFirebaseEmailFromContext extracts Firebase email from context
func GetFirebaseEmailFromContext(c *fiber.Ctx) string {
	if email, ok := c.Locals("firebase_email").(string); ok {
		return email
	}
	return ""
}

// GetFirebaseProviderFromContext extracts Firebase provider from context
func GetFirebaseProviderFromContext(c *fiber.Ctx) string {
	if provider, ok := c.Locals("firebase_provider").(string); ok {
		return provider
	}
	return ""
}

// GetFirebaseTokenFromContext extracts Firebase token from context
func GetFirebaseTokenFromContext(c *fiber.Ctx) *auth.Token {
	if token, ok := c.Locals("firebase_token").(*auth.Token); ok {
		return token
	}
	return nil
}

// VerifyFirebaseToken verifies Firebase ID token (utility function)
func VerifyFirebaseToken(idToken string) (*auth.Token, error) {
	if firebaseAuthClient == nil {
		return nil, fmt.Errorf("Firebase Auth not initialized")
	}

	return firebaseAuthClient.VerifyIDToken(context.Background(), idToken)
}
