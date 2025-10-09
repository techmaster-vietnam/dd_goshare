package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Word đại diện cho một từ vựng
type Word struct {
	ID            string         `gorm:"primaryKey;size:12" json:"id"`
	PromptID      *string        `gorm:"size:12" json:"prompt_id"`
	Text          string         `gorm:"size:200" json:"text"`
	Pronunciation string         `json:"pronunciation"`
	Meaning       string         `json:"meaning"`
	PartOfSpeech  string         `json:"part_of_speech"`
	Example       string         `json:"example"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName chỉ định tên bảng cho Word
func (Word) TableName() string {
	return "words"
}

// FillInBlank đại diện cho bài tập điền từ
type FillInBlank struct {
	ID         string          `gorm:"primaryKey;size:12" json:"id"`
	DialogID   string          `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	PromptID   *string         `gorm:"size:12;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"prompt_id"`
	WordsIndex json.RawMessage `gorm:"type:jsonb" json:"words_index"`
}

func (FillInBlank) TableName() string {
	return "fill_in_blanks"
}

// NewWordData represents the structure of newWord.json
type NewWordData struct {
	NewWord []WordItem `json:"newWord"`
}

// WordItem represents a single word item in the JSON
type WordItem struct {
	Word         string `json:"w"`
	Meaning      string `json:"meaning"`
	Example      string `json:"ex"`
	PartOfSpeech string `json:"pos"`
}
type GenWordItem struct {
	Word  string `json:"word"`
	Start int    `json:"index"`
}
type TranslateRequest struct {
	DialogID string `json:"dialog_id"`
	Word     string `json:"word"`
	Sentence string `json:"sentence"`
}

// WordPositionReport represents the report structure for word positions
type WordPositionReport struct {
	DialogID       string               `json:"dialog_id"`
	TotalWords     int                  `json:"total_words"`
	CorrectWords   int                  `json:"correct_words"`
	IncorrectWords int                  `json:"incorrect_words"`
	NotFoundWords  int                  `json:"not_found_words"`
	Words          []WordPositionStatus `json:"words"`
}

// WordPositionStatus represents the status of a single word
type WordPositionStatus struct {
	Index           int    `json:"index"`
	Word            string `json:"word"`
	CurrentPosition int    `json:"current_position"`
	CorrectPosition *int   `json:"correct_position,omitempty"`
	Status          string `json:"status"`
	StatusCode      string `json:"status_code"` // "correct", "incorrect", "not_found"
	Message         string `json:"message,omitempty"`
}
