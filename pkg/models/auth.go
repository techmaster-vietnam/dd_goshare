package models

import "time"

// AuthProvider đại diện cho nhà cung cấp OAuth (Google, Apple, SMS)
type AuthProvider struct {
	ID           string     `gorm:"primaryKey;size:50" json:"id"`
	CustomerID       string     `gorm:"size:50;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"` // Updated to match Customer.ID size
	Provider     string     `gorm:"size:20;not null" json:"provider"`                                           // google, apple, sms
	RefreshToken string     `gorm:"type:text" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// SMSVerification đại diện cho việc xác thực SMS
type SMSVerification struct {
	ID          string    `gorm:"primaryKey;size:12" json:"id"`
	PhoneNumber string    `gorm:"size:20;not null" json:"phone_number"`
	Code        string    `gorm:"size:6;not null" json:"code"`
	Verified    bool      `gorm:"default:false" json:"verified"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// TableName chỉ định tên bảng cho OAuthProvider
func (AuthProvider) TableName() string {
	return "auth_providers"
}

// TableName chỉ định tên bảng cho SMSVerification
func (SMSVerification) TableName() string {
	return "sms_verifications"
}
