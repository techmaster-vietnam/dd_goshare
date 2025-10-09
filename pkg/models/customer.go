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
	Streak      int        `gorm:"default:0" json:"streak"`
	Score       int        `gorm:"default:0" json:"score"`
	LastLogin   *time.Time `json:"last_login"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// CustomerAchievement đại diện cho thành tựu của khách hàng
type CustomerAchievement struct {
	ID            string    `gorm:"primaryKey;size:12" json:"id"`
	CustomerID    string    `gorm:"size:50;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"` // Updated to match Customer.ID
	AchievementID string    `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"achievement_id"`
	UnlockedAt    time.Time `json:"unlocked_at"`
}

// CustomerDialogExercise đại diện cho bài tập của khách hàng trong một hội thoại
type CustomerDialogExercise struct {
	ID           string     `gorm:"primaryKey;size:12" json:"id"`
	ProgressID   string     `gorm:"size:12;index:idx_progress_ex_type,unique;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"progress_id"`
	ExerciseType string     `gorm:"size:50;index:idx_progress_ex_type,unique" json:"exercise_type"`
	Status       string     `gorm:"size:50;default:pending" json:"status"`
	Score        int        `gorm:"default:0" json:"score"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// CustomerDialogProgress đại diện cho tiến trình của khách hàng trong một hội thoại
type CustomerDialogProgress struct {
	ID            string     `gorm:"primaryKey;size:12" json:"id"`
	CustomerID    string     `gorm:"size:50;uniqueIndex:idx_customer_dialog;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"` // Updated to match Customer.ID
	DialogID      string     `gorm:"size:12;uniqueIndex:idx_customer_dialog;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	OverallStatus string     `gorm:"size:50;default:in_progress" json:"overall_status"`
	OverallScore  int        `gorm:"default:0" json:"overall_score"`
	LastAccessed  time.Time  `json:"last_accessed"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// CustomerSubscription đại diện cho đăng ký của khách hàng
type CustomerSubscription struct {
	ID             string     `gorm:"primaryKey;size:12" json:"id"`
	CustomerID     string     `gorm:"size:50;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"` // Updated to match Customer.ID size
	SubscriptionID string     `gorm:"size:20;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"subscription_id"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ExpiredAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName chỉ định tên bảng cho CustomerSubscription
func (CustomerSubscription) TableName() string {
	return "customer_subscriptions"
}

// TableName chỉ định tên bảng cho CustomerDialogProgress
func (CustomerDialogProgress) TableName() string {
	return "customer_dialog_progress"
}

// TableName chỉ định tên bảng cho CustomerDialogExercise
func (CustomerDialogExercise) TableName() string {
	return "customer_dialog_exercises"
}

// TableName chỉ định tên bảng cho CustomerAchievement
func (CustomerAchievement) TableName() string {
	return "customer_achievements"
}

// TableName chỉ định tên bảng cho Customer
func (Customer) TableName() string {
	return "customers"
}
