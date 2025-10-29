package models

// DialogTag represents the many-to-many relationship between dialogs and tags
type DialogTag struct {
	ID       string `gorm:"primaryKey;size:12" json:"id"`
	DialogID string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	TagID    string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"tag_id"`

	// Quan hệ với các bảng khác
	Dialog Dialog `gorm:"foreignKey:DialogID;references:ID" json:"-"`
	Tag    Tag    `gorm:"foreignKey:TagID;references:ID" json:"-"`
}

// TableName overrides the table name used by DialogTag to `dialog_tags`
func (DialogTag) TableName() string {
	return "dialog_tags"
}
