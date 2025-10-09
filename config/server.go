package config

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// NewServerConfig tạo server config từ environment variables
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: GetEnv("SERVER_PORT", "8081"),
		Host: GetEnv("SERVER_HOST", "0.0.0.0"),
	}
}