package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func ProcessTimestamp(scriptPath, mfaTimestampPath, outputPath string) error {
	// Đọc file kịch bản tạo dialog
	scriptData, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("không thể đọc file script: %v", err)
	}

	var script struct {
		Dialog []struct {
			Speaker string `json:"speaker"`
			Say     string `json:"say"`
		} `json:"dialog"`
	}
	if err := json.Unmarshal(scriptData, &script); err != nil {
		return fmt.Errorf("không thể parse file script: %v", err)
	}

	// Đọc file MFA timestamp
	mfaData, err := os.ReadFile(mfaTimestampPath)
	if err != nil {
		return fmt.Errorf("không thể đọc file MFA timestamp: %v", err)
	}

	var mfaTimestamp struct {
		Tiers struct {
			Words struct {
				Entries [][]interface{} `json:"entries"`
			} `json:"words"`
		} `json:"tiers"`
	}
	if err := json.Unmarshal(mfaData, &mfaTimestamp); err != nil {
		return fmt.Errorf("không thể parse file MFA timestamp: %v", err)
	}

	// Tạo dialog plain text và map từng từ với vị trí
	dialogText := ""

	for _, dialog := range script.Dialog {
		// Thêm câu vào dialogPlainText
		dialogText += dialog.Say + "\n"
	}
	dialogText = strings.TrimSpace(dialogText)

	// Normalize tất cả các loại dấu nháy đơn thành ASCII apostrophe
	// Sử dụng byte replacement để đảm bảo hoạt động với UTF-8
	dialogTextBytes := []byte(dialogText)
	// Replace UTF-8 encoded U+2019 (e2 80 99) với ASCII apostrophe (27)
	dialogTextBytes = bytes.ReplaceAll(dialogTextBytes, []byte{0xe2, 0x80, 0x99}, []byte{0x27})
	// Replace UTF-8 encoded U+2018 (e2 80 98) với ASCII apostrophe (27)
	dialogTextBytes = bytes.ReplaceAll(dialogTextBytes, []byte{0xe2, 0x80, 0x98}, []byte{0x27})
	// Replace all dashes with spaces
	dialogTextBytes = bytes.ReplaceAll(dialogTextBytes, []byte{0x2d}, []byte{0x20})
	// Replace em-dashes (—) with spaces - UTF-8 encoded U+2014 (e2 80 94)
	dialogTextBytes = bytes.ReplaceAll(dialogTextBytes, []byte{0xe2, 0x80, 0x94}, []byte{0x20})
	dialogText = string(dialogTextBytes)
	dialogPlainText := strings.ToLower(dialogText)

	// Cũng tạo một bản sao của script với các câu đã được normalize để dùng cho output
	normalizedScript := make([]struct {
		Speaker string
		Say     string
	}, len(script.Dialog))

	for i, dialog := range script.Dialog {
		// Normalize câu nói của từng dialog
		sayBytes := []byte(dialog.Say)
		// Replace UTF-8 encoded U+2019 (e2 80 99) với ASCII apostrophe (27)
		sayBytes = bytes.ReplaceAll(sayBytes, []byte{0xe2, 0x80, 0x99}, []byte{0x27})
		// Replace UTF-8 encoded U+2018 (e2 80 98) với ASCII apostrophe (27)
		sayBytes = bytes.ReplaceAll(sayBytes, []byte{0xe2, 0x80, 0x98}, []byte{0x27})
		// Replace all dashes with spaces
		sayBytes = bytes.ReplaceAll(sayBytes, []byte{0x2d}, []byte{0x20})
		// Replace em-dashes (—) with spaces - UTF-8 encoded U+2014 (e2 80 94)
		sayBytes = bytes.ReplaceAll(sayBytes, []byte{0xe2, 0x80, 0x94}, []byte{0x20})

		normalizedScript[i] = struct {
			Speaker string
			Say     string
		}{
			Speaker: dialog.Speaker,
			Say:     string(sayBytes),
		}
	}

	// Xử lý timestamp và tạo kết quả
	result := struct {
		Audio     string `json:"audio"`
		Sentences []struct {
			R  string `json:"r"`
			S  string `json:"s"`
			B  int    `json:"b"`
			T0 int    `json:"t0"`
		} `json:"sentence"`
		Words [][]interface{} `json:"words"`
	}{}

	// Set audio field to the corresponding audio filename
	scriptBaseName := strings.TrimSuffix(filepath.Base(scriptPath), filepath.Ext(scriptPath))
	result.Audio = scriptBaseName + ".ogg"

	searchPos := 0
	// Duyệt qua từng từ trong MFA timestamp
	for _, entry := range mfaTimestamp.Tiers.Words.Entries {
		startTime := int(entry[0].(float64) * 1000) // Convert to milliseconds
		endTime := int(entry[1].(float64) * 1000)   // Convert to milliseconds
		word := entry[2].(string)

		// Skip các từ <unk> hoặc rỗng
		if word == "<unk>" || word == "" {
			continue
		}

		// Chuyển từ MFA sang lowercase để so sánh
		wordLower := strings.ToLower(word)

		// Tìm vị trí của từ trong dialogPlainText bắt đầu từ searchPos
		pos := strings.Index(dialogPlainText[searchPos:], wordLower)
		if pos != -1 {
			// Tính vị trí theo số ký tự Unicode
			unicodePos := utf8.RuneCountInString(dialogPlainText[:searchPos+pos])

			// Lấy từ gốc từ dialogText với độ dài chính xác của từ trong MFA
			// Sử dụng rune để đảm bảo xử lý Unicode chính xác
			dialogRunes := []rune(dialogText)
			startRune := utf8.RuneCountInString(dialogText[:searchPos+pos])
			endRune := startRune + utf8.RuneCountInString(wordLower)

			if endRune <= len(dialogRunes) {
				originalWord := string(dialogRunes[startRune:endRune])

				// Thêm từ và timestamp vào kết quả (chỉ 3 trường: start_time, word, position)
				result.Words = append(result.Words, []interface{}{
					startTime,
					endTime,
					originalWord,
					unicodePos,
				})
				// Cập nhật searchPos để tìm từ tiếp theo
				searchPos += pos + len(wordLower)
			}
		}
	}

	// Xử lý sentences
	for i, dialog := range normalizedScript {
		// Sử dụng trực tiếp speaker name vì đã là tên nhân vật
		roleName := dialog.Speaker

		// Tính vị trí bắt đầu của câu trong dialog plain text
		startPos := 0
		for j := 0; j < i; j++ {
			startPos += utf8.RuneCountInString(normalizedScript[j].Say) + 1
		}

		// Tìm timestamp cho câu (chỉ cần t0 - thời điểm bắt đầu)
		var t0 int
		for _, word := range result.Words {
			pos := word[3].(int)
			if pos >= startPos && pos < startPos+utf8.RuneCountInString(dialog.Say) {
				if t0 == 0 || word[0].(int) < t0 {
					t0 = word[0].(int)
				}
			}
		}

		result.Sentences = append(result.Sentences, struct {
			R  string `json:"r"`
			S  string `json:"s"`
			B  int    `json:"b"`
			T0 int    `json:"t0"`
		}{
			R:  roleName,
			S:  dialog.Say,
			B:  startPos,
			T0: t0,
		})
	}

	// Ghi kết quả ra file
	outputData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("không thể tạo JSON output: %v", err)
	}

	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("không thể ghi file output: %v", err)
	}

	return nil
}

/*
=== XỬ LÝ CÁC TỪ VIẾT TẮT (CONTRACTIONS) VÀ DẤU GẠCH NGANG TRONG HÀM ProcessTimestamp ===

Vấn đề:
- File script JSON sử dụng Unicode right single quotation mark (') - U+2019 (UTF-8: e2 80 99)
- File MFA timestamp sử dụng ASCII apostrophe (') - ASCII 27
- Dấu gạch ngang (-) trong script có thể khác với format trong MFA timestamp
- Điều này khiến việc matching các từ viết tắt như "don't", "we've", "I'm" và từ có dấu gạch ngang bị thất bại

Giải pháp đã áp dụng:

1. NORMALIZE DẤU NHÁY VÀ DẤU GẠCH NGANG (dòng 64-74):
   - Sử dụng byte-level replacement để chuyển đổi Unicode quotation marks thành ASCII apostrophe
   - Replace UTF-8 encoded U+2019 (e2 80 99) → ASCII apostrophe (27)
   - Replace UTF-8 encoded U+2018 (e2 80 98) → ASCII apostrophe (27)
   - Replace dấu gạch ngang (-) ASCII 2d → dấu cách (space) ASCII 20
   - Replace em-dash (—) UTF-8 encoded U+2014 (e2 80 94) → dấu cách (space) ASCII 20

2. CẢI THIỆN LOGIC MATCHING (dòng 98-131):
   - Skip các từ "<unk>" hoặc rỗng từ MFA
   - Chuyển từ MFA sang lowercase để so sánh với dialogPlainText
   - Sử dụng rune để xử lý Unicode chính xác khi lấy từ gốc
   - Đảm bảo lấy từ gốc với case đúng từ dialogText (không phải lowercase)

3. XỬ LÝ UNICODE (dòng 114-117):
   - Sử dụng utf8.RuneCountInString() để tính vị trí Unicode chính xác
   - Chuyển dialogText thành []rune để slice theo ký tự Unicode
   - Đảm bảo độ dài từ được tính theo Unicode characters, không phải bytes

Kết quả:
- Tất cả các từ viết tắt như "we've", "don't", "I'm", "let's", "it's" đều được xử lý thành công
- Timestamp được gán chính xác cho từng từ viết tắt
- Giữ nguyên case gốc của từ trong output (VD: "I'm" không phải "i'm")

Các từ viết tắt đã test thành công:
- we've, don't, I'm, I've, Spain's, It's, Vietnam's, Let's

Lưu ý: Nếu gặp vấn đề tương tự với các loại dấu nháy khác, có thể thêm các byte replacement tương ứng.
*/
