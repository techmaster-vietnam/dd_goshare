package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/services"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetAllSubscriptions lấy tất cả subscriptions
// @Summary Get all subscriptions
// @Description Retrieve all available subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse
// @Router /api/subscriptions [get]
func (h *SubscriptionHandler) GetAllSubscriptions(c *fiber.Ctx) error {
	subscriptions, err := h.subscriptionService.GetAllSubscriptions()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Failed to get subscriptions",
			Error:   err.Error(),
		})
	}

	return c.JSON(subscriptions)
}

// GetSubscriptionByID lấy subscription theo ID
// @Summary Get subscription by ID
// @Description Retrieve a subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscriptionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Subscription ID is required",
		})
	}

	subscription, err := h.subscriptionService.GetSubscriptionByID(id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
			Message: "Subscription not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(subscription)
}