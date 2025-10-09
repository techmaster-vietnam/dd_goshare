package middleware

import (
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// CustomLogger returns a custom logger middleware configuration
func CustomLogger() logger.Config {
	return logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     logger.ConfigDefault.Output,
	}
}
