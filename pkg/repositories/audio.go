package repositories

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"

	"gorm.io/gorm"
)

// AudioRepository handles database operations for audios
type AudioRepository struct {
	db *gorm.DB
}

// NewAudioRepository creates a new AudioRepository instance
func NewAudioRepository(db *gorm.DB) *AudioRepository {
	return &AudioRepository{db: db}
}

// GetAudiosByDialogID retrieves all audios for a specific dialog
func (r *AudioRepository) GetAudiosByDialogID(dialogID string) ([]models.Audio, error) {
	var audios []models.Audio
	err := r.db.Where("dialog_id = ?", dialogID).Find(&audios).Error
	if err != nil {
		return nil, err
	}
	return audios, nil
}

