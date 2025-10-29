package models

// Achievement đại diện cho thành tựu
type Achievement struct {
	ID             string `gorm:"primaryKey;size:12" json:"id"`
	Title          string `gorm:"size:200" json:"title"`
	IconUnicode    string `json:"icon_unicode"`
	ConditionType  string `gorm:"size:50" json:"condition_type"`
	ConditionValue int    `gorm:"default:0" json:"condition_value"`
	RewardPoints   int    `gorm:"default:0" json:"reward_points"`
	GroupType      string `gorm:"size:50" json:"group_type"`
}

// TableName chỉ định tên bảng cho Achievement
func (Achievement) TableName() string {
	return "achievements"
}
