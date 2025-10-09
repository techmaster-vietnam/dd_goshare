package config

// AuthConfigs now minimal since Firebase handles all authentication
// Keep this struct for potential future JWT configuration
type AuthConfigs struct {
	// Firebase handles all OAuth flows, so no individual provider configs needed
	// JWT secret for token generation (if not using Firebase tokens directly)
	JWTSecret string
}

// NewAuthConfigs creates minimal auth configs
func NewAuthConfigs() *AuthConfigs {
	return &AuthConfigs{
		JWTSecret: GetEnv("JWT_SECRET", ""),
	}
}

// Validate is now minimal since Firebase handles auth validation
func (c *AuthConfigs) Validate() error {
	// Firebase handles all authentication validation
	// No specific validations needed for local configs
	return nil
}
