package models

import (
	"time"
)

// UserStatistics đại diện cho thống kê người dùng
type UserStatistics struct {
	ID                      string     `gorm:"primaryKey;size:12" json:"id"`
	UserID                  string     `gorm:"size:12;uniqueIndex" json:"user_id"`
	TotalDialogsCompleted   int        `gorm:"default:0" json:"total_dialogs_completed"`
	TotalExercisesCompleted int        `gorm:"default:0" json:"total_exercises_completed"`
	TotalTimeSpent          int        `gorm:"default:0" json:"total_time_spent"`
	TotalScore              int        `gorm:"default:0" json:"total_score"`
	LastActive              *time.Time `json:"last_active"`
	CreatedAt               time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// DialogStatistics đại diện cho thống kê hội thoại
type DialogStatistics struct {
	ID               string    `gorm:"primaryKey;size:12" json:"id"`
	DialogID         string    `gorm:"size:12;uniqueIndex" json:"dialog_id"`
	TotalAttempts    int       `gorm:"default:0" json:"total_attempts"`
	TotalCompletions int       `gorm:"default:0" json:"total_completions"`
	AvgScore         float64   `gorm:"type:numeric(4,2);default:0.0" json:"avg_score"`
	AvgRating        float64   `gorm:"type:numeric(3,2);default:0.0" json:"avg_rating"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// SystemStatistics đại diện cho thống kê hệ thống
type SystemStatistics struct {
	ID                 string    `gorm:"primaryKey;size:12" json:"id"`
	Date               time.Time `gorm:"type:date;uniqueIndex" json:"date"`
	NewUsers           int       `gorm:"default:0" json:"new_users"`
	ActiveUsers        int       `gorm:"default:0" json:"active_users"`
	DialogsCompleted   int       `gorm:"default:0" json:"dialogs_completed"`
	ExercisesCompleted int       `gorm:"default:0" json:"exercises_completed"`
	Revenue            float64   `gorm:"type:numeric(12,2);default:0.0" json:"revenue"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName chỉ định tên bảng cho SystemStatistics
func (SystemStatistics) TableName() string {
	return "system_statistics"
}

// TableName chỉ định tên bảng cho DialogStatistics
func (DialogStatistics) TableName() string {
	return "dialog_statistics"
}

// TableName chỉ định tên bảng cho UserStatistics
func (UserStatistics) TableName() string {
	return "user_statistics"
}
