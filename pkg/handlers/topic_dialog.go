package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	service "github.com/techmaster-vietnam/dd_goshare/pkg/services"
)

type TopicDialogResponse struct {
	TopicID   string           `json:"topic_id"`
	TopicName string           `json:"topic_name"`
	Dialogs   []DialogResponse `json:"dialogs"`
}

type DialogResponse struct {
	DialogID   string `json:"dialog_id"`
	DialogName string `json:"dialog_name"`
}

type TopicDialogHandler struct {
	topicService  *service.TopicService
	dialogService *service.DialogService
}

func NewTopicDialogHandler(topicService *service.TopicService, dialogService *service.DialogService) *TopicDialogHandler {
	return &TopicDialogHandler{topicService: topicService, dialogService: dialogService}
}

func (h *TopicDialogHandler) GetAllTopicsDialogs(c *fiber.Ctx) error {
	title := c.Query("title")
	tagsParam := c.Query("tags")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	sort := c.Query("sort", "title")
	order := c.Query("order", "asc")
	asc := order != "desc"

	var tags []string
	if tagsParam != "" {
		// split by comma and trim
		for _, t := range strings.Split(tagsParam, ",") {
			tt := strings.TrimSpace(t)
			if tt != "" {
				tags = append(tags, tt)
			}
		}
	}

	result, total, err := h.topicService.SearchTopicList(title, tags, page, limit, sort, asc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Errors: ErrorItem{
				Code:    fiber.StatusInternalServerError,
				Message: "Failed to search dialogs: " + err.Error(),
			},
		})
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit) // Ceiling division
	hasNextPage := int64(page) < totalPages
	hasPrevPage := page > 1

	// Prepare response data with pagination info
	responseData := map[string]interface{}{
		"topics": result,
		"pagination": map[string]interface{}{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    total,
			"items_per_page": limit,
			"has_next_page":  hasNextPage,
			"has_prev_page":  hasPrevPage,
		},
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Code: fiber.StatusOK,
		Data: responseData,
	})
}
