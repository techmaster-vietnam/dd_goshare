package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

// CallAI gọi đến AI service với system_message và user_message
// endpoint: URL của AI service
// apiKey: API key để xác thực
// systemMessage: Tin nhắn hệ thống
// userMessage: Tin nhắn người dùng

func CallAI(systemMessage, userMessage string) (string, error) {
	apiKey := os.Getenv("API_KEY")
	apiURL := os.Getenv("GEN_IMAGE_URL")

	// Tạo request body
	reqBody := models.AIRequest{
		Model:       "gpt-4o",
		Temperature: 0.3,
		Messages: []models.Message{
			{
				Role:    "system",
				Content: systemMessage,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("lỗi khi tạo JSON request: %v", err)
	}

	// Tạo HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("lỗi khi tạo HTTP request: %v", err)
	}

	// Thêm headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Gửi request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lỗi khi gửi request: %v", err)
	}
	defer resp.Body.Close()

	// Đọc response
	var aiResp models.AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return "", fmt.Errorf("lỗi khi đọc response: %v", err)
	}

	// Kiểm tra lỗi từ response
	if aiResp.Error != nil {
		return "", fmt.Errorf("lỗi từ AI service: %s", aiResp.Error.Message)
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("không có kết quả trả về từ AI service")
	}

	// Lấy nội dung và loại bỏ markdown code block nếu có
	content := aiResp.Choices[0].Message.Content
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content, nil
}
