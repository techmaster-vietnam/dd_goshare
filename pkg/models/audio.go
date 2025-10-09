package models

// Audio represents an audio file entity
type Audio struct {
	ID       string `gorm:"primaryKey;size:12" json:"id"`
	DialogID string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	FileURL  string `json:"file_url"`
}

// TableName overrides the table name used by Audio to `audios`
func (Audio) TableName() string {
	return "audios"
}
