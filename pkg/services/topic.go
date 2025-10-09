package services

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	repo "github.com/techmaster-vietnam/dd_goshare/pkg/repositories"
)

type TopicService struct {
	topicRepo  *repo.TopicRepository
	dialogRepo *repo.DialogRepository
}

// NewTopicService creates a new TopicService instance
func NewTopicService(topicRepo *repo.TopicRepository, dialogRepo *repo.DialogRepository) *TopicService {
	return &TopicService{
		topicRepo:  topicRepo,
		dialogRepo: dialogRepo,
	}
}

// GetTopic retrieves a topic by ID
func (s *TopicService) GetTopic(id string) (*models.Topic, error) {
	if id == "" {
		return nil, ErrTopicIDRequired
	}

	return s.topicRepo.GetTopic(id)
}

// GetAllTopics retrieves all topics
func (s *TopicService) GetAllTopics() ([]models.Topic, error) {
	return s.topicRepo.GetAllTopics()
}

func (s *TopicService) SearchKeyword(keyword string) ([]models.Topic, error) {
	return s.topicRepo.GetTopicByKeyword(keyword)
}

// SearchTopicList returns dialogs grouped by topic in TitleDialogResponse format
func (s *TopicService) SearchTopicList(title string, tags []string, page, limit int, sort string, asc bool) ([]models.TitleDialogResponse, int64, error) {
	return s.topicRepo.SearchTopicList(title, tags, page, limit, sort, asc)
}
