package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID               string         `gorm:"primaryKey;size:12" json:"id"`
	CustomerID       string         `gorm:"size:50;index;not null" json:"customer_id"`
	SubscriptionID   *string        `gorm:"size:12;index" json:"subscription_id"`
	Amount           float64        `gorm:"type:numeric(10,2)" json:"amount"`
	Status           string         `gorm:"size:50;default:'pending'" json:"status"` // pending, success, failed, refunded
	Method           string         `gorm:"size:50;not null" json:"method"`           // momo, vnpay, apple_iap, google_iap...
	TransactionID    *string        `gorm:"size:255" json:"transaction_id"`           // mã giao dịch từ Apple/Google/MoMo/VNPAY
	ReceiptData      *string        `gorm:"type:text" json:"receipt_data"`            // lưu receipt base64 hoặc token
	Platform         *string        `gorm:"size:50" json:"platform"`                  // ios, android, web
	VerifiedAt       *time.Time     `json:"verified_at"`                              // thời điểm xác thực thành công
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}


// TableName chỉ định tên bảng cho Payment
func (Payment) TableName() string {
	return "payments"
}
