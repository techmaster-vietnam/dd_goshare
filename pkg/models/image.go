package models

// Image represents an image file entity
type Image struct {
	ID       string  `gorm:"primaryKey;size:12" json:"id"`
	DialogID *string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	TopicID  *string `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"topic_id"`
	FileURL  string  `json:"file_url" gorm:"type:text; not null"`
	IsFigure bool    `json:"is_figure" gorm:"type:boolean; not null;default:false"`
	AuthorID string  `gorm:"size:12;index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"author_id"`

	// Quan hệ với các bảng khác
	Author User `gorm:"foreignKey:AuthorID;references:ID" json:"-"`
}

// TableName overrides the table name used by Image to `images`
func (Image) TableName() string {
	return "images"
}

type ImageFigureRequest struct {
	Infor_1 string `json:"infor_1"`
	Infor_2 string `json:"infor_2"`
}
type ImageRequest struct {
	PathFigure string `json:"path_figure"`
}
