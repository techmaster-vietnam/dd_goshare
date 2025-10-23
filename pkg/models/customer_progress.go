package models

// CustomerTopicProgress đại diện cho tiến trình của khách hàng trong một chủ đề
type CustomerTopicProgress struct {
	CustomerID        string   `gorm:"primaryKey;size:50;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"`
	TopicID            string   `gorm:"primaryKey;size:12;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"topic_id"`
	CompletedDialogIDs []string `gorm:"type:text;serializer:json;default:'[]'" json:"completed_dialog_ids"`
	IsCompleted        bool     `gorm:"default:false" json:"is_completed"`
	NextDialogID       string   `gorm:"size:12" json:"next_dialog_id,omitempty"`
	LastUpdated        int64    `gorm:"not null" json:"last_updated"`
}

// DialogCompletion đại diện cho việc hoàn thành các bài tập trong hội thoại
type DialogCompletion struct {
	CustomerID         string `gorm:"primaryKey;size:50;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"customer_id"`
	DialogID           string `gorm:"primaryKey;size:12;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	TopicID            string `gorm:"size:12;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"topic_id,omitempty"`
	ListeningCompleted bool   `gorm:"not null;default:false" json:"listening_completed"`
	SpeakingCompleted  bool   `gorm:"not null;default:false" json:"speaking_completed"`
	WritingCompleted   bool   `gorm:"not null;default:false" json:"writing_completed"`
	ScoreSpeaking      int    `gorm:"default:0" json:"score_speaking"`
	ScoreWriting       int    `gorm:"default:0" json:"score_writing"`
	// IsSync             bool   `gorm:"not null" json:"isSync"`
}

// TableName chỉ định tên bảng cho CustomerTopicProgress
func (CustomerTopicProgress) TableName() string {
	return "customer_topic_progress"
}

// TableName chỉ định tên bảng cho DialogCompletion
func (DialogCompletion) TableName() string {
	return "dialog_completions"
}
