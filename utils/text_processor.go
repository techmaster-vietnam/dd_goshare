package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Text2ScriptTask struct{}

type TextDialog struct {
	Speaker string `json:"speaker"`
	Say     string `json:"say"`
}

type DialogData struct {
	Dialog []TextDialog `json:"dialog"`
}

func ProcessTextToScript(inputFile, outPathText, outPathJson, outPathScript string) error {
	// Đọc file input
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("không thể đọc file input: %v", err)
	}

	// Xử lý nội dung để tạo text rút gọn và dialog data
	processedContent, dialogData := processContent(string(content))

	// Tạo tên file output dựa trên tên file input
	baseName := filepath.Base(inputFile)
	outputTextFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".txt"
	outputJsonFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".json"

	outputTextPath := filepath.Join(outPathText, outputTextFileName)
	outputScriptPath := filepath.Join(outPathScript, outputTextFileName)
	outputJsonPath := filepath.Join(outPathJson, outputJsonFileName)

	// Tạo thư mục output nếu chưa tồn tại
	if err := os.MkdirAll(outPathText, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục output: %v", err)
	}
	if err := os.MkdirAll(outPathJson, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục output: %v", err)
	}

	// Ghi file text output
	if err := os.WriteFile(outputTextPath, []byte(processedContent), 0644); err != nil {
		return fmt.Errorf("không thể ghi file text output: %v", err)
	}

	// Ghi file script output
	if err := os.WriteFile(outputScriptPath, []byte(processedContent), 0644); err != nil {
		return fmt.Errorf("không thể ghi file script output: %v", err)
	}

	// Ghi file JSON output
	jsonData := DialogData{Dialog: dialogData}
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("không thể tạo JSON: %v", err)
	}

	if err := os.WriteFile(outputJsonPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("không thể ghi file JSON output: %v", err)
	}

	fmt.Printf("Đã xử lý thành công: %s -> %s và %s\n", inputFile, outputTextPath, outputJsonPath)
	return nil
}

func processContent(content string) (string, []TextDialog) {
	lines := strings.Split(content, "\n")
	var processedLines []string
	var dialogData []TextDialog
	var foundSeparator bool
	var hasSeparator bool

	// Kiểm tra xem file có separator "---" không
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			hasSeparator = true
			break
		}
	}

	// Kiểm tra xem file có format dialog (có dấu ":") không
	var countFormat int
	for _, line := range lines {
		if strings.Contains(line, ":") && len(strings.SplitN(line, ":", 2)) > 1 {
			// Kiểm tra xem có phải là dialog format thật không (không phải chỉ là dấu ":" trong câu)
			parts := strings.SplitN(line, ":", 2)
			speaker := strings.TrimSpace(parts[0])
			content := strings.TrimSpace(parts[1])

			if len(speaker) > 0 && len(speaker) < 50 && len(content) > 0 {
				countFormat++
			}

		}
	}

	// Nếu không có dialog format, xử lý như story với speaker mặc định là "guest"
	if countFormat < 7 {
		// Xử lý separator nếu có
		var contentAfterSeparator []string
		if hasSeparator {
			for _, line := range lines {
				if strings.TrimSpace(line) == "---" {
					foundSeparator = true
					continue
				}
				if foundSeparator {
					contentAfterSeparator = append(contentAfterSeparator, line)
				}
			}
		} else {
			contentAfterSeparator = lines
		}

		// Loại bỏ các [markers] và tạo nội dung theo paragraph
		var paragraphs []string
		var currentParagraph []string

		for _, line := range contentAfterSeparator {
			// Loại bỏ tất cả các [markers] ở bất kỳ vị trí nào trong câu
			re := regexp.MustCompile(`\[[^\]]+\]`)
			cleanLine := re.ReplaceAllString(line, "")

			// Thay các dấu gạch ngang bằng khoảng trắng (bao gồm hyphen, en-dash, em-dash)
			replacer := strings.NewReplacer("-", " ", "–", " ", "—", " ", "—", " ")
			cleanLine = replacer.Replace(cleanLine)

			// Gộp nhiều khoảng trắng thành một khoảng trắng
			spaceRe := regexp.MustCompile(`\s+`)
			cleanLine = spaceRe.ReplaceAllString(cleanLine, " ")

			cleanLine = strings.TrimSpace(cleanLine)

			// Nếu dòng trống, kết thúc paragraph hiện tại
			if cleanLine == "" {
				if len(currentParagraph) > 0 {
					paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
					currentParagraph = []string{}
				}
			} else {
				currentParagraph = append(currentParagraph, cleanLine)
			}
		}

		// Thêm paragraph cuối cùng nếu có
		if len(currentParagraph) > 0 {
			paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
		}

		// Nếu không có paragraph nào (không có dòng trống để phân tách), coi toàn bộ là 1 paragraph
		if len(paragraphs) == 0 {
			var allLines []string
			for _, line := range contentAfterSeparator {
				re := regexp.MustCompile(`\[[^\]]+\]`)
				cleanLine := re.ReplaceAllString(line, "")
				replacer := strings.NewReplacer("-", " ", "–", " ", "—", " ", "—", " ")
				cleanLine = replacer.Replace(cleanLine)
				spaceRe := regexp.MustCompile(`\s+`)
				cleanLine = spaceRe.ReplaceAllString(cleanLine, " ")
				cleanLine = strings.TrimSpace(cleanLine)
				if cleanLine != "" {
					allLines = append(allLines, cleanLine)
				}
			}
			if len(allLines) > 0 {
				paragraphs = append(paragraphs, strings.Join(allLines, " "))
			}
		}

		// Tạo dialog data cho mỗi paragraph với speaker là "guest"
		var processedLines []string
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph != "" {
				dialogData = append(dialogData, TextDialog{
					Speaker: "guest",
					Say:     paragraph,
				})
				processedLines = append(processedLines, paragraph)
			}
		}

		return strings.Join(processedLines, "\n"), dialogData
	}

	// Xử lý format dialog như cũ
	for _, line := range lines {
		// Bước 1: Nếu có separator, loại bỏ tất cả các dòng text phía trên dòng "---" và cả dòng "---"
		if hasSeparator {
			if strings.TrimSpace(line) == "---" {
				foundSeparator = true
				continue
			}

			if !foundSeparator {
				continue
			}
		}

		// Bước 2: Xử lý dòng có tên người nói
		if strings.Contains(line, ":") {
			// Tách phần tên người nói và nội dung
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				speaker := strings.TrimSpace(parts[0])
				content := parts[1]

				// Loại bỏ tất cả các [markers] ở bất kỳ vị trí nào trong câu
				re := regexp.MustCompile(`\[[^\]]+\]`)
				cleanContent := re.ReplaceAllString(content, "")

				// Trim space và chỉ giữ lại nội dung thực sự được phát âm
				cleanContent = strings.TrimSpace(cleanContent)

				// Chỉ thêm vào dialog nếu có nội dung thực sự
				if cleanContent != "" {
					dialogData = append(dialogData, TextDialog{
						Speaker: speaker,
						Say:     cleanContent,
					})
				}

				// Cập nhật line cho text output
				line = cleanContent
			}
		}

		// Bước 3: Loại bỏ tất cả ký tự space trước và sau mỗi dòng
		line = strings.TrimSpace(line)

		// Giữ lại tất cả các dòng, kể cả dòng trống để duy trì cấu trúc
		processedLines = append(processedLines, line)
	}

	return strings.Join(processedLines, "\n"), dialogData
}
