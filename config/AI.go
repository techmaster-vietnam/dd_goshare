package config

// AIConfig holds AI configuration
type AIConfig struct {
	APIKey       string
	BaseURL      string
	GenImageURL  string
	EditImageURL string
}

// NewAIConfig tạo AI config từ environment variables
func NewAIConfig() *AIConfig {
	return &AIConfig{
		APIKey:       GetEnv("API_KEY", ""),
		BaseURL:      GetEnv("API_URL", ""),
		GenImageURL:  GetEnv("GEN_IMAGE_URL", ""),
		EditImageURL: GetEnv("EDIT_IMAGE_URL", ""),
	}
}