package models

// Achievement đại diện cho thành tựu
type Achievement struct {
	ID        string         `gorm:"primaryKey;size:12" json:"id"`
	Title     string         `gorm:"size:200" json:"title"`
	IconURL   string         `json:"icon_url"`
}

// TableName chỉ định tên bảng cho Achievement
func (Achievement) TableName() string {
	return "achievements"
}
