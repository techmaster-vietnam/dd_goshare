package models

// WordInDialog represents the many-to-many relationship between words and dialogs
type WordInDialog struct {
	DialogID string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	WordID   string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"word_id"`

	// Quan hệ với các bảng khác
	Dialog Dialog `gorm:"foreignKey:DialogID;references:ID" json:"-"`
	Word   Word   `gorm:"foreignKey:WordID;references:ID" json:"-"`
}

// TableName overrides the table name used by WordInDialog to `word_in_dialog`
func (WordInDialog) TableName() string {
	return "word_in_dialog"
}
