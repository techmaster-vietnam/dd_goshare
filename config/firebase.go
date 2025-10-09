package config

import (
	"fmt"
)

// FirebaseConfig holds Firebase configuration for backend
type FirebaseConfig struct {
	ProjectID         string // Required: Firebase project ID
	ServiceAccountKey string // Required: JSON string of service account key for Admin SDK
}

// NewFirebaseConfig tạo Firebase config từ environment variables
func NewFirebaseConfig() *FirebaseConfig {
	return &FirebaseConfig{
		ProjectID:         GetEnv("FIREBASE_PROJECT_ID", ""),
		ServiceAccountKey: GetEnv("FIREBASE_SERVICE_ACCOUNT_KEY", ""),
	}
}

// Validate kiểm tra tính hợp lệ của Firebase config
func (c *FirebaseConfig) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}
	if c.ServiceAccountKey == "" {
		return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_KEY is required")
	}
	return nil
}
