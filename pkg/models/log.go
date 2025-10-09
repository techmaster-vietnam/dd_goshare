package models

import (
	"time"

	"gorm.io/gorm"
)

// HistoryLog đại diện cho lịch sử hoạt động
type HistoryLog struct {
	ID        string         `gorm:"primaryKey;size:12" json:"id"`
	TableRef  string         `gorm:"column:table_name;size:50;index" json:"table_name"`
	RecordID  string         `gorm:"size:12" json:"record_id"`
	UserID    string         `gorm:"size:12" json:"user_id"`
	Action    string         `gorm:"size:50" json:"action"`
	Changes   string         `gorm:"type:text" json:"changes"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PaymentLog đại diện cho lịch sử thanh toán
type PaymentLog struct {
	ID           string         `gorm:"primaryKey;size:12" json:"id"`
	PaymentID    string         `gorm:"size:12;index" json:"payment_id"`
	StatusBefore string         `json:"status_before"`
	StatusAfter  string         `json:"status_after"`
	RawResponse  string         `gorm:"type:text" json:"raw_response"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName chỉ định tên bảng cho PaymentLog
func (PaymentLog) TableName() string {
	return "payment_logs"
}

// TableName chỉ định tên bảng cho HistoryLog
func (HistoryLog) TableName() string {
	return "history_log"
}
