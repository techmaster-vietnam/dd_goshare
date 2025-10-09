package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment đại diện cho thanh toán
type Payment struct {
	ID             string         `gorm:"primaryKey;size:12" json:"id"`
	UserID         string         `gorm:"size:12;index" json:"user_id"`
	SubscriptionID *string        `gorm:"size:12" json:"subscription_id"`
	Amount         float64        `gorm:"type:numeric(10,2)" json:"amount"`
	Status         string         `gorm:"size:50;default:pending" json:"status"`
	Method         string         `gorm:"size:100" json:"method"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName chỉ định tên bảng cho Payment
func (Payment) TableName() string {
	return "payments"
}
