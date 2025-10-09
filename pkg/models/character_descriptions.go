package models

type CharacterDescription struct {
	ID           string `json:"id" gorm:"primaryKey;type:varchar(12);uniqueIndex"`
	Description1 string `json:"description_1" gorm:"type:text ; not null"`
	Description2 string `json:"description_2" gorm:"type:text ; not null"`
}

func (CharacterDescription) TableName() string {
	return "character_descriptions"
}
