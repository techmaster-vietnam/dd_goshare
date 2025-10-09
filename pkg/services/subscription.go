package services

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/repositories"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

type SubscriptionService struct {
	subscriptionRepo *repositories.SubscriptionRepository
}

func NewSubscriptionService(subscriptionRepo *repositories.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
	}
}

// GetAllSubscriptions lấy tất cả subscriptions
func (s *SubscriptionService) GetAllSubscriptions() ([]models.Subscription, error) {
	return s.subscriptionRepo.GetAllSubscriptions()
}

// GetSubscriptionByID lấy subscription theo ID
func (s *SubscriptionService) GetSubscriptionByID(id string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetSubscriptionByID(id)
}
