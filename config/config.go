package config

// Config holds all configuration for the application
type Config struct {
	Server   *ServerConfig
	Database *DBConfig
	JWT      *JWTConfig
	AI       *AIConfig
	Auth     *AuthConfigs
	Firebase *FirebaseConfig
}

// LoadAllConfigs loads all configuration from environment variables
func LoadAllConfigs() *Config {
	return &Config{
		Server:   NewServerConfig(),
		Database: NewDBConfig(),
		JWT:      NewJWTConfig(),
		AI:       NewAIConfig(),
		Auth:     NewAuthConfigs(),
		Firebase: NewFirebaseConfig(),
	}
}