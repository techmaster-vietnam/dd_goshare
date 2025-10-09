package models

import (
	"gorm.io/gorm"
)

// Tag đại diện cho một thẻ gắn nhãn
type Tag struct {
	ID        string         `gorm:"primaryKey;size:12" json:"id"`
	Name      string         `gorm:"size:100;uniqueIndex" json:"name"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName chỉ định tên bảng cho Tag
func (Tag) TableName() string {
	return "tags"
}
