package config

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
	Expiry int // in hours
}

// NewJWTConfig tạo JWT config từ environment variables
func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret: GetEnv("JWT_SECRET", "your-secret-key"),
		Expiry: GetEnvAsInt("JWT_EXPIRY", 24),
	}
}