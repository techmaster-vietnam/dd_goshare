package repositories

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

type TopicRepository struct {
	db *gorm.DB
}

// NewTopicRepository creates a new TopicRepository instance
func NewTopicRepository(db *gorm.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

// GetTopic retrieves a topic by ID
func (r *TopicRepository) GetTopic(id string) (*models.Topic, error) {
	var topic models.Topic
	err := r.db.First(&topic, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}
func (r *TopicRepository) GetTopicByKeyword(keyword string) ([]models.Topic, error) {
	var topics []models.Topic
	err := r.db.Where("LOWER(title) LIKE LOWER(?)", "%"+keyword+"%").Find(&topics).Error
	if err != nil {
		return nil, err
	}
	return topics, nil
}
// GetAllTopics retrieves all topics
func (r *TopicRepository) GetAllTopics() ([]models.Topic, error) {
	var topics []models.Topic
	err := r.db.Find(&topics).Error
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// SearchTopicList returns dialogs grouped by topic in TitleDialogResponse format
func (r *TopicRepository) SearchTopicList(title string, tags []string, page, limit int, sort string, asc bool) ([]models.TitleDialogResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Whitelist sortable columns for topics
	sortColumn := "topics.title"
	switch sort {
	case "id":
		sortColumn = "topics.id"
	case "title":
		sortColumn = "topics.title"
	case "created_at":
		sortColumn = "topics.created_at"
	case "updated_at":
		sortColumn = "topics.updated_at"
	}
	order := "ASC"
	if !asc {
		order = "DESC"
	}

	// Build base query for topics
	base := r.db.Model(&models.Topic{}).
		Select("topics.id, topics.title").
		Joins("JOIN dialogs ON dialogs.topic_id = topics.id")

	// Filter by topic title
	if title != "" {
		base = base.Where("topics.title LIKE ?", "%"+title+"%")
	}

	// Filter by tags (if any tags specified)
	if len(tags) > 0 {
		base = base.Joins("JOIN dialog_tags dt ON dt.dialog_id = dialogs.id").
			Joins("JOIN tags t ON t.id = dt.tag_id").
			Where("t.name IN ?", tags).
			Group("topics.id, topics.title")
	}

	// Count total distinct topics
	var total int64
	if len(tags) > 0 {
		// For tag filtering, we need to count distinct topics that have dialogs with the specified tags
		countQuery := r.db.Model(&models.Topic{}).
			Select("DISTINCT topics.id").
			Joins("JOIN dialogs ON dialogs.topic_id = topics.id").
			Joins("JOIN dialog_tags dt ON dt.dialog_id = dialogs.id").
			Joins("JOIN tags t ON t.id = dt.tag_id").
			Where("t.name IN ?", tags)

		if title != "" {
			countQuery = countQuery.Where("topics.title LIKE ?", "%"+title+"%")
		}

		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		// No tags filter, simple count
		if err := base.Distinct("topics.id").Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	// Fetch topics with pagination
	var topicRows []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	// Build final query with proper grouping and ordering
	query := base.Select("topics.id, topics.title").
		Order(sortColumn + " " + order).
		Limit(limit).
		Offset(offset)

	if err := query.Scan(&topicRows).Error; err != nil {
		return nil, 0, err
	}

	// For each topic, get its dialogs
	result := make([]models.TitleDialogResponse, 0, len(topicRows))
	for _, topicRow := range topicRows {
		// Get dialogs for this topic
		var dialogs []models.Dialog
		dialogQuery := r.db.Where("topic_id = ?", topicRow.ID)

		// Apply tag filter to dialogs if specified
		if len(tags) > 0 {
			dialogQuery = dialogQuery.Joins("JOIN dialog_tags dt ON dt.dialog_id = dialogs.id").
				Joins("JOIN tags t ON t.id = dt.tag_id").
				Where("t.name IN ?", tags).
				Group("dialogs.id").
				Select("dialogs.*")
		}

		if err := dialogQuery.Find(&dialogs).Error; err != nil {
			return nil, 0, err
		}

		// Convert dialogs to DialogResponse
		dialogResponses := make([]models.DialogResponse, 0, len(dialogs))
		for _, dialog := range dialogs {
			dialogResponses = append(dialogResponses, models.DialogResponse{
				DialogID:   dialog.ID,
				DialogName: dialog.Title,
			})
		}

		result = append(result, models.TitleDialogResponse{
			TopicID:   topicRow.ID,
			TopicName: topicRow.Title,
			Dialogs:   dialogResponses,
		})
	}

	return result, total, nil
}
