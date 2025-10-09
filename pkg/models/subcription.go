package models

// Subscription đại diện cho gói đăng ký
type Subscription struct {
	ID          string         `gorm:"primaryKey;size:12" json:"id"`
	Name        string         `gorm:"size:100" json:"name"`
	Description string         `json:"description"`
	Price       float64        `gorm:"type:numeric(10,2);default:0.0" json:"price"`
	Duration    string         `gorm:"size:50" json:"duration"`
}

// TableName chỉ định tên bảng cho Subscription
func (Subscription) TableName() string {
	return "subscriptions"
}
