package models

import "time"

type Subscription struct {
	ID          string    `gorm:"primaryKey;size:12" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:numeric(10,2);default:0.0" json:"price"`
	Duration    string    `gorm:"size:50;not null" json:"duration"` // monthly, yearly
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


// TableName chỉ định tên bảng cho Subscription
func (Subscription) TableName() string {
	return "subscriptions"
}
