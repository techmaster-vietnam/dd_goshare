package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// generateRandomString tạo chuỗi ngẫu nhiên với độ dài cho trước
func generateRandomString(length int) (string, error) {
	// Tính toán số byte cần thiết để tạo chuỗi ngẫu nhiên
	// Mỗi ký tự base64 có thể biểu diễn 6 bit, nên cần (length * 6) / 8 byte
	bytes := make([]byte, (length*6)/8+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Chuyển đổi sang base64 và loại bỏ các ký tự không mong muốn
	str := base64.URLEncoding.EncodeToString(bytes)
	str = strings.ReplaceAll(str, "-", "")
	str = strings.ReplaceAll(str, "_", "")

	// Chỉ lấy các ký tự A-Za-z0-9
	var result strings.Builder
	for _, char := range str {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result.WriteRune(char)
			if result.Len() == length {
				break
			}
		}
	}

	return result.String(), nil
}

// GenerateUniqueID tạo ID duy nhất theo định dạng yêu cầu
func GenerateUniqueID(contentType string) (string, error) {
	// Chuẩn hóa prefix để đảm bảo độ dài ID tối đa 12 ký tự
	prefix := normalizePrefix(contentType)

	// Tính độ dài phần ngẫu nhiên để tổng độ dài = 12
	randomLen := 12 - len(prefix)
	if randomLen < 1 {
		// Fallback an toàn, đảm bảo luôn có ít nhất 1 ký tự ngẫu nhiên
		randomLen = 1
		// Đồng thời cắt prefix về 11 ký tự nếu lỡ dài (không kỳ vọng xảy ra)
		if len(prefix) > 11 {
			prefix = prefix[:11]
		}
	}

	// Tạo phần ngẫu nhiên
	randomPart, err := generateRandomString(randomLen)
	if err != nil {
		return "", fmt.Errorf("lỗi khi tạo chuỗi ngẫu nhiên: %v", err)
	}

	// Kết hợp các phần để tạo ID
	return prefix + randomPart, nil
}

// normalizePrefix ánh xạ contentType về prefix 1 ký tự để phù hợp varchar(12)
func normalizePrefix(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch ct {
	case "", "d", "dialog", "dialogs":
		return "D"
	case "w", "word", "words":
		return "W"
	case "t", "topic", "topics":
		return "T"
	case "l", "level", "levels":
		return "L"
	case "a", "audio", "audios":
		return "A"
	case "i", "img", "image", "images":
		return "I"
	case "f", "fill", "fill_in_blank", "fill_in_blanks":
		return "F"
	case "s", "sub", "subscription", "subscriptions":
		return "S"
	case "u", "user_sub", "user_subscription", "user_subscriptions":
		return "U"
	case "auth", "auth_provider", "auth_providers":
		return "P"
	default:
		// Nếu người dùng truyền prefix 1 ký tự chữ cái, tôn trọng nó
		if len(ct) == 1 {
			ch := ct[0]
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				return strings.ToUpper(ct)
			}
		}
		return "D"
	}
}
