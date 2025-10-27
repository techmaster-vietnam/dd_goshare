package models

import (
	"encoding/json"
)

// Dialog đại diện cho một đoạn hội thoại
type Dialog struct {
	ID        string          `gorm:"primaryKey;size:12" json:"id"`
	TopicID   string          `gorm:"size:12;index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"topic_id"`
	PrevID    string          `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"prev_id"`
	NextID    string          `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"next_id"`
	Script    string          `gorm:"type:text;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"script"`
	Title     string          `gorm:"size:200;not null" json:"title"`
	RawText   string          `gorm:"type:text;not null" json:"raw_text"`
	AvgRating float64         `gorm:"default:0" json:"avg_rating"`
	Result    json.RawMessage `gorm:"type:jsonb" json:"result"`
	AuthorID  string          `gorm:"size:12;index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"author_id"`
	FixerID   string          `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"fixed_id"`

	// Quan hệ với các bảng khác
	Topic        Topic         `gorm:"foreignKey:TopicID;references:ID" json:"-"`
	Audios       Audio         `gorm:"foreignKey:DialogID" json:"-"`
	Images       []Image       `gorm:"foreignKey:DialogID" json:"-"`
	Tags         []Tag         `gorm:"many2many:dialog_tags" json:"-"`
	Words        []Word        `gorm:"many2many:word_in_dialog" json:"-"`
	FillInBlanks []FillInBlank `gorm:"foreignKey:DialogID" json:"-"`
	Comments     []Comment     `gorm:"foreignKey:DialogID" json:"-"`
	Author       User          `gorm:"foreignKey:AuthorID;references:ID" json:"-"`
	Fixer        User          `gorm:"foreignKey:FixerID;references:ID" json:"-"`
}

// TableName overrides the table name used by Dialog to `dialogs`
func (Dialog) TableName() string {
	return "dialogs"
}

// DialogResultItem represents a dialog item in API responses
type DialogResultItem struct {
	ID         string              `json:"id"`
	PrevID     string              `json:"prev_id"`
	NextID     string              `json:"next_id"`
	Image      string              `json:"image"`
	Title      string              `json:"title"`
	Audio      string              `json:"audio"`
	Sentence   []TimestampSentence `json:"sentence"`
	Words      [][]interface{}     `json:"words"`
	OtherWords []GenWordItem       `json:"other_words"`
}

// TimestampResult represents the timestamp data structure
type TimestampResult struct {
	Audio    string              `json:"audio"`
	Sentence []TimestampSentence `json:"sentence"`
	Words    [][]interface{}     `json:"words"`
}

type FillInWords struct {
	Words []GenWordItem `json:"words"`
}

// TimestampSentence represents a sentence with speaker and timing information
type TimestampSentence struct {
	R  string `json:"r"`  // speaker/role
	S  string `json:"s"`  // sentence text
	B  int    `json:"b"`  // begin position
	T0 int    `json:"t0"` // timestamp
}

// TimestampWord represents a word with timestamp information
type TimestampWord struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence"`
}
type DialogResponse struct {
	DialogID   string `json:"dialog_id"`
	PrevID     string `json:"prev_id"`
	NextID     string `json:"next_id"`
	DialogName string `json:"dialog_name"`
}
