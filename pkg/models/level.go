package models

// Level represents a difficulty level
type Level struct {
	ID        string `json:"id" gorm:"primaryKey;type:varchar(12);uniqueIndex"`
	LevelName string `json:"level_name" gorm:"type:text"`
}

// TableName overrides the table name used by Level to `levels`
func (Level) TableName() string {
	return "levels"
}
