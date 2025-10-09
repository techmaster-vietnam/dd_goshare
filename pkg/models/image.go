package models

// Image represents an image file entity
type Image struct {
	ID       string  `gorm:"primaryKey;size:12" json:"id"`
	DialogID string  `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	PromptID *string `gorm:"size:12;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"prompt_id"`
	FileURL  string  `json:"file_url"`
}

// TableName overrides the table name used by Image to `images`
func (Image) TableName() string {
	return "images"
}

type ImageFigureRequest struct {
	TopicID string `json:"topic_id"`
	Infor_1 string `json:"infor_1"`
	Infor_2 string `json:"infor_2"`
}

type ImageRequest struct {
	TopicID  string `json:"topic_id"`
	DialogID string `json:"dialog_id"`
	Infor_1  string `json:"infor_1"`
	Infor_2  string `json:"infor_2"`
}
