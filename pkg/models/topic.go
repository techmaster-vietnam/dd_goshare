package models

// Topic đại diện cho chủ đề
type Topic struct {
	ID        string         `gorm:"primaryKey;size:12" json:"id"`
	Title     string         `gorm:"size:200" json:"title"`
}

// TableName chỉ định tên bảng cho Topic
func (Topic) TableName() string {
	return "topics"
}

// TopicListItem represents a topic item in topic list responses
type TopicListItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	DialogNum     int    `json:"dialog_num"`
	FirstDialogID string `json:"first_dialog_id"`
}
type TopicListResponse struct {
	TopicList []TopicListItem `json:"topic_list"`
}

type TitleDialogResponse struct {
	TopicID   string           `json:"topic_id"`
	TopicName string           `json:"topic_name"`
	Dialogs   []DialogResponse `json:"dialogs"`
}
