package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// DBConfig holds database configuration
type DBConfig struct {
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSLMODE,required"`
}

// NewDBConfig tạo database config từ environment variables
func NewDBConfig() *DBConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	config := &DBConfig{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to parse env variables: %v", err)
	}

	return config
}