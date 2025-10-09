package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	"github.com/goccy/go-json"
)

// ImageRequest cấu trúc request để tạo ảnh
type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// ImageResponse cấu trúc response từ API tạo ảnh
type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateImage tạo ảnh bằng OpenAI DALL-E API
// prompt: Mô tả ảnh cần tạo
// n: Số lượng ảnh cần tạo (mặc định là 1)
// size: Kích thước ảnh (mặc định là "640x360")
// responseFormat: Định dạng response ("url" hoặc "b64_json")
func GenerateImage(prompt string, n int, size string, responseFormat string) ([]string, error) {
	apiKey := os.Getenv("API_KEY")
	apiURL := os.Getenv("GEN_IMAGE_URL")

	if n <= 0 {
		n = 1
	}
	if size == "" {
		size = "1024x1024"
	}

	// Tạo request body
	reqBody := ImageRequest{
		Model:  "gpt-image-1",
		Prompt: prompt,
		N:      n,
		Size:   size,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi tạo JSON request: %v", err)
	}

	// Tạo HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("lỗi khi tạo HTTP request: %v", err)
	}

	// Thêm headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Gửi request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi gửi request: %v", err)
	}
	defer resp.Body.Close()

	// Đọc response
	var imageResp ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return nil, fmt.Errorf("lỗi khi đọc response: %v", err)
	}

	// Kiểm tra lỗi từ response
	if imageResp.Error != nil {
		return nil, fmt.Errorf("lỗi từ AI service: %s", imageResp.Error.Message)
	}

	// Lấy danh sách base64 ảnh
	var base64Images []string
	for _, data := range imageResp.Data {
		if data.B64JSON != "" {
			base64Images = append(base64Images, data.B64JSON)
		}
	}

	return base64Images, nil
}
func EditImage(imagePath string, prompt string) ([]string, error) {
	apiKey := os.Getenv("API_KEY")
	apiURL := os.Getenv("GEN_IMAGE_URL")

	// Tạo form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// file ảnh reference
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Printf("không mở được file ảnh: %v", err)
		return nil, err
	}
	// part, _ := writer.CreateFormFile("image", "ChatGPT-Image-11_15_43-3-thg-9_-2025.webp")

	// Tạo header cho file upload
	fileStat, _ := file.Stat()
	filePart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="image"; filename="%s"`, fileStat.Name())},
		"Content-Type":        []string{"image/webp"}, // hoặc image/jpeg
	})
	if err != nil {
		panic(err)
	}

	io.Copy(filePart, file)
	file.Close()

	// nếu có mask
	// maskFile, _ := os.Open("mask.png")
	// maskPart, _ := writer.CreateFormFile("mask", "mask.png")
	// io.Copy(maskPart, maskFile)
	// maskFile.Close()

	// các field khác
	writer.WriteField("model", "gpt-image-1")
	writer.WriteField("prompt", prompt)
	writer.WriteField("size", "auto")
	writer.Close()

	// HTTP request
	req, _ := http.NewRequest("POST", apiURL, body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// respBody, _ := io.ReadAll(resp.Body)
	// Đọc response
	var imageResp ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return nil, fmt.Errorf("lỗi khi đọc response: %v", err)
	}

	// Kiểm tra lỗi từ response
	if imageResp.Error != nil {
		return nil, fmt.Errorf("lỗi từ AI service: %s", imageResp.Error.Message)
	}

	// Lấy danh sách base64 ảnh
	var base64Images []string
	for _, data := range imageResp.Data {
		if data.B64JSON != "" {
			base64Images = append(base64Images, data.B64JSON)
		}
	}

	return base64Images, nil
}
