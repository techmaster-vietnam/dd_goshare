package repositories

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// GetSubscriptionByID lấy subscription theo ID
func (r *SubscriptionRepository) GetSubscriptionByID(id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("id = ?", id).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// GetAllSubscriptions lấy tất cả subscriptions
func (r *SubscriptionRepository) GetAllSubscriptions() ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	err := r.db.Find(&subscriptions).Error
	return subscriptions, err
}

// GetSubscriptionByName lấy subscription theo tên
func (r *SubscriptionRepository) GetSubscriptionByName(name string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("name = ?", name).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}
