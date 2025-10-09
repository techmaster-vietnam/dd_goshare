package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiterConfig returns rate limiter configuration
func RateLimiterConfig() limiter.Config {
	return limiter.Config{
		Max:        100,                // Maximum number of requests
		Expiration: 1 * time.Minute,    // Expiration time
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
		SkipSuccessfulRequests: false,
		SkipFailedRequests:     false,
	}
}

// StrictRateLimiterConfig returns stricter rate limiter for sensitive endpoints
func StrictRateLimiterConfig() limiter.Config {
	return limiter.Config{
		Max:        10,                 // Maximum number of requests
		Expiration: 1 * time.Minute,    // Expiration time
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
		SkipSuccessfulRequests: false,
		SkipFailedRequests:     false,
	}
}

// AuthRateLimiterConfig returns rate limiter for authentication endpoints
func AuthRateLimiterConfig() limiter.Config {
	return limiter.Config{
		Max:        5,                  // Maximum number of requests
		Expiration: 15 * time.Minute,   // Expiration time
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded",
				"message": "Too many authentication attempts. Please try again later.",
			})
		},
		SkipSuccessfulRequests: false,
		SkipFailedRequests:     false,
	}
}

