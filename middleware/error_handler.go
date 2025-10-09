package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

// ErrorHandler is a custom error handler middleware
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's a fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Return JSON error response
		return c.Status(code).JSON(models.ErrorResponse{
			Success: false,
			Error:   message,
			Message: err.Error(),
		})
	}
}
