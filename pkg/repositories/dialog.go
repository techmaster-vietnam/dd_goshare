package repositories

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

type DialogRepository struct {
	db *gorm.DB
}

// NewDialogRepository creates a new DialogRepository instance
func NewDialogRepository(db *gorm.DB) *DialogRepository {
	return &DialogRepository{db: db}
}

// GetDialog retrieves a dialog by ID
func (r *DialogRepository) GetDialog(id string) (*models.Dialog, error) {
	var dialog models.Dialog
	err := r.db.First(&dialog, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dialog, nil
}
// GetAllDialogs retrieves all dialogs
func (r *DialogRepository) GetAllDialogs() ([]models.Dialog, error) {
	var dialogs []models.Dialog
	err := r.db.Find(&dialogs).Error
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}
func (r *DialogRepository) GetFirstDialogByTopicID(topicID string) (*models.Dialog, error) {
	var dialog models.Dialog
	err := r.db.Where("topic_id = ? AND (prev_id IS NULL OR prev_id = '')", topicID).First(&dialog).Error
	if err != nil {
		return nil, err
	}
	return &dialog, nil
}
func (r *DialogRepository) GetDialogTitleByKeyword(keyword string) ([]models.Dialog, error) {
	var dialogs []models.Dialog
	err := r.db.Where("LOWER(title) LIKE LOWER(?)", "%"+keyword+"%").Limit(10).Order("title ASC").Find(&dialogs).Error
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}
// GetDialogsRawTextKeyword retrieves all dialogs containing the keyword in raw_text (case-insensitive)
func (r *DialogRepository) GetDialogsRawTextKeyword(keyword string) ([]models.Dialog, error) {
	var dialogs []models.Dialog
	err := r.db.Where("LOWER(raw_text) LIKE LOWER(?)", "%"+keyword+"%").Find(&dialogs).Error
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}
func (r *DialogRepository) GetAllDialogsByTopicID(topicID string) ([]models.Dialog, error) {
	var dialogs []models.Dialog
	err := r.db.Where("topic_id = ?", topicID).Find(&dialogs).Error
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}
func (r *DialogRepository) GetTagsByDialogID(dialogID string) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Joins("JOIN dialog_tags ON dialog_tags.tag_id = tags.id").
		Where("dialog_tags.dialog_id = ?", dialogID).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
