package models

import (
	"time"
)

// Customer đại diện cho người dùng trong hệ thống
type Customer struct {
	ID          string     `gorm:"primaryKey;size:50" json:"id"` // Increased size for Firebase UID
	Name        string     `gorm:"size:50;uniqueIndex" json:"name"`
	Email       string     `gorm:"size:100;uniqueIndex" json:"email"`
	PhoneNumber *string    `gorm:"size:20;uniqueIndex" json:"phone_number"` // ✅ Thêm unique constraint, nullable
	Password    string     `gorm:"size:255" json:"-"`
	AvatarURL   string     `json:"avatar_url"`
	LastLogin   *time.Time `json:"last_login"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// CustomerAchievement đại diện cho thành tựu của khách hàng
type CustomerAchievement struct {
	CustomerID    string `gorm:"primaryKey;size:50;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"` // Updated to match Customer.ID
	AchievementID string `gorm:"primaryKey;size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"achievement_id"`
	Claimed       bool   `gorm:"default:false" json:"claimed"`
	Unlocked      bool   `gorm:"default:false" json:"unlocked"`

	// Quan hệ với các bảng khác
	Customer    Customer    `gorm:"foreignKey:CustomerID;references:ID" json:"-"`
	Achievement Achievement `gorm:"foreignKey:AchievementID;references:ID" json:"-"`
}

type CustomerSubscription struct {
	ID                  string     `gorm:"primaryKey;size:50" json:"id"`
	CustomerID          string     `gorm:"size:50;index;not null" json:"customer_id"`
	SubscriptionID      string     `gorm:"size:32;index;not null" json:"subscription_id"`
	PaymentID           *string    `gorm:"size:32;index" json:"payment_id"`
	OriginalTransaction string     `gorm:"size:255;uniqueIndex" json:"original_transaction"` // định danh duy nhất cho 1 chu kỳ
	ExpiredAt           time.Time  `json:"expired_at"`
	StartDate           time.Time  `gorm:"not null" json:"start_date"`
	EndDate             *time.Time `json:"end_date"`
	IsActive            bool       `gorm:"default:true" json:"is_active"`
	AutoRenew           bool       `gorm:"default:false" json:"auto_renew"`
	CanceledAt          *time.Time `json:"canceled_at"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type UpdateCustomerStatisticsRequest struct {
	TotalDialogsCompleted   *int `json:"total_dialogs_completed"`
	TotalExercisesCompleted *int `json:"total_exercises_completed"`
	TotalTimeSpent          *int `json:"total_time_spent"`
}

type UpdateCustomerInfoRequest struct {
	Name   *string `json:"name"`
	Streak *int    `json:"streak"`
	Score  *int    `json:"score"`
	// AvatarURL will be handled via file upload in handler
}

// TableName chỉ định tên bảng cho CustomerSubscription
func (CustomerSubscription) TableName() string {
	return "customer_subscriptions"
}

// TableName chỉ định tên bảng cho CustomerAchievement
func (CustomerAchievement) TableName() string {
	return "customer_achievements"
}

// TableName chỉ định tên bảng cho Customer
func (Customer) TableName() string {
	return "customers"
}
